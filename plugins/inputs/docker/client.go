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
	// Info retrieves system-wide information about the Docker server.
	Info(ctx context.Context) (system.Info, error)
	// ContainerList retrieves a list of containers based on the specified options.
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	// ContainerStats retrieves real-time statistics for a specific container.
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	// ContainerInspect retrieves detailed information about a specific container.
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	// ServiceList retrieves a list of services based on the specified options.
	ServiceList(ctx context.Context, options swarm.ServiceListOptions) ([]swarm.Service, error)
	// TaskList retrieves a list of tasks based on the specified options.
	TaskList(ctx context.Context, options swarm.TaskListOptions) ([]swarm.Task, error)
	// NodeList retrieves a list of nodes based on the specified options.
	NodeList(ctx context.Context, options swarm.NodeListOptions) ([]swarm.Node, error)
	// DiskUsage retrieves disk usage information.
	DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error)
	// ClientVersion retrieves the version of the Docker client.
	ClientVersion() string
	// Close releases any resources held by the client.
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

// Info retrieves system-wide information about the Docker server.
func (c *socketClient) Info(ctx context.Context) (system.Info, error) {
	return c.client.Info(ctx)
}

// ContainerList retrieves a list of containers based on the specified options.
func (c *socketClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return c.client.ContainerList(ctx, options)
}

// ContainerStats retrieves real-time statistics for a specific container.
func (c *socketClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	return c.client.ContainerStats(ctx, containerID, stream)
}

// ContainerInspect retrieves detailed information about a specific container.
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return c.client.ContainerInspect(ctx, containerID)
}

// ServiceList retrieves a list of services based on the specified options.
func (c *socketClient) ServiceList(ctx context.Context, options swarm.ServiceListOptions) ([]swarm.Service, error) {
	return c.client.ServiceList(ctx, options)
}

// TaskList retrieves a list of tasks based on the specified options.
func (c *socketClient) TaskList(ctx context.Context, options swarm.TaskListOptions) ([]swarm.Task, error) {
	return c.client.TaskList(ctx, options)
}

// NodeList retrieves a list of nodes based on the specified options.
func (c *socketClient) NodeList(ctx context.Context, options swarm.NodeListOptions) ([]swarm.Node, error) {
	return c.client.NodeList(ctx, options)
}

// DiskUsage retrieves disk usage information.
func (c *socketClient) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return c.client.DiskUsage(ctx, options)
}

// ClientVersion retrieves the version of the Docker client.
func (c *socketClient) ClientVersion() string {
	return c.client.ClientVersion()
}

// Close releases any resources held by the client.
func (c *socketClient) Close() error {
	return c.client.Close()
}
