package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
	"egress":  &CompositeNotRule{new(IngressRule)},
}

var DefaultRuleFactories = map[string]RuleFactory{
	"ingress": IngressRuleFactory,
	"egress":  EgressRuleFactory,
	"not":     NotRuleFactory,
}

var DefaultRuleCreators = map[string]RuleCreator{
	"ingress":    NewIngressRule,
	"regex":      NewRegexRule,
	"contains":   NewContainsRule,
	"counter_gt": NewCounterGTRule,
}

var DefaultCompositeRuleCreators = map[string]CompositeRuleCreator{
	"and": NewCompositeAndRule,
	"not": NewCompositeNotRule,
}
