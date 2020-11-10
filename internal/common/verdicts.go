package common

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type Verdict interface {
	Mutate(ctx *ConnectionContext) error
}

func ParseVerdict(desc string) (Verdict, error) {
	tokens := strings.Split(desc, "::")
	switch strings.ToLower(tokens[0]) {
	case "drop":
		return &VerdictDrop{}, nil
	case "accept":
		return &VerdictAccept{}, nil
	case "inc":
		if len(tokens) < 2 {
			return nil, errors.New("counter missing for inc verdict")
		}
		return &VerdictIncrement{Key: tokens[1]}, nil
	case "dec":
		if len(tokens) < 2 {
			return nil, errors.New("counter missing for dec verdict")
		}
		return &VerdictDecrement{Key: tokens[1]}, nil
	case "alert":
		if len(tokens) < 2 {
			return nil, errors.New("reason missing for alert verdict")
		}
		v := &VerdictAlert{
			Logger: logrus.WithField("reason", tokens[1]),
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown verdict: %s", tokens[0])
	}
}

type VerdictDrop struct{}

func (v *VerdictDrop) Mutate(ctx *ConnectionContext) error {
	ctx.MustDrop = true
	return nil
}

type VerdictAccept struct{}

func (v *VerdictAccept) Mutate(ctx *ConnectionContext) error {
	ctx.MustAccept = true
	return nil
}

type VerdictIncrement struct {
	Key string
}

func (v *VerdictIncrement) Mutate(ctx *ConnectionContext) error {
	ctx.Counters[v.Key] += 1
	return nil
}

type VerdictDecrement struct {
	Key string
}

func (v *VerdictDecrement) Mutate(ctx *ConnectionContext) error {
	ctx.Counters[v.Key] -= 1
	return nil
}

type VerdictAlert struct {
	Logger *logrus.Entry
}

func (v *VerdictAlert) Mutate(ctx *ConnectionContext) error {
	v.Logger.WithFields(ctx.DumpFields()).Warningf("Alert triggered")
	return nil
}
