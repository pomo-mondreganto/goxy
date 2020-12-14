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
	SetFilterEnabled(filter int, enabled bool) error
	GetFilters() []common.Filter

	fmt.Stringer
}
