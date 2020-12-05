package wrapper

import (
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	Request *http.Request
}

func (r *Request) GetIngress() bool {
	return true
}

func (r *Request) GetBody() io.ReadCloser {
	return r.Request.Body
}

func (r *Request) GetCookies() []*http.Cookie {
	return r.Request.Cookies()
}

func (r *Request) GetURL() *url.URL {
	return r.Request.URL
}
