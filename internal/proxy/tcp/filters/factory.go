package filters

import (
	"fmt"
	"goxy/internal/common"
)

func NotRuleFactory(creator RuleCreator) RuleCreator {
	return func(cfg *common.RuleConfig) (Rule, error) {
		inner, err := creator(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating inner rule: %w", err)
		}
		return &CompositeNotRule{Rule: inner}, nil
	}
}

func IngressRuleFactory(creator RuleCreator) RuleCreator {
	return func(cfg *common.RuleConfig) (Rule, error) {
		inner, err := creator(cfg)
		if err != nil {
			return nil, fmt.Errorf("creating inner rule: %w", err)
		}
		rule := &CompositeAndRule{Rules: []Rule{
			new(IngressRule),
			inner,
		}}
		return rule, nil
	}
}
