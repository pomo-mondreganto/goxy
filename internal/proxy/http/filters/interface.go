package filters

import (
	"fmt"
	"go.uber.org/atomic"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
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

type Filter struct {
	Rule    Rule
	Verdict common.Verdict

	alert    atomic.Bool
	disabled atomic.Bool
}

func (f Filter) IsEnabled() bool {
	return !f.disabled.Load()
}

func (f Filter) GetAlert() bool {
	return f.alert.Load()
}

func (f *Filter) SetEnabled(enabled bool) {
	f.disabled.Store(!enabled)
}

func (f *Filter) SetAlert(alert bool) {
	f.alert.Store(alert)
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
