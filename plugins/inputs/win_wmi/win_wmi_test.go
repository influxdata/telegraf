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
var testWmi Wmi = Wmi{Queries: []Query{testQuery}, Log: logger}

// test DoQuery
func TestWmi_DoQuery(t *testing.T) {
	require.NoError(t, CompileInputs(&testWmi))
	for _, q := range testWmi.Queries {
		require.NoError(t, q.DoQuery(acc))
	}
	// no errors in accumulator
	require.Len(t, acc.Errors, 0, "found errors accumulated by AddError()")
	// Name property collected and is the SystemDrive
	require.Equal(t, sysDrive, acc.Metrics[0].Tags["Name"])
	// FreeSpace property was collected as a field
	require.NotEmpty(t, acc.Metrics[0].Fields["FreeSpace"])
	// Only one metric was returned (because we filtered for SystemDrive)
	require.Equal(t, len(acc.Metrics), 1)
	// CompileInputs() built the correct WQL
	require.Equal(t, expectedWql, testWmi.Queries[0].query)
}

// test Init function
func TestWmi_Init(t *testing.T) {
	t.Run("NoError", func(t *testing.T) { require.NoError(t, testWmi.Init()) })
}

// test Gather function
func TestWmi_Gather(t *testing.T) {
	t.Run("NoError", func(t *testing.T) { require.NoError(t, testWmi.Gather(acc)) })
}
