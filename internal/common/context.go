package common

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type ConnectionContext struct {
	flags    map[string]bool
	counters map[string]int
	mu       sync.RWMutex
}

func (c *ConnectionContext) DumpFields() logrus.Fields {
	fields := make(logrus.Fields)
	for k, v := range c.counters {
		fields[k] = v
	}
	return fields
}

func (c *ConnectionContext) AddToCounter(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[key] += value
}

func (c *ConnectionContext) GetCounter(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := c.counters[key]
	return val
}

func (c *ConnectionContext) SetFlag(flag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flags[flag] = true
}

func (c *ConnectionContext) GetFlag(flag string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val := c.flags[flag]
	return val
}

func NewContext() *ConnectionContext {
	return &ConnectionContext{
		counters: make(map[string]int),
	}
}
