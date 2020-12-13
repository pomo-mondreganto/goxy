package filters

import (
	"fmt"
	"goxy/internal/common"
)

func NewIngressWrapper(rule Rule, _ common.RuleConfig) Rule {
	return &IngressWrapper{rule}
}

func NewEgressWrapper(rule Rule, _ common.RuleConfig) Rule {
	return &EgressWrapper{rule}
}

func NewNotWrapper(rule Rule, _ common.RuleConfig) Rule {
	return &NotWrapper{rule}
}

type IngressWrapper struct {
	rule Rule
}

func (w *IngressWrapper) Apply(ctx *common.ProxyContext, buf []byte, ingress bool) (bool, error) {
	if !ingress {
		return false, nil
	}
	res, err := w.rule.Apply(ctx, buf, ingress)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}

type EgressWrapper struct {
	rule Rule
}

func (w *EgressWrapper) Apply(ctx *common.ProxyContext, buf []byte, ingress bool) (bool, error) {
	if ingress {
		return false, nil
	}
	res, err := w.rule.Apply(ctx, buf, ingress)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return res, nil
}

type NotWrapper struct {
	rule Rule
}

func (w *NotWrapper) Apply(ctx *common.ProxyContext, buf []byte, ingress bool) (bool, error) {
	res, err := w.rule.Apply(ctx, buf, ingress)
	if err != nil {
		return false, fmt.Errorf("error in rule %T: %w", w.rule, err)
	}
	return !res, nil
}
