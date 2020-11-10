package tcp

import (
	"context"
	"errors"
	"fmt"
	"goxy/internal/common"
	tcpcommon "goxy/internal/common/tcp"
	"goxy/internal/filters/tcp"
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

const BufSize = 64 * 1024

var (
	ErrAlreadyRunning  = errors.New("proxy is already running")
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
	ErrDropped         = errors.New("connection dropped")
)

type Proxy struct {
	ListenAddr string
	TargetAddr string

	closing  bool
	wg       *sync.WaitGroup
	listener net.Listener
	logger   *logrus.Entry
	filters  []*tcp.Filter
}

func (p *Proxy) runFilters(conn *Connection, buf []byte, ingress bool) error {
	for _, f := range p.filters {
		res, err := f.Rule.Apply(buf, ingress)
		if err != nil {
			return fmt.Errorf("error in rule %T: %w", f.Rule, err)
		}
		if res {
			conn.mu.Lock()
			if err := f.Verdict.Mutate(conn.Context); err != nil {
				conn.mu.Unlock()
				return fmt.Errorf("error mutating verdict %T: %w", f.Verdict, err)
			}
			if conn.Context.MustDrop || conn.Context.MustAccept {
				conn.mu.Unlock()
				break
			}
			conn.mu.Unlock()
		}
	}
	return nil
}

func (p *Proxy) handleConnection(conn *Connection, ingress bool) error {
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

			if err := p.runFilters(conn, data, ingress); err != nil {
				logger.Errorf("Error running filters: %v", err)
			}

			conn.mu.Lock()
			if conn.Context.MustDrop {
				logger.Debugf("Dropping connection")
				conn.mu.Unlock()
				return ErrDropped
			}
			conn.mu.Unlock()

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
	return nil
}

func (p *Proxy) HandleConn(conn net.Conn) {
	defer p.wg.Done()
	defer func() {
		if err := conn.Close(); err != nil && !isConnectionClosedErr(err) {
			p.logger.Warningf("Error closing connection: %v", err)
		}
	}()

	p.logger.Debugf("Got new connection from: %v", conn.RemoteAddr())
	localConn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		p.logger.Errorf("Failed to connect to target server: %v", err)
		return
	}

	c := NewConnection(conn, localConn)

	wg := sync.WaitGroup{}
	handler := func(ingress bool, other net.Conn) {
		logger := p.logger.WithField("ingress", ingress)
		defer func() {
			if err := c.Close(); err != nil {
				logger.Fatal("Error closing bidi connection: ", err)
			}
			wg.Done()
		}()
		if err := p.handleConnection(c, ingress); err != nil {
			if !isConnectionClosedErr(err) {
				logger.Errorf("Error in connection: %v", err)
			} else {
				logger.Debugf("Closed connection: %v", err)
			}
		}
	}

	wg.Add(2)
	go handler(true, c.Local)
	go handler(false, c.Remote)
	wg.Wait()
}

func (p *Proxy) Serve() {
	defer p.wg.Done()

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
		p.wg.Add(1)
		go p.HandleConn(conn)
	}
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
	go p.Serve()
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

func NewProxy(cfg *common.ServiceConfig, rs *tcp.RuleSet) (*Proxy, error) {
	filters := make([]*tcp.Filter, 0, len(cfg.Filters))
	for _, f := range cfg.Filters {
		rule, ok := rs.GetRule(f.Rule)
		if !ok {
			return nil, fmt.Errorf("invalid rule name: %s", f.Rule)
		}
		verdict, err := tcpcommon.ParseVerdict(f.Verdict)
		if err != nil {
			return nil, fmt.Errorf("parse verdict: %w", err)
		}
		filter := &tcp.Filter{
			Rule:    rule,
			Verdict: verdict,
		}
		filters = append(filters, filter)
	}

	logger := logrus.WithField("type", "tcp").WithField("listen", cfg.Listen)
	p := &Proxy{
		ListenAddr: cfg.Listen,
		TargetAddr: cfg.Target,

		logger:  logger,
		filters: filters,
	}
	return p, nil
}
