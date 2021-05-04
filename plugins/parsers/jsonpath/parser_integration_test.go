package jsonpath

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TAKES ABOUT 3 MINUTES, how to improve?

func TestSimple(t *testing.T) {
	ctx := context.Background()
	dir, _ := os.Getwd()
	fmt.Println(dir)
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    dir + "/../../../",
			Dockerfile: "scripts/alpine.docker",
		},
		Cmd: []string{
			"telegraf", "--config", "./plugins/parsers/jsonpath/testdata/simple/simple.config",
		},
	}

	telegrafDocker, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer telegrafDocker.Terminate(ctx)
	rc, err := telegrafDocker.Logs(ctx)
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	fmt.Println(buf.String())
	require.NoError(t, err)

	ip, err := telegrafDocker.Host(ctx)
	require.NoError(t, err)
	fmt.Println(ip)
}
