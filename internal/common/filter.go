package common

import "fmt"

type Rule interface {
	fmt.Stringer
}

type Filter interface {
	GetRule() Rule
	GetVerdict() Verdict
	GetAlert() bool
	IsEnabled() bool
	SetEnabled(enabled bool)
	SetAlert(alert bool)
}
