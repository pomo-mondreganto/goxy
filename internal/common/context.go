package common

import "github.com/sirupsen/logrus"

type ConnectionContext struct {
	MustDrop   bool
	MustAccept bool
	Counters   map[string]int
}

func (c *ConnectionContext) DumpFields() logrus.Fields {
	fields := make(logrus.Fields)
	for k, v := range c.Counters {
		fields[k] = v
	}
	return fields
}

func NewContext() *ConnectionContext {
	return &ConnectionContext{
		Counters: make(map[string]int),
	}
}
