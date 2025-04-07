package docker

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

var (
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type dockerClient interface {
	Info(ctx context.Context) (system.Info, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error)
	NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
	DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error)
	ClientVersion() string
	Close() error
}

func newEnvClient() (dockerClient, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &socketClient{dockerClient}, nil
}

func newClient(host string, tlsConfig *tls.Config) (dockerClient, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}

	dockerClient, err := client.NewClientWithOpts(
		client.WithHTTPHeaders(defaultHeaders),
		client.WithHTTPClient(httpClient),
		client.WithAPIVersionNegotiation(),
		client.WithHost(host))
	if err != nil {
		return nil, err
	}

	return &socketClient{dockerClient}, nil
}

type socketClient struct {
	client *client.Client
}

func (c *socketClient) Info(ctx context.Context) (system.Info, error) {
	return c.client.Info(ctx)
}
func (c *socketClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return c.client.ContainerList(ctx, options)
}
func (c *socketClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	return c.client.ContainerStats(ctx, containerID, stream)
}
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return c.client.ContainerInspect(ctx, containerID)
}
func (c *socketClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.client.ServiceList(ctx, options)
}
func (c *socketClient) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	return c.client.TaskList(ctx, options)
}
func (c *socketClient) NodeList(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error) {
	return c.client.NodeList(ctx, options)
}
func (c *socketClient) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return c.client.DiskUsage(ctx, options)
}

func (c *socketClient) ClientVersion() string {
	return c.client.ClientVersion()
}

func (c *socketClient) Close() error {
	return c.client.Close()
}
