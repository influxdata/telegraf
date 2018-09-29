package docker_logs

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"io"
)

/*This file is inherited from telegraf docker input plugin*/
var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

func NewEnvClient() (Client, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

func NewClient(host string, tlsConfig *tls.Config) (Client, error) {
	url, err := docker.ParseHostURL(host)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	sockets.ConfigureTransport(transport, url.Scheme, url.Host)
	httpClient := &http.Client{Transport: transport}

	client, err := docker.NewClientWithOpts(docker.WithHost(host), docker.WithVersion(version),
		docker.WithHTTPClient(httpClient), docker.WithHTTPHeaders(defaultHeaders))
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
