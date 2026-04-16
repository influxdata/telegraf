package docker_log

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type dockerClient interface {
	// ContainerList lists the containers in the Docker environment.
	ContainerList(ctx context.Context, options client.ContainerListOptions) ([]container.Summary, error)
	// ContainerLogs retrieves the logs of a specific container.
	ContainerLogs(ctx context.Context, containerID string, options client.ContainerLogsOptions) (io.ReadCloser, error)
	// ContainerInspect inspects a specific container and retrieves its details.
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}

func newEnvClient() (dockerClient, error) {
	c, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &socketClient{c}, nil
}

func newClient(host string, tlsConfig *tls.Config) (dockerClient, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}
	c, err := client.New(
		client.WithHTTPHeaders(map[string]string{"User-Agent": "engine-api-cli-1.0"}),
		client.WithHTTPClient(httpClient),
		client.WithHost(host),
	)
	if err != nil {
		return nil, err
	}

	return &socketClient{client: c}, nil
}

type socketClient struct {
	client *client.Client
}

// ContainerList lists the containers in the Docker environment.
func (c *socketClient) ContainerList(ctx context.Context, options client.ContainerListOptions) ([]container.Summary, error) {
	result, err := c.client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// ContainerLogs retrieves the logs of a specific container.
func (c *socketClient) ContainerLogs(ctx context.Context, containerID string, options client.ContainerLogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}

// ContainerInspect inspects a specific container and retrieves its details.
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	result, err := c.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return container.InspectResponse{}, err
	}
	return result.Container, nil
}
