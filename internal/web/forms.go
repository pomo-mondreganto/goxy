package web

type ModelDetailRequest struct {
	ID int `uri:"id" binding:"required"`
}

type ProxyFilterDetailRequest struct {
	ID       int `uri:"id" binding:"required"`
	FilterID int `uri:"filter_id" binding:"required"`
}

type SetProxyListeningRequest struct {
	Listening bool `json:"listening"`
}

type UpdateFilterStateRequest struct {
	Enabled bool `json:"enabled"`
	Alert   bool `json:"alert"`
}
