package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

var (
	version        string
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client interface {
	Info(ctx context.Context) (types.Info, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func NewEnvClient() (Client, error) {
	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

func NewClient(host string) (Client, error) {
	client, err := docker.NewClient(host, version, nil, defaultHeaders)
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

type SocketClient struct {
	client *docker.Client
}

func (c *SocketClient) Info(ctx context.Context) (types.Info, error) {
	return c.client.Info(ctx)
}
func (c *SocketClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return c.client.ContainerList(ctx, options)
}
func (c *SocketClient) ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return c.client.ContainerStats(ctx, containerID, stream)
}
func (c *SocketClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.client.ContainerInspect(ctx, containerID)
}
