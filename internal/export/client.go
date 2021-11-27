package export

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	mgpb "goxy/lib/mongol"
)

func NewProducerClient(cc grpc.ClientConnInterface) *ProducerClient {
	return &ProducerClient{c: mgpb.NewMonGolClient(cc)}
}

type ProducerClient struct {
	c mgpb.MonGolClient
}

func (c *ProducerClient) Send(ctx context.Context, p *Packet) error {
	req := mgpb.AddPacketRequest{Packet: p.ToProto()}
	_, err := c.c.AddPacket(ctx, &req)
	if err != nil {
		return fmt.Errorf("making AddPacket request: %w", err)
	}
	return nil
}

func (c *ProducerClient) AddFilters(ctx context.Context, filters []string) ([]*mgpb.Filter, error) {
	req := mgpb.RegisterFiltersRequest{Names: filters}
	resp, err := c.c.RegisterFilters(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("making RegisterFilters request: %w", err)
	}
	return resp.GetFilters(), nil
}

func (c *ProducerClient) GetConfig(ctx context.Context) (*mgpb.Config, error) {
	resp, err := c.c.GetConfig(ctx, &mgpb.GetConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("making GetConfig request: %w", err)
	}
	return resp.GetConfig(), nil
}
