package filters

import (
	"fmt"
	"strings"

	"goxy/internal/common"
)

func NewCompositeAndRule(rs RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) < 2 {
		return nil, ErrInvalidRuleArgs
	}
	r := CompositeAndRule{rules: make([]Rule, 0, len(cfg.Args))}
	for _, name := range cfg.Args {
		rule, ok := rs.GetRule(name)
		if !ok {
			return nil, fmt.Errorf("invalid rule name: %s", name)
		}
		r.rules = append(r.rules, rule)
	}
	return r, nil
}

func NewCompositeNotRule(rs RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}

	name := cfg.Args[0]
	rule, ok := rs.GetRule(name)
	if !ok {
		return nil, fmt.Errorf("invalid rule name: %s", name)
	}

	return CompositeNotRule{rule: rule}, nil
}

type CompositeAndRule struct {
	rules []Rule
}

func (r CompositeAndRule) Apply(ctx *common.ProxyContext, v interface{}) (bool, error) {
	for _, rule := range r.rules {
		res, err := rule.Apply(ctx, v)
		if err != nil {
			return false, fmt.Errorf("error in rule %T: %w", rule, err)
		}
		if !res {
			return false, nil
		}
	}
	return true, nil
}

func (r CompositeAndRule) String() string {
	ruleNames := make([]string, 0, len(r.rules))
	for _, r := range r.rules {
		ruleNames = append(ruleNames, r.String())
	}
	return fmt.Sprintf("%s", strings.Join(ruleNames, " and "))
}

type CompositeNotRule struct {
	rule Rule
}

func (r CompositeNotRule) Apply(ctx *common.ProxyContext, v interface{}) (bool, error) {
	res, err := r.rule.Apply(ctx, v)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", r.rule, err)
	}
	return !res, nil
}

func (r CompositeNotRule) String() string {
	return fmt.Sprintf("not (%s)", r.rule)
}
