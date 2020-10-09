package main

import (
	"goxy/internal/proxy"
	"log"
)

func main() {
	p := proxy.TcpProxy{
		Port:       8080,
		RemoteAddr: "127.0.0.1",
	}
	err := p.Start()
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}
	p.Wait()
}
