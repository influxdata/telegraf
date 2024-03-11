//go:build windows
// +build windows

package win_wmi

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var sysDrive = os.Getenv("SystemDrive") + `\` // C:\

func TestBuildWqlStatements(t *testing.T) {
	plugin := &Wmi{
		Queries: []Query{
			{
				Namespace:  "ROOT\\cimv2",
				ClassName:  "Win32_Volume",
				Properties: []string{"Name", "FreeSpace", "Purpose"},
				//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
				Filter:               fmt.Sprintf(`NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`, regexp.QuoteMeta(sysDrive)),
				TagPropertiesInclude: []string{"Name"},
			},
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NotEmpty(t, plugin.Queries)

	//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
	expected := fmt.Sprintf(
		`SELECT Name, FreeSpace, Purpose FROM Win32_Volume WHERE NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`,
		regexp.QuoteMeta(sysDrive),
	)
	require.Equal(t, expected, plugin.Queries[0].query)
}

func TestInit(t *testing.T) {
	plugin := &Wmi{}
	require.NoError(t, plugin.Init())
}

func TestQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	plugin := &Wmi{
		Queries: []Query{
			{
				Namespace:  "ROOT\\cimv2",
				ClassName:  "Win32_Volume",
				Properties: []string{"Name", "FreeSpace", "Purpose"},
				//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
				Filter:               fmt.Sprintf(`NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`, regexp.QuoteMeta(sysDrive)),
				TagPropertiesInclude: []string{"Name"},
			},
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)
	// Only one metric was returned (because we filtered for SystemDrive)
	require.Len(t, acc.Metrics, 1)
	// Name property collected and is the SystemDrive
	require.Equal(t, sysDrive, acc.Metrics[0].Tags["Name"])
	// FreeSpace property was collected as a field
	require.NotEmpty(t, acc.Metrics[0].Fields["FreeSpace"])
}
