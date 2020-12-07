package wrapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Response struct {
	Response *http.Response

	bodyCache []byte
}

func (r *Response) GetForm() (map[string][]string, error) {
	// Response cannot contain a form.
	return nil, nil
}

func (r *Response) GetJSON() (interface{}, error) {
	data, err := r.GetBody()
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	result := new(interface{})
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return *result, nil
}

func (r *Response) GetIngress() bool {
	return false
}

func (r *Response) GetBody() ([]byte, error) {
	if r.bodyCache == nil {
		var err error
		r.bodyCache, err = ioutil.ReadAll(r.Response.Body)
		if err != nil {
			return nil, fmt.Errorf("reading body: err")
		}
	}
	return r.bodyCache, nil
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
