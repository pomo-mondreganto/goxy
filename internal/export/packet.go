package export

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mgpb "goxy/lib/mongol"
)

// BasePacket contains common data shared by stream.
type BasePacket struct {
	Source           string
	Endpoints        *EndpointData
	Proto            string
	ProducerStreamID string
}

// Packet extends BasePacket and adds packet-specific info, so all data in BasePacket is not reallocated.
type Packet struct {
	*BasePacket

	Content     []byte
	CaptureTime time.Time
	FilterData  uint64
	Inbound     bool
	Reversed    bool
}

// DumpEndpoints correctly returns packet endpoints (reversing if the packet is reversed).
func (p *Packet) DumpEndpoints() string {
	if p.Reversed {
		return p.Endpoints.ReversedString()
	}
	return p.Endpoints.String()
}

func (p *Packet) ToProto() *mgpb.Packet {
	return &mgpb.Packet{
		Source:           p.Source,
		Inbound:          p.Inbound,
		Reversed:         p.Reversed,
		Endpoints:        p.Endpoints.ToProto(),
		Proto:            p.Proto,
		Content:          p.Content,
		CaptureTime:      timestamppb.New(p.CaptureTime),
		FilterData:       p.FilterData,
		ProducerStreamId: p.ProducerStreamID,
	}
}
