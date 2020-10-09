package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var logger = logrus.StandardLogger()

const (
	BUF_SIZE = 65536
)

type TcpProxy struct {
	Port       int
	RemoteAddr string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
	listener   net.Listener
}

func (t *TcpProxy) Stop() {
	t.cancel()
	t.wg.Wait()
	t.wg = nil
}

func (t *TcpProxy) HandleConn(conn net.Conn) {
	logger.Infof("Got connection")
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", t.RemoteAddr, t.Port))
	if err != nil {
		logger.Errorf("Failed to connect to remote server: %v", err)
		return
	}
	for {
		data := make([]byte, BUF_SIZE)
		n, err := conn.Read(data)
		if err != nil {
			logger.Infof("Failed to read from connection: %v", err)
			return
		}

		logger.Infof("Got %d bytes", n)

		_, err = remoteConn.Write(data[:n])
		if err != nil {
			logger.Errorf("Failed to remote server: %v", err)
			return
		}
	}
}

func (t *TcpProxy) Serve() {
	defer t.wg.Done()

	var connections []net.Conn
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			logger.Warnf("proxy stopped: %v", err)
			return
		}
		connections = append(connections, conn)
		go t.HandleConn(connections[len(connections)-1])
	}
}

func (t *TcpProxy) Start() error {
	if t.wg != nil {
		// already running
		return fmt.Errorf("proxy already running")
	}

	var err error
	t.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return err
	}

	t.wg = &sync.WaitGroup{}
	t.wg.Add(1)
	go t.Serve()
	return nil
}

func (t *TcpProxy) Wait() {
	if t.wg != nil {
		t.wg.Wait()
	}
}
