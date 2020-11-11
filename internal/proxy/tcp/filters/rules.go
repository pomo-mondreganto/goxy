package filters

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"regexp"
	"strconv"
)

var (
	ErrInvalidRuleArgs = errors.New("invalid rule arguments")
)

type RegexRule struct {
	Regex *regexp.Regexp
}

func (r *RegexRule) Apply(_ *common.ConnectionContext, buf []byte, _ bool) (bool, error) {
	return r.Regex.Match(buf), nil
}

func NewRegexRule(cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	return &RegexRule{Regex: r}, nil
}

type ContainsRule struct {
	Values [][]byte
}

func (r *ContainsRule) Apply(_ *common.ConnectionContext, buf []byte, _ bool) (bool, error) {
	for _, v := range r.Values {
		if bytes.Contains(buf, v) {
			return true, nil
		}
	}
	return false, nil
}

func NewContainsRule(cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r := new(ContainsRule)
	for _, s := range cfg.Args {
		r.Values = append(r.Values, []byte(s))
	}
	return r, nil
}

type IngressRule struct{}

func (r *IngressRule) Apply(_ *common.ConnectionContext, _ []byte, ingress bool) (bool, error) {
	return ingress, nil
}

type CounterGTRule struct {
	Key   string
	Value int
}

func (r *CounterGTRule) Apply(ctx *common.ConnectionContext, _ []byte, _ bool) (bool, error) {
	return ctx.GetCounter(r.Key) > r.Value, nil
}

func NewCounterGTRule(cfg *common.RuleConfig) (Rule, error) {
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
