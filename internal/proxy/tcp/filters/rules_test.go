package filters

import (
	"goxy/internal/common"
	"regexp"
	"testing"
)

func TestContainsRule_Apply(t *testing.T) {
	type fields struct {
		value []byte
	}
	tests := []struct {
		name   string
		fields fields
		data   []byte
		want   bool
	}{
		{
			"simple_contains",
			fields{[]byte("test")},
			[]byte("some test data"),
			true,
		},
		{
			"simple_not_contains",
			fields{[]byte("test")},
			[]byte("some tst data"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ContainsRule{
				value: tt.fields.value,
			}
			got, err := r.Apply(common.NewProxyContext(), tt.data, false)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterGTRule_Apply(t *testing.T) {
	type fields struct {
		Key   string
		Value int
	}
	type contextFields struct {
		Key   string
		Value int
	}
	tests := []struct {
		name          string
		fields        fields
		data          []byte
		contextFields contextFields
		want          bool
	}{
		{
			"uninitialized false",
			fields{
				"key",
				0,
			},
			[]byte("anything"),
			contextFields{
				Key:   "",
				Value: 0,
			},
			false,
		},
		{
			"uninitialized true",
			fields{
				"key",
				-1,
			},
			[]byte("anything"),
			contextFields{
				Key:   "",
				Value: 0,
			},
			true,
		},
		{
			"simple true",
			fields{
				"key",
				0,
			},
			[]byte("anything"),
			contextFields{
				Key:   "key",
				Value: 1,
			},
			true,
		},
		{
			"simple false",
			fields{
				"key",
				5,
			},
			[]byte("anything"),
			contextFields{
				Key:   "key",
				Value: 5,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := common.NewProxyContext()
			if tt.contextFields.Key != "" {
				ctx.AddToCounter(tt.contextFields.Key, tt.contextFields.Value)
			}
			r := &CounterGTRule{
				key:   tt.fields.Key,
				value: tt.fields.Value,
			}
			got, err := r.Apply(ctx, tt.data, false)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIngressRule_Apply(t *testing.T) {
	tests := []struct {
		name    string
		ingress bool
		data    []byte
	}{
		{
			"empty ingress",
			true,
			[]byte(""),
		},
		{
			"empty egress",
			false,
			[]byte(""),
		},
		{
			"ingress with data",
			true,
			[]byte("some data"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &IngressRule{}
			got, err := r.Apply(common.NewProxyContext(), tt.data, tt.ingress)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.ingress {
				t.Errorf("Apply() got = %v, want %v", got, tt.ingress)
			}
		})
	}
}

func TestRegexRule_Apply(t *testing.T) {
	tests := []struct {
		name  string
		Regex *regexp.Regexp
		data  []byte
		want  bool
	}{
		{
			"simple_contains",
			regexp.MustCompile("test"),
			[]byte("some test data"),
			true,
		},
		{
			"simple_not_contains",
			regexp.MustCompile("test"),
			[]byte("some tst data"),
			false,
		},
		{
			"full match",
			regexp.MustCompile("^[A-Z0-9]{5}=$"),
			[]byte("ABCZ7="),
			true,
		},
		{
			"flag urlencode",
			regexp.MustCompile("[A-Z0-9]{5}(=|%3d|%3D)"),
			[]byte("ABCZ7%3d"),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexRule{
				regex: tt.Regex,
			}
			got, err := r.Apply(common.NewProxyContext(), tt.data, false)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIContainsRule_Apply(t *testing.T) {
	type fields struct {
		value []byte
	}
	tests := []struct {
		name   string
		fields fields
		data   []byte
		want   bool
	}{
		{
			"simple_contains",
			fields{[]byte("test")},
			[]byte("some tEsT data"),
			true,
		},
		{
			"simple_not_contains",
			fields{[]byte("test")},
			[]byte("some tst data"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &IContainsRule{
				value: tt.fields.value,
			}
			got, err := r.Apply(common.NewProxyContext(), tt.data, false)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Apply() got = %v, want %v", got, tt.want)
			}
		})
	}
}
