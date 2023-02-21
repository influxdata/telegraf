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
var logger = new(testutil.Logger)
var acc = new(testutil.Accumulator)

// include Name as a tag, FreeSpace as a field, and Purpose as a known-null class property
var testQuery Query = Query{
	Namespace:            "ROOT\\cimv2",
	ClassName:            "Win32_Volume",
	Properties:           []string{"Name", "FreeSpace", "Purpose"},
	Filter:               fmt.Sprintf(`NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`, regexp.QuoteMeta(sysDrive)),
	TagPropertiesInclude: []string{"Name"},
	tagFilter:            nil, // this is filled in by CompileInputs()
}
var expectedWql = fmt.Sprintf(
	`SELECT Name, FreeSpace, Purpose FROM Win32_Volume WHERE NOT Name LIKE "\\\\?\\%%" AND Name LIKE "%s"`,
	regexp.QuoteMeta(sysDrive))

// test buildWqlStatements
func TestWmi_buildWqlStatements(t *testing.T) {
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, compileInputs(&plugin))
	require.Equal(t, expectedWql, plugin.Queries[0].query)
}

// test DoQuery
func TestWmi_DoQuery(t *testing.T) {
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, compileInputs(&plugin))
	for _, q := range plugin.Queries {
		require.NoError(t, q.doQuery(acc))
	}
	// no errors in accumulator
	require.Len(t, acc.Errors, 0, "found errors accumulated by AddError()")
	// Only one metric was returned (because we filtered for SystemDrive)
	require.Len(t, acc.Metrics, 1)
	// Name property collected and is the SystemDrive
	require.Equal(t, sysDrive, acc.Metrics[0].Tags["Name"])
	// FreeSpace property was collected as a field
	require.NotEmpty(t, acc.Metrics[0].Fields["FreeSpace"])
}

// test Init function
func TestWmi_Init(t *testing.T) {
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, plugin.Init())
}

// test Gather function
func TestWmi_Gather(t *testing.T) {
	plugin := Wmi{Queries: []Query{testQuery}, Log: logger}
	require.NoError(t, plugin.Gather(acc))
}
