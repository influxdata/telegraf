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
