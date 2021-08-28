package common

import "time"

type RuleConfig struct {
	Name  string   `json:"name" mapstructure:"name"`
	Type  string   `json:"type" mapstructure:"type"`
	Field string   `json:"field" mapstructure:"field"`
	Args  []string `json:"args" mapstructure:"args"`
}

type FilterConfig struct {
	Rule    string `json:"rule" mapstructure:"rule"`
	Alert   bool   `json:"alert" mapstructure:"alert"`
	Verdict string `json:"verdict" mapstructure:"verdict"`
}

type ServiceConfig struct {
	Name           string         `json:"name" mapstructure:"name"`
	Type           string         `json:"type" mapstructure:"type"`
	Listen         string         `json:"listen" mapstructure:"listen"`
	Target         string         `json:"target" mapstructure:"target"`
	RequestTimeout *time.Duration `json:"request_timeout" mapstructure:"request_timeout"`
	Filters        []FilterConfig `json:"filters" mapstructure:"filters"`
}

type ProxyConfig struct {
	Rules    []RuleConfig    `json:"rules" mapstructure:"rules"`
	Services []ServiceConfig `json:"services" mapstructure:"services"`
}
