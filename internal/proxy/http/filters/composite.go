package filters

import (
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
)

func (c CompositeRuleCreator) PartialConvert(rs *RuleSet) RuleCreator {
	return func(cfg *common.RuleConfig) (Rule, error) {
		return c(rs, cfg)
	}
}

type CompositeAndRule struct {
	Rules []Rule
}

func (r *CompositeAndRule) Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error) {
	for _, rule := range r.Rules {
		res, err := rule.Apply(ctx, e)
		if err != nil {
			return false, fmt.Errorf("error in rule %T: %w", rule, err)
		}
		if !res {
			return false, nil
		}
	}
	return true, nil
}

func NewCompositeAndRule(rs *RuleSet, cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) < 2 {
		return nil, ErrInvalidRuleArgs
	}
	r := &CompositeAndRule{Rules: make([]Rule, 0, len(cfg.Args))}
	for _, name := range cfg.Args {
		rule, ok := rs.GetRule(name)
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

func (r *CompositeNotRule) Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error) {
	res, err := r.Rule.Apply(ctx, e)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", r.Rule, err)
	}
	return !res, nil
}

func NewCompositeNotRule(rs *RuleSet, cfg *common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}

	name := cfg.Args[0]
	rule, ok := rs.GetRule(name)
	if !ok {
		return nil, fmt.Errorf("invalid rule name: %s", name)
	}

	return &CompositeNotRule{Rule: rule}, nil
}
