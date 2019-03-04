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
		"subdir/nested2", "subdir/nested2/qux"}
	fileCountEquals(t, fc, len(matches), 9084)
}

func TestNoFiltersOnChildDir(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Directories = []string{getTestdataDir() + "/*"}
	matches := []string{"subdir/quux", "subdir/quuz",
		"subdir/nested2/qux", "subdir/nested2"}

	tags := map[string]string{"directory": getTestdataDir() + "/subdir"}
	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(4542)))
}

func TestNoRecursiveButSuperMeta(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Recursive = false
	fc.Directories = []string{getTestdataDir() + "/**"}
	matches := []string{"subdir/quux", "subdir/quuz", "subdir/nested2"}

	tags := map[string]string{"directory": getTestdataDir() + "/subdir"}
	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(4096)))
}

func TestNameFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Name = "ba*"
	matches := []string{"bar", "baz"}
	fileCountEquals(t, fc, len(matches), 0)
}

func TestNonRecursive(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Recursive = false
	matches := []string{"foo", "bar", "baz", "qux", "subdir"}
	fileCountEquals(t, fc, len(matches), 4542)
}

func TestDoubleAndSimpleStar(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Directories = []string{getTestdataDir() + "/**/*"}
	matches := []string{"qux"}
	tags := map[string]string{"directory": getTestdataDir() + "/subdir/nested2"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(446)))
}

func TestRegularOnlyFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.RegularOnly = true
	matches := []string{
		"foo", "bar", "baz", "qux", "subdir/quux", "subdir/quuz",
		"subdir/nested2/qux"}
	fileCountEquals(t, fc, len(matches), 892)
}

func TestSizeFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Size = internal.Size{Size: -100}
	matches := []string{"foo", "bar", "baz",
		"subdir/quux", "subdir/quuz"}
	fileCountEquals(t, fc, len(matches), 0)

	fc.Size = internal.Size{Size: 100}
	matches = []string{"qux", "subdir/nested2//qux"}
	fileCountEquals(t, fc, len(matches), 892)
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
		"sbudir/nested2", "subdir/nested2/qux"}
	fileCountEquals(t, fc, len(matches), 9084)

	fc.MTime = internal.Duration{Duration: fileAge}
	matches = []string{"baz"}
	fileCountEquals(t, fc, len(matches), 0)
}

func getNoFilterFileCount() FileCount {
	return FileCount{
		Directories: []string{getTestdataDir()},
		Name:        "*",
		Recursive:   true,
		RegularOnly: false,
		Size:        internal.Size{Size: 0},
		MTime:       internal.Duration{Duration: 0},
		fileFilters: nil,
	}
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filecount_test.go", "testdata", 1)
}

func fileCountEquals(t *testing.T, fc FileCount, expectedCount int, expectedSize int) {
	tags := map[string]string{"directory": getTestdataDir()}
	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)
	require.True(t, acc.HasPoint("filecount", tags, "count", int64(expectedCount)))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(expectedSize)))
}
