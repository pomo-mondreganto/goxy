package filters

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
)

type EntityConverter func(e wrapper.Entity) (interface{}, error)

type RawRuleConverter struct {
	rule            RawRule
	entityConverter EntityConverter
}

func (r *RawRuleConverter) Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error) {
	data, err := r.entityConverter(e)
	if err != nil {
		logrus.Debugf("Entity converter returned an error: %v", err)
		return false, nil
	}
	res, err := r.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", r.rule, err)
	}
	return res, nil
}

func NewRawRuleConverter(rule RawRule, ec EntityConverter) Rule {
	return &RawRuleConverter{rule, ec}
}

func JSONEntityConverter(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetJSON()
	if err != nil {
		return nil, fmt.Errorf("getting json: %w", err)
	}
	return data, nil
}

func CookiesEntityConverter(e wrapper.Entity) (interface{}, error) {
	cookies := e.GetCookies()
	result := make(map[string]string)
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result, nil
}

func HeadersEntityConverter(e wrapper.Entity) (interface{}, error) {
	return convertMapListString(e.GetHeaders()), nil
}

func QueryEntityConverter(e wrapper.Entity) (interface{}, error) {
	return convertMapListString(e.GetURL().Query()), nil
}

func BodyEntityConverter(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetBody()
	if err != nil {
		return nil, fmt.Errorf("getting body: %w", err)
	}
	return data, nil
}

func PathEntityConverter(e wrapper.Entity) (interface{}, error) {
	return e.GetURL().Path, nil
}

func FormEntityConverter(e wrapper.Entity) (interface{}, error) {
	data, err := e.GetForm()
	if err != nil {
		return nil, fmt.Errorf("getting form: %w", err)
	}
	return convertMapListString(data), nil
}

func convertMapListString(data map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}
	return result
}
