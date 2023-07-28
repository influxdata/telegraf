//go:build !freebsd

package testutil

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}

type Container struct {
	BindMounts   map[string]string
	Entrypoint   []string
	Env          map[string]string
	ExposedPorts []string
	Cmd          []string
	Image        string
	Name         string
	Networks     []string
	WaitingFor   wait.Strategy

	Address string
	Ports   map[string]string
	Logs    TestLogConsumer

	container testcontainers.Container
	ctx       context.Context
}

func (c *Container) Start() error {
	c.ctx = context.Background()

	containerMounts := make([]testcontainers.ContainerMount, 0, len(c.BindMounts))
	for k, v := range c.BindMounts {
		containerMounts = append(containerMounts, testcontainers.BindMount(v, testcontainers.ContainerMountTarget(k)))
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Mounts:       testcontainers.Mounts(containerMounts...),
			Entrypoint:   c.Entrypoint,
			Env:          c.Env,
			ExposedPorts: c.ExposedPorts,
			Cmd:          c.Cmd,
			Image:        c.Image,
			Name:         c.Name,
			Networks:     c.Networks,
			WaitingFor:   c.WaitingFor,
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(c.ctx, req)
	if err != nil {
		return fmt.Errorf("container failed to start: %w", err)
	}
	c.container = container

	c.Logs = TestLogConsumer{
		Msgs: []string{},
	}
	c.container.FollowOutput(&c.Logs)
	err = c.container.StartLogProducer(c.ctx)
	if err != nil {
		return fmt.Errorf("log producer failed: %w", err)
	}

	c.Address = "localhost"

	err = c.LookupMappedPorts()
	if err != nil {
		c.Terminate()
		return fmt.Errorf("port lookup failed: %w", err)
	}

	return nil
}

// LookupMappedPorts creates a lookup table of exposed ports to mapped ports
func (c *Container) LookupMappedPorts() error {
	if len(c.ExposedPorts) == 0 {
		return nil
	}

	if len(c.Ports) == 0 {
		c.Ports = make(map[string]string)
	}

	for _, port := range c.ExposedPorts {
		// strip off leading host port: 80:8080 -> 8080
		if strings.Contains(port, ":") {
			port = strings.Split(port, ":")[1]
		}

		p, err := c.container.MappedPort(c.ctx, nat.Port(port))
		if err != nil {
			return fmt.Errorf("failed to find %q: %w", port, err)
		}

		// strip off the transport: 80/tcp -> 80
		if strings.Contains(port, "/") {
			port = strings.Split(port, "/")[0]
		}

		fmt.Printf("mapped container port %q to host port %q\n", port, p.Port())
		c.Ports[port] = p.Port()
	}

	return nil
}

func (c *Container) Exec(cmds []string) (int, io.Reader, error) {
	return c.container.Exec(c.ctx, cmds)
}

func (c *Container) PrintLogs() {
	fmt.Println("--- Container Logs Start ---")
	for _, msg := range c.Logs.Msgs {
		fmt.Print(msg)
	}
	fmt.Println("--- Container Logs End ---")
}

func (c *Container) Terminate() {
	err := c.container.StopLogProducer()
	if err != nil {
		fmt.Println(err)
	}

	err = c.container.Terminate(c.ctx)
	if err != nil {
		fmt.Printf("failed to terminate the container: %s", err)
	}

	c.PrintLogs()
}
