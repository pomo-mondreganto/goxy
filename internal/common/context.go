package common

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type ProxyContext struct {
	flags    map[string]bool
	counters map[string]int
	mu       sync.RWMutex
}

func (c *ProxyContext) DumpFields() logrus.Fields {
	fields := make(logrus.Fields)
	for k, v := range c.counters {
		fields[k] = v
	}
	for k, v := range c.flags {
		if v {
			fields[k] = v
		}
	}
	return fields
}

func (c *ProxyContext) AddToCounter(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[key] += value
}

func (c *ProxyContext) GetCounter(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := c.counters[key]
	return val
}

func (c *ProxyContext) SetFlag(flag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flags[flag] = true
}

func (c *ProxyContext) GetFlag(flag string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := c.flags[flag]
	return val
}

func NewProxyContext() *ProxyContext {
	return &ProxyContext{
		counters: make(map[string]int),
		flags:    make(map[string]bool),
	}
}
