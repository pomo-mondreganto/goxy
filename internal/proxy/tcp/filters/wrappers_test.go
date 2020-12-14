package filters

import (
	"goxy/internal/common"
	"testing"
)

type trueRule struct {
	calls int
}

func (r *trueRule) Apply(_ *common.ProxyContext, _ []byte, _ bool) (bool, error) {
	r.calls += 1
	return true, nil
}

func (r trueRule) String() string {
	return "true"
}

func TestEgressWrapper_Apply(t *testing.T) {
	tests := []struct {
		name    string
		rule    *trueRule
		ingress bool
		want    bool
	}{
		{
			"ingress packet",
			&trueRule{},
			true,
			false,
		},
		{
			"egress packet",
			&trueRule{},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &EgressWrapper{
				rule: tt.rule,
			}
			got, err := w.Apply(common.NewProxyContext(), nil, tt.ingress)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
			if !got && tt.rule.calls > 0 {
				t.Errorf("Apply() inner rule called, but mustn't be")
			}
		})
	}
}

func TestIngressWrapper_Apply(t *testing.T) {
	tests := []struct {
		name    string
		rule    *trueRule
		ingress bool
		want    bool
	}{
		{
			"ingress packet",
			&trueRule{},
			true,
			true,
		},
		{
			"egress packet",
			&trueRule{},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &IngressWrapper{
				rule: tt.rule,
			}
			got, err := w.Apply(common.NewProxyContext(), nil, tt.ingress)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
			if !got && tt.rule.calls > 0 {
				t.Errorf("Apply() inner rule called, but mustn't be")
			}
		})
	}
}

func TestNotWrapper_Apply(t *testing.T) {
	tests := []struct {
		name    string
		rule    *trueRule
		ingress bool
		want    bool
	}{
		{
			"true packet",
			&trueRule{},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &NotWrapper{
				rule: tt.rule,
			}
			got, err := w.Apply(common.NewProxyContext(), nil, tt.ingress)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
			if tt.rule.calls == 0 {
				t.Errorf("Apply() inner rule not called, but must be")
			}
		})
	}
}
