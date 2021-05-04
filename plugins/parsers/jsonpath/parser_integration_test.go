package jsonpath

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TAKES ABOUT 3 MINUTES, how to improve?

func TestSimple(t *testing.T) {
	ctx := context.Background()

	dir, _ := os.Getwd()
	localBindDir := dir + "/testdata"

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    dir + "/../../../",
			Dockerfile: "scripts/integration_tests.docker",
		},
		BindMounts: map[string]string{
			localBindDir: "/tmp/",
		},
	}

	telegrafDocker, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer telegrafDocker.Terminate(ctx)

	_, err = telegrafDocker.Exec(ctx, []string{
		"telegraf", "--config", "./plugins/parsers/jsonpath/testdata/simple/simple.conf", "--once",
	})
	require.NoError(t, err)

	expectedMetric := `lol,host=docker name="John"`

	dat, err := ioutil.ReadFile(localBindDir + "/simple.out")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(dat), expectedMetric))
	err = os.Remove(localBindDir + "/simple.out")
	require.NoError(t, err)
}
