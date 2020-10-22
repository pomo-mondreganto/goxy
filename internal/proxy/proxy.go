package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	ErrAlreadyRunning  = errors.New("proxy is already running")
	ErrShutdownTimeout = errors.New("proxy shutdown timeout")
)

const (
	BufSize = 65536
)

type TcpProxy struct {
	Port       int
	RemoteAddr string
	closing    bool
	wg         *sync.WaitGroup
	listener   net.Listener
	logger     *logrus.Entry
}

func (t *TcpProxy) HandleConn(conn net.Conn) {
	t.logger.Debugf("Got new connection from: %v", conn.RemoteAddr())
	remoteConn, err := net.Dial("tcp", t.RemoteAddr)
	if err != nil {
		t.logger.Errorf("Failed to connect to remote server: %v", err)
		return
	}
	for {
		data := make([]byte, BufSize)
		n, err := conn.Read(data)
		if err != nil {
			t.logger.Infof("Failed to read from connection: %v", err)
			return
		}

		t.logger.Debugf("Got %d bytes", n)

		_, err = remoteConn.Write(data[:n])
		if err != nil {
			t.logger.Errorf("Failed to write to connection: %v", err)
			return
		}
	}
}

func (t *TcpProxy) Serve() {
	defer t.wg.Done()

	t.logger.Infof("Starting")

	var connections []net.Conn
	defer func() {
		for _, conn := range connections {
			if err := conn.Close(); err != nil {
				t.logger.Fatal("Error closing connection: ", err)
			}
		}
	}()

	for {
		t.logger.Debugf("Listening for connections")
		conn, err := t.listener.Accept()
		if err != nil {
			if t.closing {
				t.logger.Info("Listener exiting")
			} else {
				t.logger.Errorf("proxy stopped: %T: %v", err, err)
			}
			return
		}
		connections = append(connections, conn)
		go t.HandleConn(connections[len(connections)-1])
	}
}

func (t *TcpProxy) Start() error {
	if t.wg != nil {
		return ErrAlreadyRunning
	}

	var err error
	t.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return fmt.Errorf("running listen: %w", err)
	}

	t.wg = &sync.WaitGroup{}
	t.wg.Add(1)
	go t.Serve()
	return nil
}

func (t *TcpProxy) Shutdown(ctx context.Context) error {
	t.closing = true
	if err := t.listener.Close(); err != nil {
		return fmt.Errorf("closing listener: %w", err)
	}

	done := make(chan interface{}, 1)
	go func() {
		t.wg.Wait()
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

func NewTcpProxy(remoteHost string, remotePort, localPort int) *TcpProxy {
	logger := logrus.WithField("type", "tcp").WithField("port", localPort)
	p := &TcpProxy{
		logger:     logger,
		Port:       localPort,
		RemoteAddr: fmt.Sprintf("%s:%d", remoteHost, remotePort),
	}
	return p
}
