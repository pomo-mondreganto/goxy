package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"goxy/internal/common"
	"goxy/internal/proxy"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	configFile = flag.String("config", "config.yml", "Path to the config file in YAML format")
	logLevel   = flag.String("log_level", "DEBUG", "Log level {INFO|DEBUG|WARNING|ERROR}")
)

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

func parseConfig() {
	viper.SetConfigFile(*configFile)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("Error reading config from yaml: ", err)
	}
}

func main() {
	flag.Parse()

	setLogLevel()
	parseConfig()

	cfg := new(common.ProxyConfig)
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Fatal("Error parsing proxy config from file: ", err)
	}

	m, err := proxy.NewManager(cfg)
	if err != nil {
		logrus.Fatalf("Error creating proxy manager: %v", err)
	}
	if err := m.StartAll(); err != nil {
		logrus.Fatalf("Error starting proxy manager: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	logrus.Info("Shutting down proxies")
	if err := m.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error shutting down proxies: %v", err)
	}
	logrus.Info("Shutdown successful")
}
