package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"goxy/internal/common"
	"goxy/internal/proxy"
	"goxy/internal/web"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	configFile = flag.String("config", "config.yml", "Path to the config file in YAML format")
	logLevel   = flag.String("log_level", "INFO", "Log level {INFO|DEBUG|WARNING|ERROR}")
)

func main() {
	flag.Parse()

	setLogLevel()
	setWebServerMode()
	setConfigDefaults()
	parseConfig()

	cfg := parseProxyConfig()
	m := runProxyManager(cfg)

	s := web.NewServer(m)
	httpServer := startHttpServer(s)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	shutdownHttpServer(httpServer, ctx)
	shutdownProxyManager(m, ctx)

	logrus.Info("Shutdown successful")
}

func setLogLevel() {
	switch strings.ToUpper(*logLevel) {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARNING":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.Errorf("Invalid log level provided: %s", *logLevel)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func setWebServerMode() {
	level := logrus.StandardLogger().GetLevel()
	if level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func setConfigDefaults() {
	viper.SetDefault("web.static_dir", "front/dist")
	viper.SetDefault("web.username", "admin")
	viper.SetDefault("web.listen", "0.0.0.0:8000")
}

func parseConfig() {
	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("Error reading config from yaml: ", err)
	}
}

func parseProxyConfig() *common.ProxyConfig {
	cfg := new(common.ProxyConfig)
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Fatal("Error parsing proxy config from file: ", err)
	}
	return cfg
}

func runProxyManager(cfg *common.ProxyConfig) *proxy.Manager {
	m, err := proxy.NewManager(cfg)
	if err != nil {
		logrus.Fatalf("Error creating proxy manager: %v", err)
	}
	if err := m.StartAll(); err != nil {
		logrus.Fatalf("Error starting proxy manager: %v", err)
	}
	return m
}

func startHttpServer(s *web.Server) *http.Server {
	srv := &http.Server{
		Addr:         viper.GetString("web.listen"),
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 30,
		Handler:      s,
	}

	go func() {
		logrus.Infof("Serving on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Error serving api server: ", err)
		}
	}()

	return srv
}

func shutdownHttpServer(srv *http.Server, ctx context.Context) {
	logrus.Info("Shutting down http server")
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error shutting down http server: %v", err)
	}
}

func shutdownProxyManager(m *proxy.Manager, ctx context.Context) {
	logrus.Info("Shutting down proxies")
	if err := m.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error shutting down proxies: %v", err)
	}
}
