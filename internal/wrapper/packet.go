package wrapper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Packet is a wrapper around raw data (e.g. tcp packet) implementing Entity interface.
type Packet struct {
	Content []byte
	Ingress bool
}

func (p Packet) GetForm() (map[string][]string, error) {
	return nil, ErrNotSupported
}

func (p Packet) GetJSON() (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal(p.Content, &result); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	return result, nil
}

func (p Packet) GetBody() ([]byte, error) {
	dst := make([]byte, len(p.Content))
	copy(dst, p.Content)
	return dst, nil
}

func (p Packet) GetIngress() bool {
	return p.Ingress
}

func (p Packet) GetCookies() ([]*http.Cookie, error) {
	return nil, ErrNotSupported
}

func (p Packet) GetHeaders() (map[string][]string, error) {
	return nil, ErrNotSupported
}

func (p Packet) GetURL() (*url.URL, error) {
	return nil, ErrNotSupported
}

func (p Packet) GetRaw() ([]byte, error) {
	return p.Content, nil
}

func (p Packet) SetBody(data []byte) {
	p.Content = make([]byte, len(data))
	copy(p.Content, data)
}
