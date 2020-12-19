package tcp

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"net"
	"sync"
)

func newConnMap() *connMap {
	return &connMap{
		conns: make(map[string]net.Conn),
		mu:    new(sync.RWMutex),
		seq:   new(atomic.Int32),
	}
}

type connMap struct {
	conns map[string]net.Conn
	mu    *sync.RWMutex
	seq   *atomic.Int32
}

func (m *connMap) add(conn net.Conn) string {
	id := genConnID(conn, int(m.seq.Inc()))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[id] = conn
	return id
}

func (m *connMap) remove(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, connID)
}

func (m *connMap) length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.conns)
}

func (m *connMap) get(connID string) net.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conns[connID]
}

func (m *connMap) closeAll(logger *logrus.Entry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.Debugf("Closing %d connections", len(m.conns))
	for id, c := range m.conns {
		cl := logger.WithField("conn", id)
		if err := c.Close(); err != nil {
			if isConnectionClosedErr(err) {
				cl.Debugf("Connection already closed: %v", err)
			} else {
				cl.Errorf("Error closing connection: %v", err)
			}
		}
	}
}
