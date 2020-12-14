package filters

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"regexp"
	"strings"
)

var (
	ErrInvalidRuleArgs  = errors.New("invalid rule arguments")
	ErrInvalidInputType = errors.New("invalid input data")
)

func NewContainsRawRule(cfg common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return &ContainsRawRule{cfg.Args[0]}, nil
}

func NewIContainsRawRule(cfg common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return &IContainsRawRule{strings.ToLower(cfg.Args[0])}, nil
}

func NewRegexRawRule(cfg common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	re, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}
	return &RegexRawRule{re}, nil
}

type IngressRule struct{}

func (r *IngressRule) Apply(_ *common.ProxyContext, e wrapper.Entity) (bool, error) {
	return e.GetIngress(), nil
}

func (r *IngressRule) String() string {
	return "ingress"
}

type ContainsRawRule struct {
	value string
}

func (r *ContainsRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	stringHandler := func(s string) bool {
		return strings.Contains(s, r.value)
	}
	bytesHandler := func(b []byte) bool {
		return bytes.Contains(b, []byte(r.value))
	}
	return processGenericMatchRule(stringHandler, bytesHandler, data)
}

func (r *ContainsRawRule) String() string {
	return fmt.Sprintf("contains '%s'", r.value)
}

type IContainsRawRule struct {
	value string
}

func (r *IContainsRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	stringHandler := func(s string) bool {
		return strings.Contains(strings.ToLower(s), r.value)
	}
	bytesHandler := func(b []byte) bool {
		return bytes.Contains(bytes.ToLower(b), []byte(r.value))
	}
	return processGenericMatchRule(stringHandler, bytesHandler, data)
}

func (r *IContainsRawRule) String() string {
	return fmt.Sprintf("icontains '%s'", r.value)
}

type RegexRawRule struct {
	re *regexp.Regexp
}

func (r *RegexRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	stringHandler := func(s string) bool {
		return r.re.MatchString(s)
	}
	bytesHandler := func(b []byte) bool {
		return r.re.Match(b)
	}
	return processGenericMatchRule(stringHandler, bytesHandler, data)
}

func (r *RegexRawRule) String() string {
	return fmt.Sprintf("regex '%s'", r.re)
}

func processGenericMatchRule(sh func(string) bool, bh func([]byte) bool, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			switch v.(type) {
			case string:
				if sh(v.(string)) {
					return true, nil
				}
			case []byte:
				if bh(v.([]byte)) {
					return true, nil
				}
			}
		}
	case []interface{}:
		for _, v := range data.([]interface{}) {
			switch v.(type) {
			case string:
				if sh(v.(string)) {
					return true, nil
				}
			case []byte:
				if bh(v.([]byte)) {
					return true, nil
				}
			}
		}
	case []string:
		for _, v := range data.([]string) {
			if sh(v) {
				return true, nil
			}
		}
	case string:
		return sh(data.(string)), nil
	case []byte:
		return bh(data.([]byte)), nil
	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}
