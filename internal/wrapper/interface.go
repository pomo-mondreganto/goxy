package wrapper

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	ErrNotSupported = errors.New("not supported")
)

type Entity interface {
	GetIngress() bool
	GetCookies() ([]*http.Cookie, error)
	GetHeaders() (map[string][]string, error)
	GetURL() (*url.URL, error)

	GetBody() ([]byte, error)
	GetJSON() (interface{}, error)
	GetForm() (map[string][]string, error)

	GetRaw() ([]byte, error)

	SetBody([]byte)
}
