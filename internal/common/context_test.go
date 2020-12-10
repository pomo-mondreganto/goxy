package common

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
	"testing"
)

func TestProxyContext_AddToCounter(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value int
	}{
		{
			"positive",
			"key",
			5,
		},
		{
			"negative",
			"key",
			-5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewProxyContext()
			c.AddToCounter(tt.key, tt.value)
			if c.counters[tt.key] != tt.value {
				t.Errorf("AddToCounter(): invalid counters: want %d, got %d", tt.value, c.counters[tt.key])
			}
		})
	}
}

func TestProxyContext_DumpFields(t *testing.T) {
	type fields struct {
		flags    map[string]bool
		counters map[string]int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{"test all",
			fields{
				flags: map[string]bool{
					"some flag":  true,
					"other flag": false,
				},
				counters: map[string]int{
					"counter1": 0,
					"counter2": 5,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ProxyContext{
				flags:    tt.fields.flags,
				counters: tt.fields.counters,
				mu:       sync.RWMutex{},
			}
			wantFields := make(logrus.Fields)
			for k, v := range tt.fields.counters {
				wantFields[k] = v
			}
			for k, v := range tt.fields.flags {
				if v {
					wantFields[k] = v
				}
			}
			if got := c.DumpFields(); !reflect.DeepEqual(got, wantFields) {
				t.Errorf("DumpFields() = %v, want %v", got, wantFields)
			}
		})
	}
}

func TestProxyContext_GetCounter(t *testing.T) {
	type fields struct {
		flags    map[string]bool
		counters map[string]int
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			"get existing",
			fields{
				flags:    map[string]bool{},
				counters: map[string]int{"counter1": 123},
			},
			args{"counter1"},
			123,
		},
		{
			"get missing",
			fields{
				flags:    map[string]bool{},
				counters: map[string]int{},
			},
			args{"counter1"},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ProxyContext{
				flags:    tt.fields.flags,
				counters: tt.fields.counters,
				mu:       sync.RWMutex{},
			}
			if got := c.GetCounter(tt.args.key); got != tt.want {
				t.Errorf("GetCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyContext_GetFlag(t *testing.T) {
	type fields struct {
		flags    map[string]bool
		counters map[string]int
	}
	type args struct {
		flag string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"get existing true",
			fields{
				flags:    map[string]bool{"flag1": true},
				counters: map[string]int{},
			},
			args{"flag1"},
			true,
		},
		{
			"get existing false",
			fields{
				flags:    map[string]bool{"flag1": false},
				counters: map[string]int{},
			},
			args{"flag1"},
			false,
		},
		{
			"get missing",
			fields{
				flags:    map[string]bool{},
				counters: map[string]int{},
			},
			args{"flag1"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ProxyContext{
				flags:    tt.fields.flags,
				counters: tt.fields.counters,
				mu:       sync.RWMutex{},
			}
			if got := c.GetFlag(tt.args.flag); got != tt.want {
				t.Errorf("GetFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxyContext_SetFlag(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			"set true",
			"flag1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewProxyContext()
			c.SetFlag(tt.key)
			if _, ok := c.flags[tt.key]; !ok {
				t.Errorf("SetFlag(): flag not set")
			}
		})
	}
}
