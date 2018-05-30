package hystrix_stream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getCounterSreams(t *testing.T) {
	entry := HystrixStreamEntry{
		RequestCount:               23,
		ErrorCount:                 91,
		ErrorPercentage:            14,
		RollingCountShortCircuited: 2,
	}
	fields := getCounterFields(entry)

	require.Equal(t, 23, fields["RequestCount"])
	require.Equal(t, 91, fields["ErrorCount"])
	require.Equal(t, 14, fields["ErrorPercentage"])
	require.Equal(t, 2, fields["RollingCountShortCircuited"])
}

func Test_getTags(t *testing.T) {
	tags := getTags(HystrixStreamEntry{
		Name:       "name",
		Type:       "type",
		Group:      "group",
		ThreadPool: "tpool",
	})

	require.Equal(t, "group", tags["group"])
	require.Equal(t, "type", tags["type"])
	require.Equal(t, "tpool", tags["threadpool"])
	require.Equal(t, "name", tags["name"])
}
