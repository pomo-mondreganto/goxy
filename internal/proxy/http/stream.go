package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"goxy/internal/export"

	"github.com/google/uuid"
)

const (
	streamTerminationThreshold = time.Second * 15
)

type streamBinder struct {
	name       string
	listenPort int
	mu         sync.Mutex
	bindings   map[string]streamBinding
}

func (b *streamBinder) GetOrCreate(r *http.Request) *export.BasePacket {
	b.mu.Lock()
	defer b.mu.Unlock()
	bind, ok := b.bindings[r.RemoteAddr]
	if !ok || bind.LastSeen.Add(streamTerminationThreshold).Before(time.Now()) {
		bind = streamBinding{
			LastSeen:   time.Now(),
			BasePacket: b.getBasePacket(r),
		}
		b.bindings[r.RemoteAddr] = bind
		return bind.BasePacket
	}
	return bind.BasePacket
}

func (b *streamBinder) getBasePacket(r *http.Request) *export.BasePacket {
	srcHost, srcPort := splitAddrSafe(r.RemoteAddr, 0)
	dstHost, dstPort := splitAddrSafe(r.Host, b.listenPort)

	return &export.BasePacket{
		Source: fmt.Sprintf("goxy-%s", b.name),
		Endpoints: &export.EndpointData{
			IPSrc:   srcHost,
			IPDst:   dstHost,
			PortSrc: srcPort,
			PortDst: dstPort,
		},
		Proto:            "tcp",
		ProducerStreamID: uuid.New().String(),
	}
}

type streamBinding struct {
	LastSeen   time.Time
	BasePacket *export.BasePacket
}
