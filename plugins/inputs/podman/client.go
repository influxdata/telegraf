package podman

import (
	"context"
	"fmt"
	"time"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/system"
	"github.com/containers/podman/v3/pkg/domain/entities"
)

type Client interface {
	Info() (*define.Info, error)
	ContainerList(ctx context.Context, filters map[string][]string) ([]entities.ListContainer, error)
	ContainerStats(context.Context, string) (*define.ContainerStats, error)
	Background() context.Context
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

func (c *SocketClient) Background() context.Context {
	return c.client
}

func (c *SocketClient) ContainerList(ctx context.Context, filters map[string][]string) ([]entities.ListContainer, error) {
	return containers.List(ctx, nil)
}

//For now it will return the first recieved container's stat
func (c *SocketClient) ContainerStats(ctx context.Context, container string) (*define.ContainerStats, error) {
	stats, err := containers.Stats(ctx, []string{container}, nil)
	if err != nil {
		return nil, err
	}
	select {
	case containerStats := <-stats:
		if containerStats.Error != nil {
			return nil, containerStats.Error
		} else if len(containerStats.Stats) <= 0 {
			return nil, errNoStats
		}
		//return last recieved stat
		return &containerStats.Stats[len(containerStats.Stats)-1], nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("Invalid number of stats")
	}
}
