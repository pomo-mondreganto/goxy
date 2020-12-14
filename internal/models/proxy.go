package models

import "goxy/internal/common"

type FilterDescription struct {
	ID      int    `json:"id"`
	ProxyID int    `json:"proxy_id"`
	Rule    string `json:"rule"`
	Verdict string `json:"verdict"`
	Enabled bool   `json:"enabled"`
}

type ProxyDescription struct {
	ID                 int                   `json:"id"`
	Service            *common.ServiceConfig `json:"service"`
	Listening          bool                  `json:"listening"`
	FilterDescriptions []FilterDescription   `json:"filter_descriptions"`
}
