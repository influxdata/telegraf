package docker_log

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

type dockerClient interface {
	// ContainerList lists the containers in the Docker environment.
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	// ContainerLogs retrieves the logs of a specific container.
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	// ContainerInspect inspects a specific container and retrieves its details.
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
		docker.WithHTTPHeaders(map[string]string{"User-Agent": "engine-api-cli-1.0"}),
		docker.WithHTTPClient(httpClient),
		docker.WithAPIVersionNegotiation(),
		docker.WithHost(host))

	if err != nil {
		return nil, err
	}
	return &socketClient{client}, nil
}

type socketClient struct {
	client *docker.Client
}

// ContainerList lists the containers in the Docker environment.
func (c *socketClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return c.client.ContainerList(ctx, options)
}

// ContainerLogs retrieves the logs of a specific container.
func (c *socketClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}

// ContainerInspect inspects a specific container and retrieves its details.
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return c.client.ContainerInspect(ctx, containerID)
}
