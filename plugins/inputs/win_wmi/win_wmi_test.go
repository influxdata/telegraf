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

// initialize test data
var sysDrive = fmt.Sprintf(`%s\`, os.Getenv("SystemDrive")) // C:\

// include Name as a tag, FreeSpace as a field, and Purpose as a known-null class property
var testQuery = Query{
	Namespace:  "ROOT\\cimv2",
	ClassName:  "Win32_Volume",
	Properties: []string{"Name", "FreeSpace", "Purpose"},
	//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
	Filter:               fmt.Sprintf(`NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`, regexp.QuoteMeta(sysDrive)),
	TagPropertiesInclude: []string{"Name"},
	tagFilter:            nil, // this is filled in by CompileInputs()
}

//nolint:gocritic // sprintfQuotedString - "%s" used by purpose, string escaping is done by special function
var expectedWql = fmt.Sprintf(
	`SELECT Name, FreeSpace, Purpose FROM Win32_Volume WHERE NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`,
	regexp.QuoteMeta(sysDrive),
)

// test buildWqlStatements
func TestWmi_buildWqlStatements(t *testing.T) {
	var logger = new(testutil.Logger)
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, compileInputs(&plugin))
	require.Equal(t, expectedWql, plugin.Queries[0].query)
}

// test DoQuery
func TestWmi_DoQuery(t *testing.T) {
	var logger = new(testutil.Logger)
	var acc = new(testutil.Accumulator)
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, compileInputs(&plugin))
	for _, q := range plugin.Queries {
		require.NoError(t, q.doQuery(acc))
	}
	// no errors in accumulator
	require.Empty(t, acc.Errors)
	// Only one metric was returned (because we filtered for SystemDrive)
	require.Len(t, acc.Metrics, 1)
	// Name property collected and is the SystemDrive
	require.Equal(t, sysDrive, acc.Metrics[0].Tags["Name"])
	// FreeSpace property was collected as a field
	require.NotEmpty(t, acc.Metrics[0].Fields["FreeSpace"])
}

// test Init function
func TestWmi_Init(t *testing.T) {
	var logger = new(testutil.Logger)
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, plugin.Init())
}

// test Gather function
func TestWmi_Gather(t *testing.T) {
	var logger = new(testutil.Logger)
	var acc = new(testutil.Accumulator)
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Gather(acc))
	// no errors in accumulator
	require.Empty(t, acc.Errors)
}
