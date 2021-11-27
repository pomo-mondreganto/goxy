package filters

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"goxy/internal/common"
	"goxy/internal/wrapper"
)

var (
	ErrInvalidRuleArgs  = errors.New("invalid rule arguments")
	ErrInvalidInputType = errors.New("invalid input data")
)

func NewContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return ContainsRule{cfg.Args[0]}, nil
}
func NewIContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return IContainsRule{strings.ToLower(cfg.Args[0])}, nil
}

func NewRegexRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	re, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}
	return RegexRule{re}, nil
}

type IngressRule struct{}

func (r IngressRule) Apply(_ *common.ProxyContext, v interface{}) (bool, error) {
	e, ok := v.(wrapper.Entity)
	if !ok {
		return false, fmt.Errorf("can only work on entities, not %v %T", v, v)
	}
	return e.GetIngress(), nil
}

func (r IngressRule) String() string {
	return "ingress"
}

type ContainsRule struct {
	value string
}

func (r ContainsRule) Apply(_ *common.ProxyContext, v interface{}) (bool, error) {
	switch v.(type) {
	case wrapper.Entity:
		body, err := v.(wrapper.Entity).GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		return bytes.Contains(body, []byte(r.value)), nil
	case map[string]interface{}:
		for k := range v.(map[string]interface{}) {
			if k == r.value {
				return true, nil
			}
		}
		return false, nil
	case []interface{}:
		for _, v := range v.([]interface{}) {
			switch v.(type) {
			case string:
				if strings.Contains(v.(string), r.value) {
					return true, nil
				}
			case []byte:
				if bytes.Contains(v.([]byte), []byte(r.value)) {
					return true, nil
				}
			}
		}
		return false, nil
	case []string:
		for _, s := range v.([]string) {
			if s == r.value {
				return true, nil
			}
		}
		return false, nil
	case string:
		return strings.Contains(v.(string), r.value), nil
	case []byte:
		return bytes.Contains(v.([]byte), []byte(r.value)), nil
	}

	return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
}

func (r ContainsRule) String() string {
	return fmt.Sprintf("contains '%s'", r.value)
}

type IContainsRule struct {
	value string
}

func (r IContainsRule) Apply(_ *common.ProxyContext, v interface{}) (bool, error) {
	switch v.(type) {
	case wrapper.Entity:
		body, err := v.(wrapper.Entity).GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		return bytes.Contains(bytes.ToLower(body), []byte(r.value)), nil
	case map[string]interface{}:
		for k := range v.(map[string]interface{}) {
			if strings.ToLower(k) == r.value {
				return true, nil
			}
		}
		return false, nil
	case []interface{}:
		for _, v := range v.([]interface{}) {
			switch v.(type) {
			case string:
				if strings.Contains(strings.ToLower(v.(string)), r.value) {
					return true, nil
				}
			case []byte:
				if bytes.Contains(bytes.ToLower(v.([]byte)), []byte(r.value)) {
					return true, nil
				}
			}
		}
		return false, nil

	case []string:
		for _, s := range v.([]string) {
			if strings.ToLower(s) == r.value {
				return true, nil
			}
		}
		return false, nil

	case string:
		return strings.Contains(strings.ToLower(v.(string)), r.value), nil
	case []byte:
		return bytes.Contains(bytes.ToLower(v.([]byte)), []byte(r.value)), nil
	}

	return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
}

func (r IContainsRule) String() string {
	return fmt.Sprintf("icontains '%s'", r.value)
}

type RegexRule struct {
	re *regexp.Regexp
}

func (r RegexRule) Apply(_ *common.ProxyContext, v interface{}) (bool, error) {
	switch v.(type) {
	case wrapper.Entity:
		body, err := v.(wrapper.Entity).GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		return r.re.Match(body), nil
	case string:
		return r.re.MatchString(v.(string)), nil
	case []byte:
		return r.re.Match(v.([]byte)), nil
	}

	return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
}

func (r RegexRule) String() string {
	return fmt.Sprintf("regex '%s'", r.re)
}
