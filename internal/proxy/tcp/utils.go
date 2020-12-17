package tcp

import (
	"errors"
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
