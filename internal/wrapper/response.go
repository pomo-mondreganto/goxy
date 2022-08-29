package wrapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

// Response is a wrapper around http.Response implementing Entity interface.
// It's expected that Response.Body is already wrapped with BodyReader.
type Response struct {
	Response *http.Response
}

func (r Response) GetForm() (map[string][]string, error) {
	// Response cannot contain a form.
	return nil, nil
}

func (r Response) GetJSON() (interface{}, error) {
	defer r.resetBody()
	dec := json.NewDecoder(r.Response.Body)
	var result interface{}
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return result, nil
}

func (r Response) GetIngress() bool {
	return false
}

func (r Response) GetBody() ([]byte, error) {
	defer r.resetBody()
	buf, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return buf, nil
}

func (r Response) GetCookies() ([]*http.Cookie, error) {
	return r.Response.Cookies(), nil
}

func (r Response) GetHeaders() (map[string][]string, error) {
	return r.Response.Header, nil
}

func (r Response) GetURL() (*url.URL, error) {
	if r.Response.Request != nil {
		return r.Response.Request.URL, nil
	}
	return nil, nil
}

func (r Response) GetRaw() ([]byte, error) {
	defer r.resetBody()
	data, err := httputil.DumpResponse(r.Response, true)
	if err != nil {
		return nil, fmt.Errorf("dumping request: %w", err)
	}
	return data, err
}

func (r Response) SetBody(data []byte) {
	r.resetBody()
	r.Response.Body = NewBodyReaderFromRaw(data)
}

func (r Response) resetBody() {
	if err := r.Response.Body.Close(); err != nil {
		logrus.Errorf("Error resetting response body: %v", err)
	}
}
