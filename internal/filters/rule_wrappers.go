package filters

import (
	"fmt"
	"strings"

	"goxy/internal/common"
	"goxy/internal/wrapper"
)

func NewIngressWrapper(r Rule, _ common.RuleConfig) Rule {
	return CompositeAndRule{
		rules: []Rule{IngressRule{}, r},
	}
}

func NewEgressWrapper(r Rule, cfg common.RuleConfig) Rule {
	return CompositeNotRule{rule: NewIngressWrapper(r, cfg)}
}

func NewNotWrapper(r Rule, _ common.RuleConfig) Rule {
	return CompositeNotRule{rule: r}
}

func NewAnyWrapper(r Rule, _ common.RuleConfig) Rule {
	return AnyWrapper{r}
}

func NewFieldWrapper(r Rule, cfg common.RuleConfig) Rule {
	fieldChain := strings.Split(cfg.Field, ".")
	return FieldWrapper{r, fieldChain}
}

type AnyWrapper struct {
	rule Rule
}

func (w AnyWrapper) Apply(ctx *common.ProxyContext, v interface{}) (bool, error) {
	switch v.(type) {
	case map[string]interface{}:
		for _, v := range v.(map[string]interface{}) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("in rule %v: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}
		return false, nil

	case []interface{}:
		for _, v := range v.([]interface{}) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("in rule %v: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}
		return false, nil

	case []string:
		for _, v := range v.([]string) {
			res, err := w.rule.Apply(ctx, v)
			if err != nil {
				return false, fmt.Errorf("in rule %v: %w", w.rule, err)
			}
			if res {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("data type %T: %w", v, ErrInvalidInputType)
}

func (w AnyWrapper) String() string {
	return fmt.Sprintf("any %s", w.rule)
}

type FieldWrapper struct {
	rule       Rule
	fieldChain []string
}

func (w FieldWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	result := data
	for _, f := range w.fieldChain {
		switch result.(type) {
		case map[string]interface{}:
			next, ok := result.(map[string]interface{})[f]
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
		return false, fmt.Errorf("in rule %v: %w", w.rule, err)
	}
	return res, nil
}

func (w FieldWrapper) String() string {
	fieldRepr := strings.Join(w.fieldChain, ".")
	return fmt.Sprintf("field '%s' %s", fieldRepr, w.rule)
}

type ConvertingWrapper struct {
	rule      Rule
	converter EntityConverter
}

func (w ConvertingWrapper) Apply(ctx *common.ProxyContext, v interface{}) (bool, error) {
	e, ok := v.(wrapper.Entity)
	if !ok {
		return false, fmt.Errorf("not an entity: %v", v)
	}
	data, err := w.converter.Convert(e)
	if err != nil {
		return false, fmt.Errorf("converting with %s: %w", w.converter, err)
	}
	res, err := w.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("in rule %v: %w", w.rule, err)
	}
	return res, nil
}

func (w ConvertingWrapper) String() string {
	return fmt.Sprintf("%s %s", w.converter, w.rule)
}
