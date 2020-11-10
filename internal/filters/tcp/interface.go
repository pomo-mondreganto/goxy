package tcp

import (
	"fmt"
	"goxy/internal/common"
	"goxy/internal/common/tcp"
	"strings"
)

type Rule interface {
	Apply(buf []byte, ingress bool) (bool, error)
}

type RuleCreator func(cfg *common.RuleConfig) (Rule, error)
type CompositeRuleCreator func(rules map[string]Rule, cfg *common.RuleConfig) (Rule, error)

type RuleSet struct {
	Rules map[string]Rule
}

func (rs *RuleSet) GetRule(name string) (Rule, bool) {
	if rule, ok := rs.Rules[name]; ok {
		return rule, true
	}
	if rule, ok := DefaultRules[name]; ok {
		return rule, true
	}
	return nil, false
}

func NewRuleSet(cfg []common.RuleConfig) (*RuleSet, error) {
	rs := &RuleSet{Rules: make(map[string]Rule)}

	for _, rc := range cfg {
		if strings.HasPrefix(rc.Type, "tcp") {
			tokens := strings.Split(rc.Type, "::")
			if len(tokens) < 2 {
				return nil, fmt.Errorf("invalid rule type: %s", rc.Type)
			}

			name := tokens[1]
			if rule, ok := DefaultRules[name]; ok {
				rs.Rules[rc.Name] = rule
				continue
			}

			if creator, ok := DefaultRuleCreators[name]; ok {
				rule, err := creator(&rc)
				if err != nil {
					return nil, fmt.Errorf("error creating rule %s: %w", name, err)
				}
				rs.Rules[rc.Name] = rule
				continue
			}

			if creator, ok := DefaultCompositeRuleCreators[name]; ok {
				rule, err := creator(rs.Rules, &rc)
				if err != nil {
					return nil, fmt.Errorf("error creating rule %s: %w", name, err)
				}
				rs.Rules[rc.Name] = rule
				continue
			}

			return nil, fmt.Errorf("unknown rule type: %s", rc.Type)
		}
	}

	return rs, nil
}

type Filter struct {
	Rule    Rule
	Verdict tcp.Verdict
}
