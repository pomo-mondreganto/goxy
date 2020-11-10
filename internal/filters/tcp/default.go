package tcp

var DefaultRules = map[string]Rule{
	"ingress": &IngressRule{},
	"egress":  &CompositeNotRule{Rule: &IngressRule{}},
}

var DefaultRuleCreators = map[string]RuleCreator{
	"regex":    NewRegexRule,
	"contains": NewContainsRule,
}

var DefaultCompositeRuleCreators = map[string]CompositeRuleCreator{
	"and": NewCompositeAndRule,
	"not": NewCompositeNotRule,
}
