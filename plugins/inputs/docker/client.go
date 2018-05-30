package docker

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
)

var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client interface {
	Info(ctx context.Context) (types.Info, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error)
	NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
}

func NewEnvClient() (Client, error) {
	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

func NewClient(host string, tlsConfig *tls.Config) (Client, error) {
	proto, addr, _, err := docker.ParseHost(host)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	sockets.ConfigureTransport(transport, proto, addr)
	httpClient := &http.Client{Transport: transport}

	client, err := docker.NewClient(host, version, httpClient, defaultHeaders)
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
func (c *SocketClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.client.ServiceList(ctx, options)
}
func (c *SocketClient) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	return c.client.TaskList(ctx, options)
}
func (c *SocketClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	return c.client.NodeList(ctx, options)
}
