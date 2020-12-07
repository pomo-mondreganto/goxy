package wrapper

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	Request *http.Request
}

func (r *Request) GetJSON() interface{} {
	dec := json.NewDecoder(r.GetBody())
	result := new(interface{})
	if err := dec.Decode(result); err != nil {
		logrus.Warningf("Error decoding JSON in request: %v", err)
		return nil
	}
	return *result
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
