package tcp

import (
	"github.com/sirupsen/logrus"
	"goxy/internal/common"
	"net"
)

type Connection struct {
	Remote  net.Conn
	Local   net.Conn
	Context *common.ProxyContext
	Logger  *logrus.Entry
}

func (c *Connection) CloseCounterpart(ingress bool) error {
	if ingress {
		if err := c.Local.Close(); err != nil && !isConnectionClosedErr(err) {
			return err
		}
	} else {
		if err := c.Remote.Close(); err != nil && !isConnectionClosedErr(err) {
			return err
		}
	}
	return nil
}

func newConnection(remote net.Conn, local net.Conn) *Connection {
	return &Connection{
		Remote:  remote,
		Local:   local,
		Context: common.NewProxyContext(),
		Logger:  logrus.WithField("src", remote.RemoteAddr()),
	}
}
