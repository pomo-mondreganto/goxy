package filters

import (
	"fmt"
	"goxy/internal/common"
)

type CompositeAndRule struct {
	Rules []Rule
}

func (r *CompositeAndRule) Apply(ctx *common.ConnectionContext, buf []byte, ingress bool) (bool, error) {
	for _, rule := range r.Rules {
		res, err := rule.Apply(ctx, buf, ingress)
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

func (r *CompositeNotRule) Apply(ctx *common.ConnectionContext, buf []byte, ingress bool) (bool, error) {
	res, err := r.Rule.Apply(ctx, buf, ingress)
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
