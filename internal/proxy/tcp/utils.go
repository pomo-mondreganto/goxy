package tcp

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func isConnectionClosedErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrDropped) {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}

func genConnID(c net.Conn, num int) string {
	return fmt.Sprintf("%s:%d", c.RemoteAddr(), num)
}
