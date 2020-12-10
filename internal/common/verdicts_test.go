package common

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestParseVerdict(t *testing.T) {
	type args struct {
		desc string
	}
	tests := []struct {
		name    string
		args    args
		want    Verdict
		wantErr bool
	}{
		{
			"alert",
			args{"alert::some message"},
			&VerdictAlert{
				logrus.WithField("reason", "some message"),
			},
			false,
		},
		{
			"increment",
			args{"inc::test"},
			&VerdictIncrement{Key: "test"},
			false,
		},
		{
			"increment with space",
			args{"inc::test something"},
			&VerdictIncrement{Key: "test something"},
			false,
		},
		{
			"decrement",
			args{"dec::test"},
			&VerdictDecrement{Key: "test"},
			false,
		},
		{
			"decrement with space",
			args{"dec::test something"},
			&VerdictDecrement{Key: "test something"},
			false,
		},
		{
			"accept",
			args{"accept"},
			&VerdictSetFlag{Key: AcceptFlag},
			false,
		},
		{
			"drop",
			args{"drop"},
			&VerdictSetFlag{Key: DropFlag},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVerdict(tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVerdict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVerdict() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestVerdictAlert_Mutate(t *testing.T) {
	type fields struct {
		Logger *logrus.Entry
	}
	type args struct {
		ctx *ProxyContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"test no error",
			fields{Logger: logrus.StandardLogger().WithFields(nil)},
			args{ctx: NewProxyContext()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VerdictAlert{
				Logger: tt.fields.Logger,
			}
			if err := v.Mutate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerdictDecrement_Mutate(t *testing.T) {
	type fields struct {
		Key string
	}
	type args struct {
		ctx *ProxyContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"simple increment",
			fields{"counter"},
			args{NewProxyContext()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VerdictDecrement{
				Key: tt.fields.Key,
			}
			if err := v.Mutate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}
			cntVal := tt.args.ctx.GetCounter(tt.fields.Key)
			if cntVal != -1 {
				t.Errorf("Mutate() want counter %d, got %d", cntVal, -1)
			}
		})
	}
}

func TestVerdictIncrement_Mutate(t *testing.T) {
	type fields struct {
		Key string
	}
	type args struct {
		ctx *ProxyContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"simple increment",
			fields{"counter"},
			args{NewProxyContext()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VerdictIncrement{
				Key: tt.fields.Key,
			}
			if err := v.Mutate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}
			cntVal := tt.args.ctx.GetCounter(tt.fields.Key)
			if cntVal != 1 {
				t.Errorf("Mutate() want counter %d, got %d", cntVal, 1)
			}
		})
	}
}

func TestVerdictSetFlag_Mutate(t *testing.T) {
	type fields struct {
		Key string
	}
	type args struct {
		ctx *ProxyContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"simple increment",
			fields{"counter"},
			args{NewProxyContext()},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VerdictSetFlag{
				Key: tt.fields.Key,
			}
			if err := v.Mutate(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mutate() error = %v, wantErr %v", err, tt.wantErr)
			}
			cntVal := tt.args.ctx.GetFlag(tt.fields.Key)
			if !cntVal {
				t.Errorf("Mutate() want counter %t, got %t", cntVal, false)
			}
		})
	}
}
