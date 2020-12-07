package wrapper

import (
	"net/http"
	"net/url"
)

type Entity interface {
	GetIngress() bool
	GetCookies() []*http.Cookie
	GetHeaders() map[string][]string
	GetURL() *url.URL

	GetBody() ([]byte, error)
	GetJSON() (interface{}, error)
	GetForm() (map[string][]string, error)
}
