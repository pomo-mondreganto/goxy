package main

import (
	"github.com/sirupsen/logrus"
	"log"
	"net"
)

func main() {
	for i := 0; i < 10000; i++ {
		c, err := net.Dial("tcp", "localhost:1337")
		if err != nil {
			log.Fatalf("Failed to create conn: %v", err)
		}
		if _, err = c.Write([]byte("kek\n")); err != nil {
			logrus.Errorf("Error on write: %v", err)
		}
		if err = c.Close(); err != nil {
			logrus.Errorf("Error on close: %v", err)
		}
	}
}
