package filters

import (
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
)

type IngressRule struct{}

func (r *IngressRule) Apply(_ *common.ProxyContext, e wrapper.Entity) (bool, error) {
	return e.GetIngress(), nil
}
