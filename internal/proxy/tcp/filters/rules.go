package filters

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidRuleArgs = errors.New("invalid rule arguments")
)

func NewRegexRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	return &RegexRule{Regex: r}, nil
}

func NewContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r := new(ContainsRule)
	for _, s := range cfg.Args {
		r.Values = append(r.Values, []byte(s))
	}
	return r, nil
}

func NewIContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r := new(ContainsRule)
	for _, s := range cfg.Args {
		r.Values = append(r.Values, []byte(strings.ToLower(s)))
	}
	return r, nil
}

func NewIngressRule(_ RuleSet, _ common.RuleConfig) (Rule, error) {
	return new(IngressRule), nil
}

func NewCounterGTRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 2 {
		return nil, ErrInvalidRuleArgs
	}
	val, err := strconv.Atoi(cfg.Args[1])
	if err != nil {
		return nil, fmt.Errorf("parsing value: %w", err)
	}
	r := &CounterGTRule{
		Key:   cfg.Args[0],
		Value: val,
	}
	return r, nil
}

type RegexRule struct {
	Regex *regexp.Regexp
}

func (r *RegexRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	return r.Regex.Match(buf), nil
}

type ContainsRule struct {
	Values [][]byte
}

func (r *ContainsRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	for _, v := range r.Values {
		if bytes.Contains(buf, v) {
			return true, nil
		}
	}
	return false, nil
}

type IContainsRule struct {
	Values [][]byte
}

func (r *IContainsRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	for _, v := range r.Values {
		if bytes.Contains(bytes.ToLower(buf), v) {
			return true, nil
		}
	}
	return false, nil
}

type IngressRule struct{}

func (r *IngressRule) Apply(_ *common.ProxyContext, _ []byte, ingress bool) (bool, error) {
	return ingress, nil
}

type CounterGTRule struct {
	Key   string
	Value int
}

func (r *CounterGTRule) Apply(ctx *common.ProxyContext, _ []byte, _ bool) (bool, error) {
	return ctx.GetCounter(r.Key) > r.Value, nil
}
