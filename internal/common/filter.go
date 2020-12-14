package common

import "fmt"

type Rule interface {
	fmt.Stringer
}

type Filter interface {
	GetRule() Rule
	GetVerdict() Verdict
	IsEnabled() bool
	SetEnabled(enabled bool)
}
