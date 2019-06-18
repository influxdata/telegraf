package docker_log

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

/*This file is inherited from telegraf docker input plugin*/
var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func NewEnvClient() (Client, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

func NewClient(host string, tlsConfig *tls.Config) (Client, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}
	client, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithHTTPClient(httpClient),
		docker.WithVersion(version),
		docker.WithHost(host))

	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

type SocketClient struct {
	client *docker.Client
}

func (c *SocketClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return c.client.ContainerList(ctx, options)
}

func (c *SocketClient) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}
func (c *SocketClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.client.ContainerInspect(ctx, containerID)
}
