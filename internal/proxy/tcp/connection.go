package tcp

import (
	"goxy/internal/common"
	"net"
	"sync"
)

type Connection struct {
	Remote  net.Conn
	Local   net.Conn
	Context *common.ConnectionContext

	mu sync.Mutex
}

func (c *Connection) Close() error {
	if err := c.Local.Close(); err != nil && !isConnectionClosedErr(err) {
		return err
	}
	if err := c.Remote.Close(); err != nil && !isConnectionClosedErr(err) {
		return err
	}
	return nil
}

func NewConnection(r net.Conn, l net.Conn) *Connection {
	return &Connection{
		Remote:  r,
		Local:   l,
		Context: common.NewContext(),
	}
}
