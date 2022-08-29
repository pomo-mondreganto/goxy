package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go.uber.org/atomic"

	"goxy/internal/common"
	"goxy/internal/export"
	"goxy/internal/filters"
	"goxy/internal/wrapper"

	"github.com/sirupsen/logrus"
)

var (
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
	ErrInvalidFilter   = errors.New("no such filter")
)

func NewProxy(cfg common.ServiceConfig, rs *filters.RuleSet, pc *export.ProducerClient) (*Proxy, error) {
	fts := make([]filters.Filter, 0, len(cfg.Filters))
	for _, f := range cfg.Filters {
		rule, ok := rs.GetRule(f.Rule)
		if !ok {
			return nil, fmt.Errorf("invalid rule name: %s", f.Rule)
		}
		verdict, err := common.ParseVerdict(f.Verdict)
		if err != nil {
			return nil, fmt.Errorf("parse verdict: %w", err)
		}
		filter := filters.Filter{
			Rule:    rule,
			Verdict: verdict,
		}
		filter.SetAlert(f.Alert)
		fts = append(fts, filter)
	}

	target, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, fmt.Errorf("parsing target: %w", err)
	}

	timeout := time.Second * 5
	if cfg.RequestTimeout != 0 {
		timeout = cfg.RequestTimeout
	}

	logger := logrus.WithFields(logrus.Fields{
		"type":   "http",
		"listen": cfg.Listen,
	})
	_, listenPortStr, err := net.SplitHostPort(cfg.Listen)
	if err != nil {
		return nil, fmt.Errorf("parsing listen addr: %w", err)
	}
	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid listen port %s: %w", listenPortStr, err)
	}
	p := &Proxy{
		ListenAddr: cfg.Listen,
		Target:     target,
		Name:       cfg.Name,

		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		binder: &streamBinder{
			name:       cfg.Name,
			listenPort: listenPort,
			bindings:   make(map[string]streamBinding),
		},
		serviceConfig: cfg,
		logger:        logger,
		filters:       fts,
		producer:      pc,
		listening:     atomic.NewBool(false),
	}

	p.server = &http.Server{
		Addr:         p.ListenAddr,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
		Handler:      p.getHandler(),
	}

	if cfg.TLS != nil {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.Cert, cfg.TLS.Key)
		if err != nil {
			return nil, fmt.Errorf("loading tls config: %w", err)
		}
		p.TLSConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		p.client.Transport = &http.Transport{
			TLSClientConfig:   p.TLSConfig,
			ForceAttemptHTTP2: true,
		}
		p.server.TLSConfig = p.TLSConfig
	}

	return p, nil
}

type Proxy struct {
	ListenAddr string
	Target     *url.URL
	Name       string
	TLSConfig  *tls.Config

	binder        *streamBinder
	serviceConfig common.ServiceConfig
	closing       bool
	listening     *atomic.Bool
	server        *http.Server
	client        *http.Client
	wg            sync.WaitGroup
	logger        *logrus.Entry
	filters       []filters.Filter
	producer      *export.ProducerClient
}

func (p *Proxy) GetListening() bool {
	return p.listening.Load()
}

func (p *Proxy) SetListening(state bool) {
	p.listening.Store(state)
}

func (p *Proxy) SetFilterState(filter int, enabled, alert bool) error {
	if filter < 0 || filter >= len(p.filters) {
		return ErrInvalidFilter
	}
	p.filters[filter].SetEnabled(enabled)
	p.filters[filter].SetAlert(alert)
	return nil
}

func (p *Proxy) Start() error {
	p.wg.Add(1)
	p.SetListening(true)

	go p.serve()
	return nil
}

func (p *Proxy) Shutdown(ctx context.Context) error {
	p.closing = true
	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down server: %w", err)
	}
	p.client.CloseIdleConnections()

	done := make(chan interface{}, 1)
	go func() {
		p.wg.Wait()
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ErrShutdownTimeout
	case <-done:
		break
	}
	return nil
}

func (p *Proxy) GetConfig() *common.ServiceConfig {
	return &p.serviceConfig
}

func (p *Proxy) String() string {
	return fmt.Sprintf("HTTP proxy %s (%s)", p.Name, p.ListenAddr)
}

func (p *Proxy) GetFilters() []common.Filter {
	result := make([]common.Filter, 0, len(p.filters))
	for _, f := range p.filters {
		f := f
		result = append(result, &f)
	}
	return result
}

func (p *Proxy) runFilters(pctx *common.ProxyContext, e wrapper.Entity) error {
	for _, f := range p.filters {
		if !f.IsEnabled() {
			continue
		}
		res, err := f.Rule.Apply(pctx, e)
		if err != nil {
			return fmt.Errorf("error in rule %T: %w", f.Rule, err)
		}
		if res {
			if f.GetAlert() {
				p.logger.Warningf("Rule %v triggered", f.Rule)
			}
			if err := f.Verdict.Mutate(pctx); err != nil {
				return fmt.Errorf("error mutating verdict %T: %w", f.Verdict, err)
			}
			if pctx.GetFlag(common.DropFlag) || pctx.GetFlag(common.AcceptFlag) {
				break
			}
		}
	}
	return nil
}

func (p *Proxy) getHandler() http.HandlerFunc {
	handleError := func(w http.ResponseWriter) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
	handleDrop := func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusNoContent)
	}
	wrapBody := func(body io.ReadCloser) (io.ReadCloser, error) {
		w, err := wrapper.NewBodyReader(body)
		if err != nil {
			return nil, fmt.Errorf("creating reader: %w", err)
		}
		if err := body.Close(); err != nil {
			return nil, fmt.Errorf("closing original body: %w", err)
		}
		return w, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		reqLogger := p.logger.WithFields(logrus.Fields{
			"side": "request",
			"addr": r.RemoteAddr,
		})
		respLogger := p.logger.WithFields(logrus.Fields{
			"side": "response",
			"addr": r.RemoteAddr,
		})

		reqLogger.Debugf("New request: %v", r)

		if !p.GetListening() {
			reqLogger.Debugf("Proxy is not listening, dropping")
			handleDrop(w)
			return
		}

		var err error
		if r.Body, err = wrapBody(r.Body); err != nil {
			reqLogger.Errorf("Error wrapping body: %v", err)
			handleError(w)
			return
		}
		r.Header.Set("Host", r.Host)

		pctx := common.NewProxyContext()
		reqEntity := &wrapper.Request{Request: r}
		if err := p.runFilters(pctx, reqEntity); err != nil {
			reqLogger.Errorf("Error running filters: %v", err)
			handleError(w)
			return
		}

		base := p.binder.GetOrCreate(r)
		// Do not terminate export if request is terminated.
		if err := p.exportEntity(context.Background(), reqEntity, base, reqLogger); err != nil {
			reqLogger.Errorf("Error exporting packet: %v", err)
			handleError(w)
			return
		}

		if pctx.GetFlag(common.DropFlag) {
			reqLogger.Debugf("Dropping connection")
			handleDrop(w)
			return
		}

		r.URL.Scheme = p.Target.Scheme
		r.URL.Host = p.Target.Host
		r.RequestURI = ""
		r.Host = ""
		r.Header.Del("Accept-Encoding")
		reqLogger.Debugf("Making a request: %v", r)
		response, err := p.client.Do(r)
		if err != nil {
			respLogger.Errorf("Error making target request: %v", err)
			handleError(w)
			return
		}

		if response.Body, err = wrapBody(response.Body); err != nil {
			respLogger.Errorf("Error wrapping body: %v", err)
			handleError(w)
			return
		}

		respEntity := &wrapper.Response{Response: response}
		if err := p.runFilters(pctx, respEntity); err != nil {
			respLogger.Errorf("Error running filters: %v", err)
			handleError(w)
			return
		}
		// Do not terminate export if request is terminated.
		if err := p.exportEntity(context.Background(), respEntity, base, respLogger); err != nil {
			respLogger.Errorf("Error exporting packet: %v", err)
			handleError(w)
			return
		}

		if pctx.GetFlag(common.DropFlag) {
			respLogger.Debugf("Dropping connection")
			handleDrop(w)
			return
		}

		for k, vals := range response.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(response.StatusCode)

		if _, err := io.Copy(w, response.Body); err != nil {
			respLogger.Errorf("Error copying body: %v", err)
			handleError(w)
			return
		}
	}
}

func (p *Proxy) exportEntity(ctx context.Context, e wrapper.Entity, base *export.BasePacket, logger *logrus.Entry) error {
	if p.producer == nil {
		logger.Debug("Exporter is disabled")
		return nil
	}
	body, err := e.GetRaw()
	if err != nil {
		return fmt.Errorf("getting raw data: %w", err)
	}
	reqPacket := export.Packet{
		BasePacket:  base,
		Content:     body,
		CaptureTime: time.Now(),
		FilterData:  0, // TODO: add filters.
		Inbound:     e.GetIngress(),
		Reversed:    !e.GetIngress(),
	}
	if err := p.producer.Send(ctx, &reqPacket); err != nil {
		logger.Warningf("Error sending to producer: %v", err)
	}
	return nil
}

func (p *Proxy) serve() {
	defer p.wg.Done()

	p.logger.Info("Starting")

	if p.TLSConfig != nil {
		if err := p.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			p.logger.Errorf("Error in server: %v", err)
		}
	} else {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Errorf("Error in server: %v", err)
		}
	}
	p.logger.Infof("Server shutdown complete")
}

func splitAddrSafe(addr string, defPort int) (host string, port int) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = defPort
	} else {
		if port, err = strconv.Atoi(portStr); err != nil {
			port = defPort
		}
	}
	return host, port
}
