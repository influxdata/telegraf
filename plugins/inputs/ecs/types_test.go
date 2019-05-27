package ecs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseTask(t *testing.T) {
	r, err := os.Open("testdata/metadata.golden")
	require.NoError(t, err)

	parsed, err := unmarshalTask(r)
	require.NoError(t, err)

	require.Equal(t, validMeta, *parsed)
}

func Test_parseStats(t *testing.T) {
	r, err := os.Open("testdata/stats.golden")
	require.NoError(t, err)

	parsed, err := unmarshalStats(r)
	require.NoError(t, err)
	require.Equal(t, validStats, parsed)
}

func Test_mergeTaskStats(t *testing.T) {
	metadata, err := os.Open("testdata/metadata.golden")
	require.NoError(t, err)

	parsedMetadata, err := unmarshalTask(metadata)
	require.NoError(t, err)

	stats, err := os.Open("testdata/stats.golden")
	require.NoError(t, err)

	parsedStats, err := unmarshalStats(stats)
	require.NoError(t, err)

	mergeTaskStats(parsedMetadata, parsedStats)

	for _, cont := range parsedMetadata.Containers {
		require.Equal(t, validStats[cont.ID], cont.Stats)
	}
}
