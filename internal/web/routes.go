package web

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s Server) statusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (s Server) proxyListingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxies := s.ProxyManager.DumpProxies()
		c.JSON(http.StatusOK, gin.H{"proxies": proxies})
	}
}

func (s Server) setProxyListening() gin.HandlerFunc {
	return func(c *gin.Context) {
		idReq := new(ModelDetailRequest)
		if err := c.ShouldBindUri(idReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dataReq := new(SetProxyListeningRequest)
		if err := c.ShouldBindJSON(dataReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.ProxyManager.SetProxyListening(idReq.ID, dataReq.Listening); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func (s Server) updateFilterState() gin.HandlerFunc {
	return func(c *gin.Context) {
		detReq := new(ProxyFilterDetailRequest)
		if err := c.ShouldBindUri(detReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		dataReq := new(UpdateFilterStateRequest)
		if err := c.ShouldBindJSON(dataReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.ProxyManager.SetFilterState(detReq.ID, detReq.FilterID, dataReq.Enabled, dataReq.Alert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
