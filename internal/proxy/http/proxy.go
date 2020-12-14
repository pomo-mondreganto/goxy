package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"goxy/internal/common"
	"goxy/internal/proxy/http/filters"
	"goxy/internal/proxy/http/wrapper"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
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
	wg            sync.WaitGroup
	logger        *logrus.Entry
	filters       []filters.Filter
}

func (p *Proxy) GetListening() bool {
	return p.listening.Load()
}

func (p *Proxy) SetListening(state bool) {
	p.listening.Store(state)
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
	return fmt.Sprintf("HTTP proxy %s", p.ListenAddr)
}

func (p *Proxy) GetFilterDescriptions() []string {
	result := make([]string, 0, len(p.filters))
	for _, f := range p.filters {
		result = append(result, f.String())
	}
	return result
}

func (p *Proxy) runFilters(pctx *common.ProxyContext, e wrapper.Entity) error {
	for _, f := range p.filters {
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

func (p *Proxy) getHandler() http.HandlerFunc {
	handleError := func(w http.ResponseWriter) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
	handleDrop := func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusNoContent)
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

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			reqLogger.Errorf("Error reading body: %v", err)
			handleError(w)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
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
		r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		response, err := p.client.Do(r)
		if err != nil {
			respLogger.Errorf("Error making target request: %v", err)
			handleError(w)
			return
		}

		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			respLogger.Errorf("Error reading body: %v", err)
			handleError(w)
			return
		}
		if err := response.Body.Close(); err != nil {
			respLogger.Errorf("Error closing body: %v", err)
			handleError(w)
			return
		}

		response.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
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

		response.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
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
