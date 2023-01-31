package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

var (
	timeout     = 120
	killTimeout = 5
)

type BytesBuffer struct {
	*bytes.Buffer
}

func (b *BytesBuffer) Close() error {
	return nil
}

type LXDClient struct {
	Client lxd.InstanceServer
}

// Connect to the LXD socket.
func (c *LXDClient) Connect() error {
	client, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		client, err = lxd.ConnectLXDUnix("/var/snap/lxd/common/lxd/unix.socket", nil)
		if err != nil {
			return err
		}
	}

	c.Client = client
	return nil
}

// Create a container using a specific remote and alias.
func (c *LXDClient) Create(name string, remote string, alias string) error {
	fmt.Printf("creating %s with %s:%s\n", name, remote, alias)

	if c.Client == nil {
		err := c.Connect()
		if err != nil {
			return err
		}
	}

	server := ""
	switch remote {
	case "images":
		server = "https://images.linuxcontainers.org"
	case "ubuntu":
		server = "https://cloud-images.ubuntu.com/releases"
	case "ubuntu-daily":
		server = "https://cloud-images.ubuntu.com/daily"
	default:
		return fmt.Errorf("unknown remote: %s", remote)
	}

	req := api.ContainersPost{
		Name: name,
		Source: api.ContainerSource{
			Type:     "image",
			Mode:     "pull",
			Protocol: "simplestreams",
			Server:   server,
			Alias:    alias,
		},
	}

	// Get LXD to create the container (background operation)
	op, err := c.Client.CreateContainer(req)
	if err != nil {
		return err
	}

	// Wait for the operation to complete
	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Delete the given container.
func (c *LXDClient) Delete(name string) error {
	fmt.Println("deleting", name)
	op, err := c.Client.DeleteContainer(name)
	if err != nil {
		return err
	}

	return op.Wait()
}

// Run a command returning result struct. Will kill commands that take longer
// than 120 seconds.
func (c *LXDClient) Exec(name string, command ...string) error {
	fmt.Printf("$ %s\n", strings.Join(command, " "))

	cmd := []string{"/usr/bin/timeout", "-k", strconv.Itoa(killTimeout), strconv.Itoa(timeout)}
	cmd = append(cmd, command...)
	req := api.ContainerExecPost{
		Command:   cmd,
		WaitForWS: true,
	}

	output := &BytesBuffer{bytes.NewBuffer(nil)}

	args := lxd.ContainerExecArgs{
		Stdout:   output,
		Stderr:   output,
		DataDone: make(chan bool),
	}

	op, err := c.Client.ExecContainer(name, req, &args)
	if err != nil {
		return err
	}

	err = op.Wait()
	if err != nil {
		return err
	}

	// Wait for any remaining I/O to be flushed
	<-args.DataDone

	// get the return code
	opAPI := op.Get()
	rc := int(opAPI.Metadata["return"].(float64))

	if rc != 0 {
		return fmt.Errorf(output.String())
	}

	fmt.Println(output.String())

	return nil
}

// Push file to container.
func (c *LXDClient) Push(name string, src string, dst string) error {
	fmt.Printf("cp %s %s%s\n", src, name, dst)
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", src, err)
	}
	defer f.Close()

	return c.Client.CreateContainerFile(name, dst, lxd.ContainerFileArgs{
		Content: f,
		Mode:    0644,
	})
}

// Start the given container.
func (c *LXDClient) Start(name string) error {
	fmt.Println("starting", name)
	reqState := api.ContainerStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err := c.Client.UpdateContainerState(name, reqState, "")
	if err != nil {
		return err
	}

	return op.Wait()
}

// Stop the given container.
func (c *LXDClient) Stop(name string) error {
	fmt.Println("stopping", name)
	reqState := api.ContainerStatePut{
		Action:  "stop",
		Force:   true,
		Timeout: 10,
	}

	op, err := c.Client.UpdateContainerState(name, reqState, "")
	if err != nil {
		return err
	}

	return op.Wait()
}
