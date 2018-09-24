package filecount

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNoFilters(t *testing.T) {
	fc := getNoFilterFileCount()
	matches := []string{"foo", "bar", "baz", "qux",
		"subdir/", "subdir/quux", "subdir/quuz",
		"subdir/qux"}
	require.True(t, fileCountEquals(fc, len(matches), 4988))
}

func TestNameFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Name = "ba*"
	matches := []string{"bar", "baz"}
	require.True(t, fileCountEquals(fc, len(matches), 0))
}

func TestNonRecursive(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Recursive = false
	matches := []string{"foo", "bar", "baz", "qux", "subdir"}
	require.True(t, fileCountEquals(fc, len(matches), 4542))
}

func TestRegularOnlyFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.RegularOnly = true
	matches := []string{
		"foo", "bar", "baz", "qux", "subdir/quux", "subdir/quuz",
		"subdir/qux"}
	require.True(t, fileCountEquals(fc, len(matches), 892))
}

func TestSizeFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Size = "-100B"
	matches := []string{"foo", "bar", "baz",
		"subdir/quux", "subdir/quuz"}
	require.True(t, fileCountEquals(fc, len(matches), 0))

	fc.Size = "100B"
	matches = []string{"qux", "subdir/qux"}
	require.True(t, fileCountEquals(fc, len(matches), 892))
}

func TestMTimeFilter(t *testing.T) {
	oldFile := filepath.Join(getTestdataDir(), "baz")
	mtime := time.Date(1979, time.December, 14, 18, 25, 5, 0, time.UTC)
	if err := os.Chtimes(oldFile, mtime, mtime); err != nil {
		t.Skip("skipping mtime filter test.")
	}
	fileAge := time.Since(mtime) - (60 * time.Second)

	fc := getNoFilterFileCount()
	fc.MTime = internal.Duration{Duration: -fileAge}
	matches := []string{"foo", "bar", "qux",
		"subdir/", "subdir/quux", "subdir/quuz",
		"subdir/qux"}
	require.True(t, fileCountEquals(fc, len(matches), 4988))

	fc.MTime = internal.Duration{Duration: fileAge}
	matches = []string{"baz"}
	require.True(t, fileCountEquals(fc, len(matches), 0))
}

func TestRecursivePrint(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.RecursivePrint = true;

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)
	tags := map[string]string{"directory": getTestdataDir() + string(os.PathSeparator) + "subdir"}

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(3)))
	require.True(t, acc.HasPoint("filecount", tags, "size", int64(446)))
}

func getNoFilterFileCount() FileCount {
	return FileCount{
		Directory:          getTestdataDir(),
		CountSize:          true,
		Name:               "*",
		Recursive:          true,
		RegularOnly:        false,
		Size:               "0B",
		MTime:              internal.Duration{Duration: 0},
		RecursivePrint:     false,
		RecursivePrintSize: "0B",
		fileFilters:        nil,
	}
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filecount_test.go", "testdata/", 1)
}

func fileCountEquals(fc FileCount, expectedCount int, expectedSize int) bool {
	tags := map[string]string{"directory": getTestdataDir()}
	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)
	return acc.HasPoint("filecount", tags, "count", int64(expectedCount)) &&
	 acc.HasPoint("filecount", tags, "size", int64(expectedSize))
}
