package filters

import (
	"fmt"

	"goxy/internal/wrapper"
)

type JsonEntityConverter struct{}

func (c JsonEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetJSON()
	if err != nil {
		return nil, fmt.Errorf("getting json: %w", err)
	}
	return data, nil
}

func (c JsonEntityConverter) String() string {
	return "json"
}

type CookiesEntityConverter struct{}

func (c CookiesEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	cookies, err := e.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("getting cookies: %w", err)
	}
	result := make(map[string]string)
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result, nil
}

func (c CookiesEntityConverter) String() string {
	return "cookie"
}

type HeadersEntityConverter struct{}

func (c HeadersEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	headers, err := e.GetHeaders()
	if err != nil {
		return nil, fmt.Errorf("getting headers: %w", err)
	}
	return convertMapListString(headers), nil
}

func (c HeadersEntityConverter) String() string {
	return "headers"
}

type QueryEntityConverter struct{}

func (c QueryEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	url, err := e.GetURL()
	if err != nil {
		return nil, fmt.Errorf("getting url: %w", err)
	}
	return convertMapListString(url.Query()), nil
}

func (c QueryEntityConverter) String() string {
	return "query"
}

type BodyEntityConverter struct{}

func (c BodyEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetBody()
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	return data, nil
}

func (c BodyEntityConverter) String() string {
	return "body"
}

type PathEntityConverter struct{}

func (c PathEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	url, err := e.GetURL()
	if err != nil {
		return nil, fmt.Errorf("getting url: %w", err)
	}
	return url.Path, nil
}

func (c PathEntityConverter) String() string {
	return "path"
}

type FormEntityConverter struct{}

func (c FormEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetForm()
	if err != nil {
		return nil, fmt.Errorf("getting form: %w", err)
	}
	return convertMapListString(data), nil
}

func (c FormEntityConverter) String() string {
	return "form"
}

type RawEntityConverter struct{}

func (c RawEntityConverter) Convert(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetRaw()
	if err != nil {
		return nil, fmt.Errorf("getting raw data: %w", err)
	}
	return data, nil
}

func (c RawEntityConverter) String() string {
	return "raw"
}

func convertMapListString(data map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	return result
}
