package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"goxy/internal/proxy"
	"net/http"
	"os"
	"path"
)

func NewServer(pm *proxy.Manager) *Server {
	ms := &Server{
		ProxyManager: pm,
		Router:       gin.New(),
		StaticDir:    viper.GetString("web.static_dir"),
		AuthData: &BasicAuthData{
			Username: viper.GetString("web.username"),
			Password: viper.GetString("web.password"),
		},
	}
	ms.registerMiddleware()
	ms.registerRoutes()
	return ms
}

type BasicAuthData struct {
	Username string
	Password string
}

type Server struct {
	ProxyManager *proxy.Manager
	Router       *gin.Engine
	StaticDir    string
	AuthData     *BasicAuthData
}

func (s *Server) registerRoutes() {
	api := s.Router.Group("/api")
	{
		api.GET("/status/", s.statusHandler())
		api.GET("/proxies/", s.proxyListingHandler())
		api.PUT("/proxies/:id/listening/", s.setProxyListening())
		api.PUT("/proxies/:id/filter_enabled/", s.setFilterEnabled())
	}

	logrus.Infof("Serving static dir: %s", s.StaticDir)

	s.Router.NoRoute(func(c *gin.Context) {
		realPath := path.Join(s.StaticDir, c.Request.URL.Path)
		if _, err := os.Stat(realPath); os.IsNotExist(err) {
			realPath = path.Join(s.StaticDir, "index.html")
		}
		c.File(realPath)
	})

	logrus.Info("Routes registered successfully")
}

func (s *Server) registerMiddleware() {
	s.Router.Use(gin.Recovery())
	s.Router.Use(loggerMiddleware())

	s.Router.Use(gzip.Gzip(gzip.DefaultCompression))
	s.Router.Use(cors.Default())

	if !gin.IsDebugging() {
		s.Router.Use(gin.BasicAuth(gin.Accounts{
			s.AuthData.Username: s.AuthData.Password,
		}))
	}

	logrus.Info("Middleware registered successfully")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
