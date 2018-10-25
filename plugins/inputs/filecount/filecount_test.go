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
	fc := getNoFilterFileCount("*")
	matches := []string{"foo", "bar", "baz", "qux", "subdir/", "subdir/quux", "subdir/quuz"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func TestNoFiltersOnChildDir(t *testing.T) {
	fc := getNoFilterFileCount("testdata/*")
	matches := []string{"subdir/quux", "subdir/quuz"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata/subdir", len(matches)))
}

func TestNameFilter(t *testing.T) {
	fc := getNoFilterFileCount("testdata")
	fc.Name = "ba*"
	matches := []string{"bar", "baz"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func TestNonRecursive(t *testing.T) {
	fc := getNoFilterFileCount("testdata")
	fc.Recursive = false
	matches := []string{"foo", "bar", "baz", "qux", "subdir"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func TestRegularOnlyFilter(t *testing.T) {
	fc := getNoFilterFileCount("testdata")
	fc.RegularOnly = true
	matches := []string{
		"foo", "bar", "baz", "qux", "subdir/quux", "subdir/quuz",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func TestSizeFilter(t *testing.T) {
	fc := getNoFilterFileCount("testdata")
	fc.Size = internal.Size{Size: -100}
	matches := []string{"foo", "bar", "baz", "subdir/quux", "subdir/quuz"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))

	fc.Size = internal.Size{Size: 100}
	matches = []string{"qux"}

	acc = testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func TestMTimeFilter(t *testing.T) {
	oldFile := filepath.Join(getTestdataDir("testdata"), "baz")
	mtime := time.Date(1979, time.December, 14, 18, 25, 5, 0, time.UTC)
	if err := os.Chtimes(oldFile, mtime, mtime); err != nil {
		t.Skip("skipping mtime filter test.")
	}
	fileAge := time.Since(mtime) - (60 * time.Second)

	fc := getNoFilterFileCount("testdata")
	fc.MTime = internal.Duration{Duration: -fileAge}
	matches := []string{"foo", "bar", "qux", "subdir/", "subdir/quux", "subdir/quuz"}

	acc := testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))

	fc.MTime = internal.Duration{Duration: fileAge}
	matches = []string{"baz"}

	acc = testutil.Accumulator{}
	acc.GatherError(fc.Gather)

	require.True(t, assertFileCount(&acc, "testdata", len(matches)))
}

func getNoFilterFileCount(dir string) FileCount {
	return FileCount{
		Directories: []string{getTestdataDir(dir)},
		Name:        "*",
		Recursive:   true,
		RegularOnly: false,
		Size:        internal.Size{Size: 0},
		MTime:       internal.Duration{Duration: 0},
		fileFilters: nil,
	}
}

func getTestdataDir(dir string) string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filecount_test.go", dir, 1)
}

func assertFileCount(acc *testutil.Accumulator, expectedDir string, expectedCount int) bool {
	tags := map[string]string{"directory": getTestdataDir(expectedDir)}
	return acc.HasPoint("filecount", tags, "count", int64(expectedCount))
}
