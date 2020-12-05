package filters

import (
	"fmt"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"strings"
)

type Rule interface {
	Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error)
}

type RuleCreator func(cfg *common.RuleConfig) (Rule, error)
type CompositeRuleCreator func(rs *RuleSet, cfg *common.RuleConfig) (Rule, error)

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
		tokens := strings.Split(rc.Type, "::")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("invalid rule: %s", rc.Type)
		}
		if tokens[0] == "http" {
			return nil, fmt.Errorf("invalid rule: %s", rc.Type)
		}
	}

	return rs, nil
}

type Filter struct {
	Rule    Rule
	Verdict common.Verdict
}
