package wrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Request struct {
	Request *http.Request

	bodyCache []byte
}

func (r *Request) GetForm() (map[string][]string, error) {
	data, err := r.GetBody()
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	r.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	if err := r.Request.ParseForm(); err != nil {
		return nil, fmt.Errorf("parsing form: %w", err)
	}
	return r.Request.Form, nil
}

func (r *Request) GetJSON() (interface{}, error) {
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

func (r *Request) GetBody() ([]byte, error) {
	if r.bodyCache == nil {
		var err error
		r.bodyCache, err = ioutil.ReadAll(r.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("reading body: %w", err)
		}
	}
	return r.bodyCache, nil
}

func (r *Request) GetIngress() bool {
	return true
}

func (r *Request) GetCookies() []*http.Cookie {
	return r.Request.Cookies()
}

func (r *Request) GetHeaders() map[string][]string {
	return r.Request.Header
}

func (r *Request) GetURL() *url.URL {
	return r.Request.URL
}
