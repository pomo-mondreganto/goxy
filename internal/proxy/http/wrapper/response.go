package wrapper

import (
	"io"
	"net/http"
	"net/url"
)

type Response struct {
	Response *http.Response
}

func (r *Response) GetIngress() bool {
	return false
}

func (r *Response) GetBody() io.ReadCloser {
	return r.Response.Body
}

func (r *Response) GetCookies() []*http.Cookie {
	return r.Response.Cookies()
}

func (r *Response) GetURL() *url.URL {
	if r.Response.Request != nil {
		return r.Response.Request.URL
	}
	return nil
}
