//go:build !freebsd
// +build !freebsd

package testutil

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Container struct {
	Image        string
	Entrypoint   []string
	Env          map[string]string
	ExposedPorts []string
	WaitingFor   wait.Strategy

	Address string
	Port    string

	container testcontainers.Container
	ctx       context.Context
}

func (c *Container) Start() error {
	c.ctx = context.Background()

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        c.Image,
			Env:          c.Env,
			ExposedPorts: c.ExposedPorts,
			Entrypoint:   c.Entrypoint,
			WaitingFor:   c.WaitingFor,
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(c.ctx, req)
	if err != nil {
		return fmt.Errorf("container failed to start: %s", err)
	}
	c.container = container

	c.Address, err = c.container.Host(c.ctx)
	if err != nil {
		return fmt.Errorf("container host address failed: %s", err)
	}

	// assume the first port is the one the test will connect to
	// additional ports can be used for the waiting for section
	if len(c.ExposedPorts) > 0 {
		p, err := c.container.MappedPort(c.ctx, nat.Port(c.ExposedPorts[0]))
		if err != nil {
			return fmt.Errorf("container host port failed: %s", err)
		}
		c.Port = p.Port()
	}

	return nil
}

func (c *Container) Terminate() error {
	err := c.container.Terminate(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to terminate the container: %s", err)
	}

	return nil
}
