package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"goxy/internal/common"
	tcpfilters "goxy/internal/filters/tcp"
	"goxy/internal/proxy/tcp"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	viper.SetConfigFile("config.yml")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("Error reading config from yaml: ", err)
	}
}

func main() {
	pc := new(common.ProxyConfig)
	if err := viper.Unmarshal(&pc); err != nil {
		logrus.Fatal("Error parsing proxy config from file: ", err)
	}

	tcpRuleSet, err := tcpfilters.NewRuleSet(pc.Rules)
	if err != nil {
		logrus.Fatal("Error creating tcp ruleset: ", err)
	}

	tcpProxies := make([]*tcp.Proxy, 0)
	for _, s := range pc.Services {
		if s.Type == "tcp" {
			p, err := tcp.NewProxy(&s, tcpRuleSet)
			if err != nil {
				logrus.Fatal("Error creating tcp proxy: ", err)
			}
			tcpProxies = append(tcpProxies, p)
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
	wg.Wait()
	logrus.Info("Shutdown successful")
}
