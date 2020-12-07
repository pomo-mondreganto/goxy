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

var DefaultCompositeRuleCreators = map[string]CompositeRuleCreator{
	"and": NewCompositeAndRule,
	"not": NewCompositeNotRule,
}

var DefaultEntityConverters = map[string]EntityConverter{
	"json":   JSONEntityConverter,
	"cookie": CookiesEntityConverter,
	"query":  QueryEntityConverter,
	"body":   BodyEntityConverter,
	"path":   PathEntityConverter,
	"form":   FormEntityConverter,
}

var DefaultRawRuleCreators = map[string]RawRuleCreator{
	"contains": NewContainsStringRawRule,
}

var DefaultRawRuleWrappers = map[string]RawRuleWrapperCreator{
	"any":   NewAnyWrapper,
	"array": NewArrayWrapper,
}
