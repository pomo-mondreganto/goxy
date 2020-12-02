package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
	"egress":  &CompositeNotRule{Rule: new(IngressRule)},
}

var DefaultRuleCreators = map[string]RuleCreator{
	"regex":      NewRegexRule,
	"contains":   NewContainsRule,
	"counter_gt": NewCounterGTRule,

	"not_contains": NotRuleFactory(NewContainsRule),
	"not_regex":    NotRuleFactory(NewRegexRule),

	"ingress_contains": IngressRuleFactory(NewContainsRule),
	"ingress_regex":    IngressRuleFactory(NewRegexRule),
}

var DefaultCompositeRuleCreators = map[string]CompositeRuleCreator{
	"and": NewCompositeAndRule,
	"not": NewCompositeNotRule,
}
