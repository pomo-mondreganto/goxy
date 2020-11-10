package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"regexp"
)

var (
	ErrInvalidRuleArgs = errors.New("invalid rule arguments")
)

type RegexRule struct {
	Regex *regexp.Regexp
}

func (r *RegexRule) Apply(buf []byte, _ bool) (bool, error) {
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

func (r *ContainsRule) Apply(buf []byte, _ bool) (bool, error) {
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

func (r *IngressRule) Apply(_ []byte, ingress bool) (bool, error) {
	return ingress, nil
}

type CompositeAndRule struct {
	Rules []Rule
}

func (r *CompositeAndRule) Apply(buf []byte, ingress bool) (bool, error) {
	for _, rule := range r.Rules {
		res, err := rule.Apply(buf, ingress)
		if err != nil {
			return false, fmt.Errorf("error in rule %T: %w", rule, err)
		}
		if !res {
			return false, nil
		}
	}
	return true, nil
}

func NewCompositeAndRule(rules map[string]Rule, cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) < 2 {
		return nil, ErrInvalidRuleArgs
	}
	r := &CompositeAndRule{Rules: make([]Rule, 0, len(cfg.Args))}
	for _, name := range cfg.Args {
		rule, ok := rules[name]
		if !ok {
			return nil, fmt.Errorf("invalid rule name: %s", name)
		}
		r.Rules = append(r.Rules, rule)
	}
	return r, nil
}

type CompositeNotRule struct {
	Rule Rule
}

func (r *CompositeNotRule) Apply(buf []byte, ingress bool) (bool, error) {
	res, err := r.Rule.Apply(buf, ingress)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", r.Rule, err)
	}
	return !res, nil
}

func NewCompositeNotRule(rules map[string]Rule, cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}

	name := cfg.Args[0]
	rule, ok := rules[name]
	if !ok {
		return nil, fmt.Errorf("invalid rule name: %s", name)
	}

	return &CompositeNotRule{Rule: rule}, nil
}
