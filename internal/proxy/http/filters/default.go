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
	"json":    JSONEntityConverter,
	"cookies": CookiesEntityConverter,
	"query":   QueryEntityConverter,
	"body":    BodyEntityConverter,
	"path":    PathEntityConverter,
	"form":    FormEntityConverter,
	"headers": HeadersEntityConverter,
}

var DefaultRawRuleCreators = map[string]RawRuleCreator{
	"contains":  NewContainsRawRule,
	"icontains": NewIContainsRawRule,
	"regex":     NewRegexRawRule,
}

var DefaultRawRuleWrappers = map[string]RawRuleWrapperCreator{
	"any":   NewAnyWrapper,
	"array": NewArrayWrapper,
}
