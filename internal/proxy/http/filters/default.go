package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
}
