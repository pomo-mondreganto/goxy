package models

import "goxy/internal/common"

type ProxyDescription struct {
	ID        int                   `json:"id"`
	Service   *common.ServiceConfig `json:"service"`
	Listening bool                  `json:"listening"`
}
