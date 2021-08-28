package proxy

import (
	"context"
	"fmt"
	"goxy/internal/common"
)

type Proxy interface {
	Start() error
	Shutdown(ctx context.Context) error
	GetConfig() *common.ServiceConfig
	GetListening() bool
	SetListening(state bool)
	SetFilterState(filter int, enabled, alert bool) error
	GetFilters() []common.Filter

	fmt.Stringer
}
