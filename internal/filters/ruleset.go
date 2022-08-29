package filters

import (
	"fmt"
	"strings"

	"goxy/internal/common"
)

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
		tokens := strings.Split(rc.Type, "::")
		if len(tokens) == 0 {
			return nil, fmt.Errorf("invalid rule: %s", rc.Type)
		}

		var (
			err       error
			rule      Rule
			converted bool
		)

		lastToken := tokens[len(tokens)-1]
		if creator, ok := DefaultRuleCreators[lastToken]; ok {
			// default rule type
			if rule, err = creator(rs, rc); err != nil {
				return nil, fmt.Errorf("creating rule %s: %w", lastToken, err)
			}
		} else {
			return nil, fmt.Errorf("invalid rule %s: last token invalid", rc.Type)
		}

		for i := len(tokens) - 2; i >= 0; i -= 1 {
			ruleName := tokens[i]
			if entityConverter, ok := DefaultEntityConverters[ruleName]; ok {
				if converted {
					return nil, fmt.Errorf("double conversion in %s", rc.Type)
				}
				converted = true

				// regular rules started, need to convert.
				// if field is specified for rule, we need to wrap it into FieldWrapper.
				if rc.Field != "" {
					rule = NewFieldWrapper(rule, rc)
				}
				rule = ConvertingWrapper{
					rule:      rule,
					converter: entityConverter,
				}
				continue
			} else if wrapperCreator, ok := DefaultRuleWrappers[ruleName]; ok {
				rule = wrapperCreator(rule, rc)
			} else {
				return nil, fmt.Errorf("unexpected token %s for %s", ruleName, rc.Type)
			}
		}
		if !converted {
			rule = ConvertingWrapper{
				rule:      rule,
				converter: RawEntityConverter{},
			}
		}
		rs.Rules[rc.Name] = rule
	}

	return &rs, nil
}
