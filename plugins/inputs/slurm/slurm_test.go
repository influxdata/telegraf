package slurm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLs(t *testing.T) {
	for _, url := range []string{"http://example.com:6820", "https://example.com:6820"} {
		plugin := Slurm{
			URL: url,
		}
		require.NoError(t, plugin.Init())
	}

	for _, url := range []string{"httpp://example.com:6820", "httpss://example.com:6820"} {
		plugin := Slurm{
			URL: url,
		}
		require.Error(t, plugin.Init())
	}
}
