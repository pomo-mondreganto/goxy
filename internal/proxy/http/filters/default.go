package filters

var DefaultRules = map[string]Rule{
	"ingress": new(IngressRule),
	"egress":  &CompositeNotRule{new(IngressRule)},
}

var DefaultTransformers = map[string]Rule{
	"transform_volga": NewTransformVolgaCTF(),
}

var DefaultRuleWrappers = map[string]RuleWrapperCreator{
	"ingress": NewIngressWrapper,
	"egress":  NewEgressWrapper,
	"not":     NewNotWrapper,
}

var DefaultRuleCreators = map[string]RuleCreator{
	"and": NewCompositeAndRule,
	"not": NewCompositeNotRule,
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

var DefaultRawRuleCreators = map[string]RawRuleCreator{
	"contains":  NewContainsRawRule,
	"icontains": NewIContainsRawRule,
	"regex":     NewRegexRawRule,
}

var DefaultRawRuleWrappers = map[string]RawRuleWrapperCreator{
	"any":   NewAnyWrapper,
	"array": NewArrayWrapper,
	"not":   NewNotWrapperRaw,
}
