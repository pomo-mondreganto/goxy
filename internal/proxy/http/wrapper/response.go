package wrapper

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type Response struct {
	Response *http.Response
}

func (r *Response) GetJSON() interface{} {
	dec := json.NewDecoder(r.GetBody())
	result := new(interface{})
	if err := dec.Decode(result); err != nil {
		logrus.Warningf("Error decoding JSON in response: %v", err)
		return nil
	}
	return result
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
