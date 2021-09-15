package wrapper

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Request is a wrapper around http.Request implementing Entity interface.
// It's expected that Request.Body is already wrapped with BodyReader.
type Request struct {
	Request *http.Request
}

func (r Request) GetForm() (map[string][]string, error) {
	defer r.resetBody()
	if err := r.Request.ParseForm(); err != nil {
		return nil, fmt.Errorf("parsing form: %w", err)
	}
	return r.Request.Form, nil
}

func (r Request) GetJSON() (interface{}, error) {
	defer r.resetBody()
	dec := json.NewDecoder(r.Request.Body)
	result := new(interface{})
	if err := dec.Decode(result); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return *result, nil
}

func (r Request) GetBody() ([]byte, error) {
	defer r.resetBody()
	buf, err := ioutil.ReadAll(r.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return buf, nil
}

func (r Request) GetIngress() bool {
	return true
}

func (r Request) GetCookies() []*http.Cookie {
	return r.Request.Cookies()
}

func (r Request) GetHeaders() map[string][]string {
	return r.Request.Header
}

func (r Request) GetURL() *url.URL {
	return r.Request.URL
}

func (r Request) SetBody(data []byte) {
	r.resetBody()
	r.Request.Body = NewBodyReaderFromRaw(data)
}

func (r Request) resetBody() {
	if err := r.Request.Body.Close(); err != nil {
		logrus.Errorf("Error resetting request body: %v", err)
	}
}
