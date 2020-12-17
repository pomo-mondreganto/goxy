package filters

import (
	"bytes"
	"errors"
	"fmt"
	"goxy/internal/common"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidRuleArgs = errors.New("invalid rule arguments")
)

func NewIngressRule(_ RuleSet, _ common.RuleConfig) (Rule, error) {
	return IngressRule{}, nil
}

func NewRegexRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r, err := regexp.Compile(cfg.Args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	return RegexRule{regex: r}, nil
}

func NewContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r := ContainsRule{value: []byte(cfg.Args[0])}
	return r, nil
}

func NewIContainsRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	r := IContainsRule{value: []byte(strings.ToLower(cfg.Args[0]))}
	return r, nil
}

func NewCounterGTRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 2 {
		return nil, ErrInvalidRuleArgs
	}
	val, err := strconv.Atoi(cfg.Args[1])
	if err != nil {
		return nil, fmt.Errorf("parsing value: %w", err)
	}
	r := CounterGTRule{
		key:   cfg.Args[0],
		value: val,
	}
	return r, nil
}

type IngressRule struct{}

func (r IngressRule) Apply(_ *common.ProxyContext, _ []byte, ingress bool) (bool, error) {
	return ingress, nil
}

func (r IngressRule) String() string {
	return "ingress"
}

type RegexRule struct {
	regex *regexp.Regexp
}

func (r RegexRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	return r.regex.Match(buf), nil
}

func (r RegexRule) String() string {
	return fmt.Sprintf("regex '%s'", r.regex)
}

type ContainsRule struct {
	value []byte
}

func (r ContainsRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	return bytes.Contains(buf, r.value), nil
}

func (r ContainsRule) String() string {
	return fmt.Sprintf("contains '%s'", string(r.value))
}

type IContainsRule struct {
	value []byte
}

func (r IContainsRule) Apply(_ *common.ProxyContext, buf []byte, _ bool) (bool, error) {
	return bytes.Contains(bytes.ToLower(buf), r.value), nil
}

func (r IContainsRule) String() string {
	return fmt.Sprintf("icontains '%s'", string(r.value))
}

type CounterGTRule struct {
	key   string
	value int
}

func (r CounterGTRule) Apply(ctx *common.ProxyContext, _ []byte, _ bool) (bool, error) {
	return ctx.GetCounter(r.key) > r.value, nil
}

func (r CounterGTRule) String() string {
	return fmt.Sprintf("counter '%s' > %d", r.key, r.value)
}
