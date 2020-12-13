package filters

import (
	"fmt"
	"goxy/internal/common"
	"strings"
)

type Rule interface {
	Apply(ctx *common.ProxyContext, buf []byte, ingress bool) (bool, error)
}

type RuleCreator func(rs RuleSet, cfg common.RuleConfig) (Rule, error)
type RuleWrapperCreator func(rule Rule, cfg common.RuleConfig) Rule

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
	rs := RuleSet{Rules: make(map[string]Rule)}

	for _, rc := range cfg {
		if strings.HasPrefix(rc.Type, "tcp::") {
			tokens := strings.Split(rc.Type, "::")
			if len(tokens) < 2 {
				return nil, fmt.Errorf("invalid rule: %s", rc.Type)
			}

			// the last rule in chain must be either the composite rule or some rule creator
			lastToken := tokens[len(tokens)-1]

			var rule Rule
			var err error
			if creator, ok := DefaultRuleCreators[lastToken]; ok {
				if rule, err = creator(rs, rc); err != nil {
					return nil, fmt.Errorf("creating rule %s: %w", lastToken, err)
				}
			} else {
				return nil, fmt.Errorf("invalid rule %s: last token invalid", rc.Type)
			}

			for i := len(tokens) - 2; i > 0; i -= 1 {
				wrapperName := tokens[i]
				wrapper, ok := DefaultRuleWrappers[wrapperName]
				if !ok {
					return nil, fmt.Errorf("invalid wrapper name: %s", wrapperName)
				}
				rule = wrapper(rule, rc)
			}

			rs.Rules[rc.Name] = rule
		}
	}

	return &rs, nil
}

type Filter struct {
	Rule    Rule
	Verdict common.Verdict
}
