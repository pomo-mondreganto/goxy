package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"goxy/internal/common"
	"goxy/internal/proxy/http"
	"goxy/internal/proxy/tcp"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	httpfilters "goxy/internal/proxy/http/filters"
	tcpfilters "goxy/internal/proxy/tcp/filters"
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

	pc := new(common.ProxyConfig)
	if err := viper.Unmarshal(&pc); err != nil {
		logrus.Fatal("Error parsing proxy config from file: ", err)
	}

	tcpRuleSet, err := tcpfilters.NewRuleSet(pc.Rules)
	if err != nil {
		logrus.Fatalf("Error creating tcp ruleset: %v", err)
	}

	httpRuleSet, err := httpfilters.NewRuleSet(pc.Rules)
	if err != nil {
		logrus.Fatalf("Error creating http ruleset: %v", err)
	}

	tcpProxies := make([]*tcp.Proxy, 0)
	for _, s := range pc.Services {
		if s.Type == "tcp" {
			p, err := tcp.NewProxy(s, tcpRuleSet)
			if err != nil {
				logrus.Fatalf("Error creating tcp proxy: %v", err)
			}
			tcpProxies = append(tcpProxies, p)
		}
	}

	httpProxies := make([]*http.Proxy, 0)
	for _, s := range pc.Services {
		if s.Type == "http" {
			p, err := http.NewProxy(s, httpRuleSet)
			if err != nil {
				logrus.Fatalf("Error creating http proxy: %v", err)
			}
			httpProxies = append(httpProxies, p)
		}
	}

	wg := sync.WaitGroup{}
	for _, p := range tcpProxies {
		wg.Add(1)
		go func(p *tcp.Proxy) {
			defer wg.Done()
			if err := p.Start(); err != nil {
				logrus.Fatalf("Error starting tcp proxy: %v", err)
			}
		}(p)
	}

	for _, p := range httpProxies {
		wg.Add(1)
		go func(p *http.Proxy) {
			defer wg.Done()
			if err := p.Start(); err != nil {
				logrus.Fatalf("Error starting http proxy: %v", err)
			}
		}(p)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c

	logrus.Info("Shutting down tcp proxies")
	for _, p := range tcpProxies {
		wg.Add(1)
		go func(p *tcp.Proxy) {
			defer wg.Done()
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			if err := p.Shutdown(ctx); err != nil {
				logrus.Fatalf("Error shutting down tcp proxy: %v", err)
			}
		}(p)
	}

	logrus.Info("Shutting down http proxies")
	for _, p := range httpProxies {
		wg.Add(1)
		go func(p *http.Proxy) {
			defer wg.Done()
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			if err := p.Shutdown(ctx); err != nil {
				logrus.Fatalf("Error shutting down http proxy: %v", err)
			}
		}(p)
	}

	wg.Wait()
	logrus.Info("Shutdown successful")
}
