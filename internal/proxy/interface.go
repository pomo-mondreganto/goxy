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

	fmt.Stringer
}
