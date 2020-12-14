package web

type ModelDetailRequest struct {
	ID int `uri:"id" binding:"required"`
}

type SetProxyListeningRequest struct {
	Listening bool `json:"listening"`
}

type SetFilterListeningRequest struct {
	ID      int  `json:"id" binding:"required"`
	Enabled bool `json:"enabled"`
}
