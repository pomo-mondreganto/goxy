package common

type RuleConfig struct {
	Name  string
	Type  string
	Field string
	Args  []string
}

type FilterConfig struct {
	Rule    string
	Verdict string
}

type ServiceConfig struct {
	Name    string
	Type    string
	Listen  string
	Target  string
	Filters []*FilterConfig
}

type ProxyConfig struct {
	Rules    []*RuleConfig
	Services []*ServiceConfig
}
