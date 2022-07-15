package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DropFlag   = "drop"
	AcceptFlag = "accept"
)

type Verdict interface {
	Mutate(ctx *ProxyContext) error
	fmt.Stringer
}

func ParseVerdict(desc string) (Verdict, error) {
	if desc == "" {
		return VerdictDummy{}, nil
	}
	tokens := strings.Split(desc, "::")
	switch strings.ToLower(tokens[0]) {
	case "drop":
		v := VerdictSetFlag{DropFlag}
		return v, nil
	case "accept":
		v := VerdictSetFlag{AcceptFlag}
		return v, nil
	case "inc":
		if len(tokens) < 2 {
			return nil, errors.New("counter missing for inc verdict")
		}
		return VerdictIncrement{Key: tokens[1]}, nil
	case "dec":
		if len(tokens) < 2 {
			return nil, errors.New("counter missing for dec verdict")
		}
		return VerdictDecrement{Key: tokens[1]}, nil
	case "alert":
		if len(tokens) < 2 {
			return nil, errors.New("reason missing for alert verdict")
		}
		v := VerdictAlert{
			Logger: logrus.WithField("reason", tokens[1]),
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown verdict: %s", tokens[0])
	}
}

type VerdictSetFlag struct {
	Key string
}

func (v VerdictSetFlag) Mutate(ctx *ProxyContext) error {
	ctx.SetFlag(v.Key)
	return nil
}

func (v VerdictSetFlag) String() string {
	return fmt.Sprintf("set '%s'", v.Key)
}

type VerdictIncrement struct {
	Key string
}

func (v VerdictIncrement) Mutate(ctx *ProxyContext) error {
	ctx.AddToCounter(v.Key, 1)
	return nil
}

func (v VerdictIncrement) String() string {
	return fmt.Sprintf("inc '%s'", v.Key)
}

type VerdictDecrement struct {
	Key string
}

func (v VerdictDecrement) Mutate(ctx *ProxyContext) error {
	ctx.AddToCounter(v.Key, -1)
	return nil
}

func (v VerdictDecrement) String() string {
	return fmt.Sprintf("dec '%s'", v.Key)
}

type VerdictAlert struct {
	Logger *logrus.Entry
}

func (v VerdictAlert) Mutate(ctx *ProxyContext) error {
	v.Logger.WithFields(ctx.DumpFields()).Warningf("Alert triggered")
	return nil
}

func (v VerdictAlert) String() string {
	return "alert"
}

type VerdictDummy struct{}

func (v VerdictDummy) Mutate(*ProxyContext) error {
	return nil
}

func (v VerdictDummy) String() string {
	return "dummy"
}
