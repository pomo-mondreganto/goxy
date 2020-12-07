package filters

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"strings"
)

var (
	ErrInvalidRuleArgs  = errors.New("invalid rule arguments")
	ErrInvalidInputType = errors.New("invalid input data")
)

type IngressRule struct{}

func (r *IngressRule) Apply(_ *common.ProxyContext, e wrapper.Entity) (bool, error) {
	return e.GetIngress(), nil
}

type ContainsStringRawRule struct {
	Value string
}

func (r *ContainsStringRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			if v == r.Value {
				return true, nil
			}
		}
	case []interface{}:
		for _, v := range data.([]interface{}) {
			if v == r.Value {
				return true, nil
			}
		}
	case []string:
		for _, v := range data.([]string) {
			if v == r.Value {
				return true, nil
			}
		}
	case string:
		return strings.Contains(data.(string), r.Value), nil
	case []byte:
		return bytes.Contains(data.([]byte), []byte(r.Value)), nil
	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}

func NewContainsStringRawRule(cfg *common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return &ContainsStringRawRule{Value: cfg.Args[0]}, nil
}
