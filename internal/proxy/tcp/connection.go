package tcp

import (
	"github.com/sirupsen/logrus"
	"goxy/internal/common"
	"net"
)

type Connection struct {
	Remote  net.Conn
	Local   net.Conn
	Context *common.ConnectionContext
	Logger  *logrus.Entry
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

func NewConnection(remote net.Conn, local net.Conn) *Connection {
	return &Connection{
		Remote:  remote,
		Local:   local,
		Context: common.NewContext(),
		Logger:  logrus.WithField("src", remote.RemoteAddr()),
	}
}
