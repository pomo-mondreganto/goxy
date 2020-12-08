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

func NewContainsRawRule(cfg *common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return &ContainsRawRule{cfg.Args[0]}, nil
}

func NewIContainsRawRule(cfg *common.RuleConfig) (RawRule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return &IContainsRawRule{strings.ToLower(cfg.Args[0])}, nil
}

func NewRegexRawRule(cfg *common.RuleConfig) (RawRule, error) {
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

type ContainsRawRule struct {
	Value string
}

func (r *ContainsRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			switch v.(type) {
			case string:
				if strings.Contains(v.(string), r.Value) {
					return true, nil
				}
			case []byte:
				if bytes.Contains(data.([]byte), []byte(r.Value)) {
					return true, nil
				}
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

type IContainsRawRule struct {
	Value string
}

func (r *IContainsRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			switch v.(type) {
			case string:
				if strings.Contains(strings.ToLower(v.(string)), r.Value) {
					return true, nil
				}
			case []byte:
				if bytes.Contains(bytes.ToLower(data.([]byte)), []byte(r.Value)) {
					return true, nil
				}
			}
		}
	case []string:
		for _, v := range data.([]string) {
			if strings.ToLower(v) == r.Value {
				return true, nil
			}
		}
	case string:
		return strings.Contains(strings.ToLower(data.(string)), r.Value), nil
	case []byte:
		return bytes.Contains(bytes.ToLower(data.([]byte)), []byte(r.Value)), nil
	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}

type RegexRawRule struct {
	re *regexp.Regexp
}

func (r *RegexRawRule) Apply(_ *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			switch v.(type) {
			case string:
				if r.re.MatchString(v.(string)) {
					return true, nil
				}
			case []byte:
				if r.re.Match(v.([]byte)) {
					return true, nil
				}
			}
		}
	case []string:
		for _, v := range data.([]string) {
			if r.re.MatchString(v) {
				return true, nil
			}
		}
	case string:
		return r.re.MatchString(data.(string)), nil
	case []byte:
		return r.re.Match(data.([]byte)), nil
	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}
