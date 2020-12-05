package wrapper

import (
	"io"
	"net/http"
	"net/url"
)

type Entity interface {
	GetBody() io.ReadCloser
	GetCookies() []*http.Cookie
	GetURL() *url.URL
	GetIngress() bool
}
