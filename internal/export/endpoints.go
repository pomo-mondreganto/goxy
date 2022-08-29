package export

import (
	"fmt"

	mgpb "goxy/lib/mongol"
)

// EndpointData stores packet endpoints in plain format.
type EndpointData struct {
	IPSrc   string
	IPDst   string
	PortSrc int
	PortDst int
}

func (e *EndpointData) String() string {
	return fmt.Sprintf("%s:%d->%s:%d", e.IPSrc, e.PortSrc, e.IPDst, e.PortDst)
}

// Reversed returns EndpointData with swapped source and destination.
func (e *EndpointData) Reversed() *EndpointData {
	res := &EndpointData{
		IPSrc:   e.IPDst,
		IPDst:   e.IPSrc,
		PortSrc: e.PortDst,
		PortDst: e.PortSrc,
	}
	return res
}

// ReversedString returns a string representation of reversed endpoints (no need to allocate new structure).
func (e *EndpointData) ReversedString() string {
	return fmt.Sprintf("%s:%d->%s:%d", e.IPDst, e.PortDst, e.IPSrc, e.PortSrc)
}

func (e *EndpointData) ToProto() *mgpb.EndpointData {
	return &mgpb.EndpointData{
		IpSrc:   e.IPSrc,
		IpDst:   e.IPDst,
		PortSrc: int32(e.PortSrc),
		PortDst: int32(e.PortDst),
	}
}
