package tcp

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"goxy/internal/common"
	"goxy/internal/proxy/tcp/filters"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

const BufSize = 64 * 1024

var (
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
	ErrDropped         = errors.New("connection dropped")
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

	logger := logrus.WithField("type", "tcp").WithField("listen", cfg.Listen)
	p := &Proxy{
		ListenAddr: cfg.Listen,
		TargetAddr: cfg.Target,

		serviceConfig: cfg,
		logger:        logger,
		filters:       fts,
		conns:         make(map[string]net.Conn),
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
	conns         map[string]net.Conn
	connSeq       atomic.Int32
	wg            *sync.WaitGroup
	listener      net.Listener
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
	p.SetListening(true)

	var err error
	p.listener, err = net.Listen("tcp", p.ListenAddr)
	if err != nil {
		return fmt.Errorf("running listen: %w", err)
	}

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
		p.logger.Debugf("Closing %d connections", len(p.conns))
		for id, c := range p.conns {
			logger := p.logger.WithField("conn", id)
			if err := c.Close(); err != nil {
				if isConnectionClosedErr(err) {
					logger.Debugf("Connection already closed: %v", err)
				} else {
					logger.Errorf("Error closing connection: %v", err)
				}
			}
		}

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

func (p Proxy) runFilters(pctx *common.ProxyContext, buf []byte, ingress bool) error {
	for _, f := range p.filters {
		if !f.IsEnabled() {
			continue
		}
		res, err := f.Rule.Apply(pctx, buf, ingress)
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

func (p Proxy) String() string {
	return fmt.Sprintf("TCP proxy %s", p.ListenAddr)
}

func (p Proxy) GetFilters() []common.Filter {
	result := make([]common.Filter, 0, len(p.filters))
	for _, f := range p.filters {
		f := f
		result = append(result, &f)
	}
	return result
}

func (p Proxy) oneSideHandler(conn *Connection, logger *logrus.Entry, ingress bool) error {
	var (
		src io.Reader
		dst io.Writer
	)
	if ingress {
		src = conn.Remote
		dst = conn.Local
	} else {
		src = conn.Local
		dst = conn.Remote
	}

	buf := make([]byte, BufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {

			data := buf[:nr]

			if err := p.runFilters(conn.Context, data, ingress); err != nil {
				logger.Errorf("Error running filters: %v", err)
			}

			if conn.Context.GetFlag(common.DropFlag) {
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

func (p Proxy) handleConnection(id string) {
	defer p.wg.Done()

	conn := p.conns[id]
	connLogger := p.logger.WithField("conn", id)
	defer func() {
		if err := conn.Close(); err != nil && !isConnectionClosedErr(err) {
			connLogger.Warningf("Error closing connection: %v", err)
		}
		delete(p.conns, id)
	}()

	connLogger.Debugf("Connection received")
	localConn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		connLogger.Errorf("Failed to connect to target: %v", err)
		return
	}

	c := newConnection(conn, localConn)

	handler := func(wg *sync.WaitGroup, ingress bool) {
		defer wg.Done()
		logger := connLogger.WithField("ingress", ingress)
		defer func() {
			logger.Debug("Closing counterpart connection")
			if err := c.CloseCounterpart(ingress); err != nil {
				logger.Fatalf("Error closing counterpart connection: %v", err)
			}
			logger.Debug("Counterpart connection closed")
		}()
		if err := p.oneSideHandler(c, logger, ingress); err != nil {
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

	p.logger.Infof("Starting")

	for {
		p.logger.Debugf("Listening for connections")
		conn, err := p.listener.Accept()
		if err != nil {
			if p.closing {
				p.logger.Info("Listener exiting")
			} else {
				p.logger.Errorf("Proxy stopped: %T: %v", err, err)
			}
			return
		}

		connID := genConnID(conn, int(p.connSeq.Inc()))
		logger := p.logger.WithField("conn", connID)

		if !p.GetListening() {
			logger.Debugf("Proxy closed, dropping the connection")
			if err := conn.Close(); err != nil {
				logger.Errorf("Error dropping the connection: %v", err)
			}
			continue
		}

		p.conns[connID] = conn
		p.wg.Add(1)
		go p.handleConnection(connID)
	}
}
