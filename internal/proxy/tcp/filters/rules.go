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

func NewVolgaCTFRule(_ RuleSet, cfg common.RuleConfig) (Rule, error) {
	if len(cfg.Args) != 1 {
		return nil, ErrInvalidRuleArgs
	}
	defaultAlp := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.")
	r := VolgaModifyRule{
		alpOriginal: defaultAlp,
		alpShifted:  []byte(cfg.Args[0]),
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

type VolgaModifyRule struct {
	alpOriginal []byte
	alpShifted  []byte
}

func findByteInSlice(buf []byte, value byte) int {
	for i, elem := range buf {
		if value == elem {
			return i
		}
	}
	return -1
}

func (r VolgaModifyRule) Apply(_ *common.ProxyContext, buf []byte, ingress bool) (bool, error) {
	alpFrom := r.alpOriginal
	alpTo := r.alpShifted
	if ingress {
		alpFrom, alpTo = alpTo, alpFrom
	}
	prefix := `VolgaCTF{`
	suffix := `}`
	VolgaCTFFlagRegex := regexp.MustCompile(prefix + `[\w-]*\.[\w-]*\.[\w-]*` + suffix)
	allIndices := VolgaCTFFlagRegex.FindAllIndex(buf, -1)
	for _, indices := range allIndices {
		start, end := indices[0]+len(prefix), indices[1]-len(suffix)
		for i := start; i < end; i++ {
			alpIndex := findByteInSlice(alpFrom, buf[i])
			if alpIndex == -1 {
				continue
			}
			buf[i] = alpTo[alpIndex]
		}
	}
	return len(allIndices) != 0, nil
}

func (r VolgaModifyRule) String() string {
	return "VolgaCTF flag editor"
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
