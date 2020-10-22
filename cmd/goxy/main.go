package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"goxy/internal/proxy"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	p := proxy.NewTcpProxy("127.0.0.1", 1337, 1338)
	go func() {
		if err := p.Start(); err != nil {
			logrus.Fatalf("Error starting tcp proxy: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c

	logrus.Info("Shutting down tcp proxy")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	if err := p.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error shutting down tcp proxy: %v", err)
	}
	logrus.Info("Shutdown successful")
}
