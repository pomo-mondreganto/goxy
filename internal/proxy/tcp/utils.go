package tcp

import "strings"

func isConnectionClosedErr(err error) bool {
	if err == nil {
		return false
	}
	if err == ErrDropped {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection")
}
