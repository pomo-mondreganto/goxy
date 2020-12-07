package filters

import (
	"fmt"
	"goxy/internal/common"
	"strings"
)

func NewAnyWrapper(r RawRule, _ *common.RuleConfig) RawRule {
	return &AnyWrapper{r}
}

func NewArrayWrapper(r RawRule, _ *common.RuleConfig) RawRule {
	return &ArrayWrapper{r}
}

func NewFieldWrapper(r RawRule, cfg *common.RuleConfig) RawRule {
	fieldChain := strings.Split(cfg.Field, ".")
	return &FieldWrapper{r, fieldChain}
}

type AnyWrapper struct {
	rule RawRule
}

func (w *AnyWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
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
	default:
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}

	return false, nil
}

type ArrayWrapper struct {
	rule RawRule
}

func (w *ArrayWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	if _, ok := data.([]interface{}); !ok {
		return false, fmt.Errorf("data type %T: %w", data, ErrInvalidInputType)
	}
	res, err := w.rule.Apply(ctx, data)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}

type FieldWrapper struct {
	rule       RawRule
	fieldChain []string
}

func (w *FieldWrapper) Apply(ctx *common.ProxyContext, data interface{}) (bool, error) {
	result := data
	for _, f := range w.fieldChain {
		converted, ok := result.(map[string]interface{})
		if !ok {
			//return false, fmt.Errorf("no key %s: not a map: %w", f, ErrInvalidInputType)
			return false, nil
		}
		next, ok := converted[f]
		if !ok {
			//return false, fmt.Errorf("no key %s: %w", f, ErrInvalidInputType)
			return false, nil
		}
		result = next
	}
	res, err := w.rule.Apply(ctx, result)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}
