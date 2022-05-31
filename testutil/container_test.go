package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := Container{
		Image: "docksal/empty",
	}

	err := container.Start()
	require.NoError(t, err)

	err = container.Terminate()
	require.NoError(t, err)
}

func TestMappedPortLookup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cases := []struct {
		name     string
		port     string
		expected string
	}{
		{"random", "80", "80"},
		{"only 80", "80:80", "80"},
		{"only 80", "80:80/tcp", "80"},
		{"only 8080", "8080:80", "8080"},
		{"only 8080", "8080:80/tcp", "8080"},
	}

	for _, tc := range cases {
		container := Container{
			Image:        "nginx:stable-alpine",
			ExposedPorts: []string{tc.port},
		}

		err := container.Start()
		require.NoError(t, err)

		if tc.name == "random" {
			require.NotEqual(t, tc.expected, container.Ports["80"])
		} else {
			require.Equal(t, tc.expected, container.Ports["80"])
		}

		err = container.Terminate()
		require.NoError(t, err)
	}
}

func TestBadImageName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := Container{
		Image: "fAk3-n4mE",
	}

	err := container.Start()
	require.Error(t, err)
}
