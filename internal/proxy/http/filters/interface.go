package filters

import (
	"fmt"
	"go.uber.org/atomic"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"strings"
)

type Rule interface {
	Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error)
	fmt.Stringer
}

type RawRule interface {
	Apply(ctx *common.ProxyContext, data interface{}) (bool, error)
	fmt.Stringer
}

type EntityConverter interface {
	Convert(e wrapper.Entity) (interface{}, error)
	fmt.Stringer
}

type RuleCreator func(rs RuleSet, cfg common.RuleConfig) (Rule, error)
type RawRuleCreator func(cfg common.RuleConfig) (RawRule, error)

type RuleWrapperCreator func(rule Rule, cfg common.RuleConfig) Rule
type RawRuleWrapperCreator func(rule RawRule, cfg common.RuleConfig) RawRule

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
		if strings.HasPrefix(rc.Type, "http::") {
			tokens := strings.Split(rc.Type, "::")
			if len(tokens) < 3 {
				return nil, fmt.Errorf("invalid rule: %s", rc.Type)
			}

			var (
				ok         bool
				err        error
				rawRule    RawRule
				rawCreator RawRuleCreator
				creator    RuleCreator
				rule       Rule
			)

			// Some block at the end of rules (can be empty) contains only raw rules.
			// Rules before are regular rules, and between these two blocks we need to insert
			// the RawRuleConverterWrapper.

			// the last rule in chain must be either the composite rule or raw rule.
			lastToken := tokens[len(tokens)-1]
			if creator, ok = DefaultRuleCreators[lastToken]; ok {
				// default rule type
				if rule, err = creator(rs, rc); err != nil {
					return nil, fmt.Errorf("creating rule %s: %w", lastToken, err)
				}
			} else if rawCreator, ok = DefaultRawRuleCreators[lastToken]; ok {
				// rule is regular raw rule
				if rawRule, err = rawCreator(rc); err != nil {
					return nil, fmt.Errorf("creating raw rule %s: %w", lastToken, err)
				}
			} else {
				return nil, fmt.Errorf("invalid rule %s: last token invalid", rc.Type)
			}

			for i := len(tokens) - 2; i > 0; i -= 1 {
				ruleName := tokens[i]
				if rawRule != nil {
					if wrapperCreator, ok := DefaultRawRuleWrappers[ruleName]; ok {
						rawRule = wrapperCreator(rawRule, rc)
						continue
					} else if entityConverter, ok := DefaultEntityConverters[ruleName]; ok {
						// regular rules started, need to convert.
						// if field is specified for rule, we need to wrap it into FieldWrapper.
						if rc.Field != "" {
							rawRule = NewFieldWrapper(rawRule, rc)
						}
						rule = NewRawRuleConverter(rawRule, entityConverter)
						rawRule = nil
						continue
					} else {
						return nil, fmt.Errorf("no entity converter with name %s for rule %s", ruleName, rc.Type)
					}
				}

				if wrapperCreator, ok := DefaultRuleWrappers[ruleName]; ok {
					rule = wrapperCreator(rule, rc)
				} else {
					return nil, fmt.Errorf("no wrapper for name %s", ruleName)
				}
			}

			if rawRule != nil {
				return nil, fmt.Errorf("entity converter for %s not specified", rc.Type)
			}

			rs.Rules[rc.Name] = rule
		}
	}

	return &rs, nil
}

type Filter struct {
	Rule    Rule
	Verdict common.Verdict

	disabled atomic.Bool
}

func (f Filter) IsEnabled() bool {
	return !f.disabled.Load()
}

func (f *Filter) SetEnabled(enabled bool) {
	f.disabled.Store(!enabled)
}

func (f Filter) GetRule() common.Rule {
	return f.Rule
}

func (f Filter) GetVerdict() common.Verdict {
	return f.Verdict
}

func (f Filter) String() string {
	return fmt.Sprintf("if %s: %s", f.Rule, f.Verdict)
}
