package filters

import (
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
)

type EntityConverter func(e wrapper.Entity) interface{}

type RawRuleConverter struct {
	rule            RawRule
	entityConverter EntityConverter
}

func (r *RawRuleConverter) Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error) {
	data := r.entityConverter(e)
	res, err := r.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", r.rule, err)
	}
	return res, nil
}

func NewRawRuleConverter(rule RawRule, ec EntityConverter) Rule {
	return &RawRuleConverter{rule, ec}
}

func JSONEntityConverter(e wrapper.Entity) interface{} {
	return e.GetJSON()
}

func CookiesEntityConverter(e wrapper.Entity) interface{} {
	cookies := e.GetCookies()
	result := make(map[string]string)
	for _, cookie := range cookies {
		result[cookie.Name] = cookie.Value
	}
	return result
}

func QueryEntityConverter(e wrapper.Entity) interface{} {
	return e.GetURL().Query()
}
