package filters

import (
	"fmt"
	"goxy/internal/common"
	"strings"
)

type Rule interface {
	Apply(ctx *common.ProxyContext, buf []byte, ingress bool) (bool, error)
}

type RuleCreator func(cfg *common.RuleConfig) (Rule, error)
type CompositeRuleCreator func(rs *RuleSet, cfg *common.RuleConfig) (Rule, error)
type RuleFactory func(creator RuleCreator) RuleCreator

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

func NewRuleSet(cfg []*common.RuleConfig) (*RuleSet, error) {
	rs := &RuleSet{Rules: make(map[string]Rule)}

	for _, rc := range cfg {
		if strings.HasPrefix(rc.Type, "tcp::") {
			tokens := strings.Split(rc.Type, "::")
			if len(tokens) < 2 {
				return nil, fmt.Errorf("invalid rule: %s", rc.Type)
			}
			// Create base rule creator first
			baseName := tokens[len(tokens)-1]

			creator, ok := DefaultRuleCreators[baseName]
			if !ok {
				var compositeCreator CompositeRuleCreator
				if compositeCreator, ok = DefaultCompositeRuleCreators[baseName]; ok {
					creator = compositeCreator.PartialConvert(rs)
				}
			}
			if !ok {
				return nil, fmt.Errorf("unknown rule type: %s", rc.Type)
			}

			for i := len(tokens) - 2; i > 0; i -= 1 {
				factoryName := tokens[i]
				factory, ok := DefaultRuleFactories[factoryName]
				if !ok {
					return nil, fmt.Errorf("invalid factory name: %s", factoryName)
				}
				creator = factory(creator)
			}

			rule, err := creator(rc)
			if err != nil {
				return nil, fmt.Errorf("creating rule %s: %w", rc.Type, err)
			}
			rs.Rules[rc.Name] = rule
		}
	}

	return rs, nil
}

type Filter struct {
	Rule    Rule
	Verdict common.Verdict
}
