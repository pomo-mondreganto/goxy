package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
	"egress":  &CompositeNotRule{new(IngressRule)},
}

var DefaultRuleWrappers = map[string]RuleWrapperCreator{
	"ingress": NewIngressWrapper,
	"egress":  NewEgressWrapper,
	"not":     NewNotWrapper,
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
