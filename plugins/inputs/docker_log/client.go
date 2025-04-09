package docker_log

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

// This file is inherited from telegraf docker input plugin
var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type dockerClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}

func newEnvClient() (dockerClient, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &socketClient{client}, nil
}

func newClient(host string, tlsConfig *tls.Config) (dockerClient, error) {
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
	return &socketClient{client}, nil
}

type socketClient struct {
	client *docker.Client
}

func (c *socketClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return c.client.ContainerList(ctx, options)
}

func (c *socketClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return c.client.ContainerInspect(ctx, containerID)
}
