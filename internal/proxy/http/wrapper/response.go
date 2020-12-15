package wrapper

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Response is a wrapper around http.Response implementing Entity interface.
// It's expected that Response.Body is already wrapped with BodyReader.
type Response struct {
	Response *http.Response
}

func (r *Response) GetForm() (map[string][]string, error) {
	// Response cannot contain a form.
	return nil, nil
}

func (r *Response) GetJSON() (interface{}, error) {
	defer r.resetBody()
	dec := json.NewDecoder(r.Response.Body)
	result := new(interface{})
	if err := dec.Decode(result); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return *result, nil
}

func (r *Response) GetIngress() bool {
	return false
}

func (r *Response) GetBody() ([]byte, error) {
	defer r.resetBody()
	buf, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return buf, nil
}

func (r *Response) GetCookies() []*http.Cookie {
	return r.Response.Cookies()
}

func (r *Response) GetHeaders() map[string][]string {
	return r.Response.Header
}

func (r *Response) GetURL() *url.URL {
	if r.Response.Request != nil {
		return r.Response.Request.URL
	}
	return nil
}

func (r *Response) resetBody() {
	if err := r.Response.Body.Close(); err != nil {
		logrus.Errorf("Error resetting response body: %v", err)
	}
}
