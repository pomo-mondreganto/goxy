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
	return ContainsRule{
		name:  cfg.Name,
		value: cfg.Args[0],
	}, nil
}
func NewIContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	return IContainsRule{
		name:  cfg.Name,
		value: strings.ToLower(cfg.Args[0]),
	}, nil
}

func NewRegexRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	re, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}
	return RegexRule{
		name: cfg.Name,
		re:   re,
	}, nil
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
	name  string
	value string
}

func (r ContainsRule) Apply(pctx *common.ProxyContext, v interface{}) (bool, error) {
	var count int
	switch t := v.(type) {
	case wrapper.Entity:
		body, err := t.GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		count = bytes.Count(body, []byte(r.value))
	case map[string]interface{}:
		for k := range t {
			if k == r.value {
				count++
			}
		}
	case []interface{}:
		for _, v := range t {
			switch el := v.(type) {
			case string:
				count += strings.Count(el, r.value)
			case []byte:
				count += bytes.Count(el, []byte(r.value))
			}
		}
	case []string:
		for _, s := range t {
			if s == r.value {
				count++
			}
		}
	case string:
		count = strings.Count(t, r.value)
	case []byte:
		count = bytes.Count(t, []byte(r.value))
	default:
		return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
	}
	pctx.AddToCounter(r.name, count)
	return count > 0, nil
}

func (r ContainsRule) String() string {
	return fmt.Sprintf("contains '%s'", r.value)
}

type IContainsRule struct {
	name  string
	value string
}

func (r IContainsRule) Apply(pctx *common.ProxyContext, v interface{}) (bool, error) {
	var count int
	switch t := v.(type) {
	case wrapper.Entity:
		body, err := t.GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		count = bytes.Count(bytes.ToLower(body), []byte(r.value))
	case map[string]interface{}:
		for k := range t {
			if strings.ToLower(k) == r.value {
				count++
			}
		}
	case []interface{}:
		for _, v := range t {
			switch el := v.(type) {
			case string:
				if strings.ToLower(el) == r.value {
					count++
				}
			case []byte:
				if bytes.EqualFold(el, []byte(r.value)) {
					count++
				}
			}
		}
	case []string:
		for _, s := range t {
			if strings.ToLower(s) == r.value {
				count++
			}
		}
	case string:
		count = strings.Count(strings.ToLower(t), r.value)
	case []byte:
		count = bytes.Count(bytes.ToLower(t), []byte(r.value))
	default:
		return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
	}
	pctx.AddToCounter(r.name, count)
	return count > 0, nil
}

func (r IContainsRule) String() string {
	return fmt.Sprintf("icontains '%s'", r.value)
}

type RegexRule struct {
	name string
	re   *regexp.Regexp
}

func (r RegexRule) Apply(pctx *common.ProxyContext, v interface{}) (bool, error) {
	var count int
	switch t := v.(type) {
	case wrapper.Entity:
		body, err := t.GetBody()
		if err != nil {
			return false, fmt.Errorf("getting body: %w", err)
		}
		count = len(r.re.FindAllIndex(body, -1))
	case string:
		count = len(r.re.FindAllStringIndex(t, -1))
	case []byte:
		count = len(r.re.FindAllIndex(t, -1))
	default:
		return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
	}
	pctx.AddToCounter(r.name, count)
	return count > 0, nil
}

func (r RegexRule) String() string {
	return fmt.Sprintf("regex '%s'", r.re)
}
