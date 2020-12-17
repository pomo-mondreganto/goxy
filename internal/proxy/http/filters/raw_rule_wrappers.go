package filters

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"strings"
)

func NewAnyWrapper(r RawRule, _ common.RuleConfig) RawRule {
	return AnyWrapper{r}
}

func NewArrayWrapper(r RawRule, _ common.RuleConfig) RawRule {
	return ArrayWrapper{r}
}

func NewFieldWrapper(r RawRule, cfg common.RuleConfig) RawRule {
	fieldChain := strings.Split(cfg.Field, ".")
	return FieldWrapper{r, fieldChain}
}

func NewNotWrapperRaw(r RawRule, _ common.RuleConfig) RawRule {
	return RawNotWrapper{r}
}

func NewRawRuleConverter(rule RawRule, ec EntityConverter) Rule {
	return RawRuleConverterWrapper{rule, ec}
}

type AnyWrapper struct {
	rule RawRule
}

func (w AnyWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case map[string]interface{}:
		for _, v := range data.(map[string]interface{}) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}

	case []interface{}:
		for _, v := range data.([]interface{}) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}

	case []string:
		for _, v := range data.([]string) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}

	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}

func (w AnyWrapper) String() string {
	return fmt.Sprintf("any %s", w.rule)
}

type ArrayWrapper struct {
	rule RawRule
}

func (w ArrayWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	switch data.(type) {
	case []interface{}:
		res, err := w.rule.Apply(ctx, data.([]interface{}))
		if err != nil {
			return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
		}
		return res, nil
	case []string:
		res, err := w.rule.Apply(ctx, data.([]string))
		if err != nil {
			return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
		}
		return res, nil
	default:
		return false, nil
	}
}

func (w ArrayWrapper) String() string {
	return fmt.Sprintf("is array and %s", w.rule)
}

type FieldWrapper struct {
	rule       RawRule
	fieldChain []string
}

func (w FieldWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	result := data
	for _, f := range w.fieldChain {
		switch result.(type) {
		case map[string]interface{}:
			converted := result.(map[string]interface{})
			next, ok := converted[f]
			if !ok {
				return false, nil
			}
			result = next
		default:
			return false, nil
		}
	}
	res, err := w.rule.Apply(ctx, result)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}

func (w FieldWrapper) String() string {
	fieldRepr := strings.Join(w.fieldChain, ".")
	return fmt.Sprintf("field '%s' %s", fieldRepr, w.rule)
}

type RawNotWrapper struct {
	rule RawRule
}

func (w RawNotWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	res, err := w.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return !res, nil
}

func (w RawNotWrapper) String() string {
	return fmt.Sprintf("not (%s)", w.rule)
}

type RawRuleConverterWrapper struct {
	rule RawRule
	ec   EntityConverter
}

func (w RawRuleConverterWrapper) Apply(ctx *common.ProxyContext, e wrapper.Entity) (bool, error) {
	data, err := w.ec.Convert(e)
	if err != nil {
		logrus.Debugf("Entity converter returned an error: %v", err)
		return false, nil
	}
	res, err := w.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}

func (w RawRuleConverterWrapper) String() string {
	return fmt.Sprintf("%s %s", w.ec, w.rule)
}
