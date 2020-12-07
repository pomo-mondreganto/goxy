package tcp

import (
	"context"
	"errors"
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/tcp/filters"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

const BufSize = 64 * 1024

const (
	dropFlag   = "drop"
	acceptFlag = "accept"
)

var (
	ErrAlreadyRunning  = errors.New("proxy is already running")
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
	ErrDropped         = errors.New("connection dropped")
)

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

	logger := logrus.WithField("type", "tcp").WithField("listen", cfg.Listen)
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

	serviceConfig *common.ServiceConfig
	closing       bool
	wg            *sync.WaitGroup
	listener      net.Listener
	logger        *logrus.Entry
	filters       []*filters.Filter
}

func (p *Proxy) Start() error {
	if p.wg != nil {
		return ErrAlreadyRunning
	}

	var err error
	p.listener, err = net.Listen("tcp", p.ListenAddr)
	if err != nil {
		return fmt.Errorf("running listen: %w", err)
	}

	p.wg = &sync.WaitGroup{}
	p.wg.Add(1)
	go p.serve()
	return nil
}

func (p *Proxy) Shutdown(ctx context.Context) error {
	p.closing = true
	if err := p.listener.Close(); err != nil {
		return fmt.Errorf("closing listener: %w", err)
	}

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
	return p.serviceConfig
}

func (p *Proxy) runFilters(pctx *common.ProxyContext, buf []byte, ingress bool) error {
	for _, f := range p.filters {
		res, err := f.Rule.Apply(pctx, buf, ingress)
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

func (p *Proxy) oneSideHandler(conn *Connection, ingress bool) error {
	var src io.Reader
	var dst io.Writer

	if ingress {
		src = conn.Remote
		dst = conn.Local
	} else {
		src = conn.Local
		dst = conn.Remote
	}

	logger := p.logger.WithField("ingress", ingress)

	buf := make([]byte, BufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {

			data := buf[:nr]

			if err := p.runFilters(conn.Context, data, ingress); err != nil {
				logger.Errorf("Error running filters: %v", err)
			}

			if conn.Context.GetFlag(dropFlag) {
				logger.Debugf("Dropping connection")
				return ErrDropped
			}

			nw, ew := dst.Write(data)
			if ew != nil {
				return fmt.Errorf("proxy connection write: %w", ew)
			}
			if nr != nw {
				return fmt.Errorf("proxt connection write: %w", io.ErrShortWrite)
			}
		}
		if er != nil {
			if er != io.EOF {
				return fmt.Errorf("proxy connection read: %w", er)
			}
			break
		}
	}

	logger.Debug("Done handling connection")
	return nil
}

func (p *Proxy) handleConnection(conn net.Conn) {
	defer p.wg.Done()
	defer func() {
		if err := conn.Close(); err != nil && !isConnectionClosedErr(err) {
			p.logger.Warningf("Error closing connection from %v: %v", conn.RemoteAddr(), err)
		}
	}()

	p.logger.Debugf("Got new connection from: %v", conn.RemoteAddr())
	localConn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		p.logger.Errorf("Failed to connect to target server: %v", err)
		return
	}

	c := newConnection(conn, localConn)

	handler := func(wg *sync.WaitGroup, ingress bool) {
		logger := c.Logger.WithField("ingress", ingress)
		defer wg.Done()
		defer func() {
			logger.Debug("Closing bidi connection")
			if err := c.Close(); err != nil {
				logger.Fatal("Error closing bidi connection: ", err)
			}
			logger.Debug("Connection closed")
		}()
		if err := p.oneSideHandler(c, ingress); err != nil {
			if !isConnectionClosedErr(err) {
				logger.Errorf("Error in connection: %v", err)
			} else {
				logger.Debugf("Closed connection: %v", err)
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go handler(&wg, true)
	go handler(&wg, false)
	wg.Wait()
}

func (p *Proxy) serve() {
	defer p.wg.Done()
	conns := make([]net.Conn, 0)
	defer func() {
		p.logger.Debugf("Closing %d connections", len(conns))
		for _, c := range conns {
			logger := p.logger.WithField("src", c.RemoteAddr())
			if err := c.Close(); err != nil {
				if isConnectionClosedErr(err) {
					logger.Debugf("Connection already closed: %v", err)
				} else {
					logger.Errorf("Error closing connection: %v", err)
				}
			}
		}
	}()

	p.logger.Infof("Starting")

	for {
		p.logger.Debugf("Listening for connections")
		conn, err := p.listener.Accept()
		if err != nil {
			if p.closing {
				p.logger.Info("Listener exiting")
			} else {
				p.logger.Errorf("proxy stopped: %T: %v", err, err)
			}
			return
		}
		conns = append(conns, conn)
		p.wg.Add(1)
		go p.handleConnection(conn)
	}
}
