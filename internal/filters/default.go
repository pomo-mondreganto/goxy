package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
	"egress":  &CompositeNotRule{new(IngressRule)},
}

var DefaultRuleWrappers = map[string]RuleWrapperCreator{
	"ingress": NewIngressWrapper,
	"egress":  NewEgressWrapper,
	"not":     NewNotWrapper,
	"any":     NewAnyWrapper,
	"field":   NewFieldWrapper,
}

var DefaultRuleCreators = map[string]RuleCreator{
	"and":       NewCompositeAndRule,
	"not":       NewCompositeNotRule,
	"contains":  NewContainsRule,
	"icontains": NewIContainsRule,
	"regex":     NewRegexRule,
}

var DefaultEntityConverters = map[string]EntityConverter{
	"json":    JsonEntityConverter{},
	"cookies": CookiesEntityConverter{},
	"query":   QueryEntityConverter{},
	"body":    BodyEntityConverter{},
	"path":    PathEntityConverter{},
	"form":    FormEntityConverter{},
	"headers": HeadersEntityConverter{},
}
