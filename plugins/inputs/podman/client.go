package podman

import (
	"context"
	"fmt"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/system"
	"github.com/containers/podman/v3/pkg/domain/entities"
)

type Client interface {
	Info() (*define.Info, error)
	ContainerList(filters map[string][]string) ([]entities.ListContainer, error)
	ContainerStats(string) (*entities.ContainerStatsReport, error)
}

func NewClient(socket string, ctx context.Context) (Client, error) {
	// Connect to Podman socket
	connText, err := bindings.NewConnection(ctx, socket)
	if err != nil {
		return nil, err
	}
	return &SocketClient{connText}, nil
}

type SocketClient struct {
	client context.Context
}

func (c *SocketClient) Info() (*define.Info, error) {
	return system.Info(c.client, nil)
}

func (c *SocketClient) ContainerList(filters map[string][]string) ([]entities.ListContainer, error) {
	return containers.List(c.client, nil)
}

func (c *SocketClient) ContainerStats(container string) (*entities.ContainerStatsReport, error) {
	stats, err := containers.Stats(c.client, []string{container}, nil)
	if err != nil {
		return nil, err
	} else if len(stats) != 1 {
		return nil, fmt.Errorf("Invalid number of stats")
	}
	containerStats := <-stats
	return &containerStats, nil
}
