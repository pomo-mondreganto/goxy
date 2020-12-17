package http

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"goxy/internal/common"
	"goxy/internal/proxy/http/filters"
	"goxy/internal/proxy/http/wrapper"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
	ErrInvalidFilter   = errors.New("no such filter")
)

func NewProxy(cfg common.ServiceConfig, rs *filters.RuleSet) (*Proxy, error) {
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
		fts = append(fts, filter)
	}

	logger := logrus.WithField("type", "http").WithField("listen", cfg.Listen)
	p := &Proxy{
		ListenAddr: cfg.Listen,
		TargetAddr: cfg.Target,

		serviceConfig: cfg,
		logger:        logger,
		filters:       fts,
		wg:            new(sync.WaitGroup),
	}
	return p, nil
}

type Proxy struct {
	ListenAddr string
	TargetAddr string

	serviceConfig common.ServiceConfig
	closing       bool
	listening     atomic.Bool
	server        *http.Server
	client        *http.Client
	wg            *sync.WaitGroup
	logger        *logrus.Entry
	filters       []filters.Filter
}

func (p Proxy) GetListening() bool {
	return p.listening.Load()
}

func (p *Proxy) SetListening(state bool) {
	p.listening.Store(state)
}

func (p *Proxy) SetFilterEnabled(filter int, enabled bool) error {
	if filter < 0 || filter >= len(p.filters) {
		return ErrInvalidFilter
	}
	p.filters[filter].SetEnabled(enabled)
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

func (p Proxy) GetConfig() *common.ServiceConfig {
	return &p.serviceConfig
}

func (p Proxy) String() string {
	return fmt.Sprintf("HTTP proxy %s", p.ListenAddr)
}

func (p Proxy) GetFilters() []common.Filter {
	result := make([]common.Filter, 0, len(p.filters))
	for _, f := range p.filters {
		f := f
		result = append(result, &f)
	}
	return result
}

func (p Proxy) runFilters(pctx *common.ProxyContext, e wrapper.Entity) error {
	for _, f := range p.filters {
		if !f.IsEnabled() {
			continue
		}
		res, err := f.Rule.Apply(pctx, e)
		if err != nil {
			return fmt.Errorf("error in rule %T: %w", f.Rule, err)
		}
		if res {
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

func (p Proxy) getHandler() http.HandlerFunc {
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

	reqLogger := p.logger.WithField("side", "request")
	respLogger := p.logger.WithField("side", "response")

	return func(w http.ResponseWriter, r *http.Request) {
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

		pctx := common.NewProxyContext()
		reqEntity := &wrapper.Request{Request: r}
		if err := p.runFilters(pctx, reqEntity); err != nil {
			reqLogger.Errorf("Error running filters: %v", err)
			handleError(w)
			return
		}

		if pctx.GetFlag(common.DropFlag) {
			reqLogger.Debugf("Dropping connection")
			handleDrop(w)
			return
		}

		r.URL.Scheme = "http"
		r.URL.Host = p.TargetAddr
		r.RequestURI = ""
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

		if pctx.GetFlag(common.DropFlag) {
			respLogger.Debugf("Dropping connection")
			handleDrop(w)
			return
		}

		w.WriteHeader(response.StatusCode)
		for k, vals := range response.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		if _, err := io.Copy(w, response.Body); err != nil {
			respLogger.Errorf("Error copying body: %v", err)
			handleError(w)
			return
		}
	}
}

func (p *Proxy) serve() {
	defer p.wg.Done()

	p.logger.Info("Starting")

	p.client = &http.Client{
		Timeout: time.Second * 5,
	}

	p.server = &http.Server{
		Addr:         p.ListenAddr,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
		Handler:      p.getHandler(),
	}

	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.logger.Errorf("Error in server: %v", err)
	}
	p.logger.Infof("Server shutdown complete")
}
