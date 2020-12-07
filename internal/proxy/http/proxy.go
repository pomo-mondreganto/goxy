package http

import (
	"context"
	"errors"
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/filters"
	"goxy/internal/proxy/http/wrapper"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	dropFlag   = "drop"
	acceptFlag = "accept"
)

var (
	ErrAlreadyRunning  = errors.New("proxy is already running")
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
)

type Proxy struct {
	ListenAddr string
	TargetAddr string

	closing bool
	server  *http.Server
	client  *http.Client
	wg      *sync.WaitGroup
	logger  *logrus.Entry
	filters []*filters.Filter
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
			if pctx.GetFlag(dropFlag) || pctx.GetFlag(acceptFlag) {
				break
			}
		}
	}
	return nil
}

func (p *Proxy) GetHandler() http.HandlerFunc {
	handleError := func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("Internal problem")); err != nil {
			p.logger.Errorf("Error writing response: %v", err)
		}
	}
	handleDrop := func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusNoContent)
	}

	reqLogger := p.logger.WithField("side", "request")
	respLogger := p.logger.WithField("side", "response")

	return func(w http.ResponseWriter, r *http.Request) {
		reqLogger.Debugf("New request: %v", r)
		pctx := common.NewProxyContext()

		reqEntity := &wrapper.Request{Request: r}
		if err := p.runFilters(pctx, reqEntity); err != nil {
			reqLogger.Errorf("Error running filters: %v", err)
			handleError(w)
			return
		}

		if pctx.GetFlag(dropFlag) {
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

		respEntity := &wrapper.Response{Response: response}
		if err := p.runFilters(pctx, respEntity); err != nil {
			respLogger.Errorf("Error running filters: %v", err)
			handleError(w)
			return
		}

		if pctx.GetFlag(dropFlag) {
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

func (p *Proxy) Serve() {
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
		Handler:      p.GetHandler(),
	}

	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.logger.Errorf("Error in server: %v", err)
	}
	p.logger.Infof("Server shutdown complete")
}

func (p *Proxy) Start() error {
	if p.wg != nil {
		return ErrAlreadyRunning
	}

	p.wg = &sync.WaitGroup{}
	p.wg.Add(1)
	go p.Serve()
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

func NewProxy(cfg *common.ServiceConfig, rs *filters.RuleSet) (*Proxy, error) {
	fts := make([]*filters.Filter, 0, len(cfg.Filters))
	for _, f := range cfg.Filters {
		rule, ok := rs.GetRule(f.Rule)
		if !ok {
			return nil, fmt.Errorf("invalid rule name: %s", f.Rule)
		}
		verdict, err := common.ParseVerdict(f.Verdict)
		if err != nil {
			return nil, fmt.Errorf("parse verdict: %w", err)
		}
		filter := &filters.Filter{
			Rule:    rule,
			Verdict: verdict,
		}
		fts = append(fts, filter)
	}

	logger := logrus.WithField("type", "http").WithField("listen", cfg.Listen)
	p := &Proxy{
		ListenAddr: cfg.Listen,
		TargetAddr: cfg.Target,

		logger:  logger,
		filters: fts,
	}
	return p, nil
}
