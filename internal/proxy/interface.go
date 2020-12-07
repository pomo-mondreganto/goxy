package proxy

import (
	"context"
	"goxy/internal/common"
)

type Proxy interface {
	Start() error
	Shutdown(ctx context.Context) error
	GetConfig() *common.ServiceConfig
}
