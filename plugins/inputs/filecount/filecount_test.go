//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package filecount

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNoFilters(t *testing.T) {
	fc := getNoFilterFileCount()
	matches := []string{"foo", "bar", "baz", "qux",
		"subdir/", "subdir/quux", "subdir/quuz",
		"subdir/nested2", "subdir/nested2/qux"}
	fileCountEquals(t, fc, len(matches), 5096)
}

func TestNoFiltersOnChildDir(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Directories = []string{getTestdataDir() + "/*"}
	matches := []string{"subdir/quux", "subdir/quuz",
		"subdir/nested2/qux", "subdir/nested2"}

	tags := map[string]string{"directory": getTestdataDir() + "/subdir"}
	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fc.Gather))
	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(600)))
}

func TestNoRecursiveButSuperMeta(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Recursive = false
	fc.Directories = []string{getTestdataDir() + "/**"}
	matches := []string{"subdir/quux", "subdir/quuz", "subdir/nested2"}

	tags := map[string]string{"directory": getTestdataDir() + "/subdir"}
	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fc.Gather))

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(200)))
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

	fileCountEquals(t, fc, len(matches), 4496)
}

func TestDoubleAndSimpleStar(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Directories = []string{getTestdataDir() + "/**/*"}
	matches := []string{"qux"}

	tags := map[string]string{"directory": getTestdataDir() + "/subdir/nested2"}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fc.Gather))

	require.True(t, acc.HasPoint("filecount", tags, "count", int64(len(matches))))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(400)))
}

func TestRegularOnlyFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.RegularOnly = true
	matches := []string{
		"foo", "bar", "baz", "qux", "subdir/quux", "subdir/quuz",
		"subdir/nested2/qux"}

	fileCountEquals(t, fc, len(matches), 800)
}

func TestSizeFilter(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.Size = config.Size(-100)
	matches := []string{"foo", "bar", "baz",
		"subdir/quux", "subdir/quuz"}
	fileCountEquals(t, fc, len(matches), 0)

	fc.Size = config.Size(100)
	matches = []string{"qux", "subdir/nested2//qux"}

	fileCountEquals(t, fc, len(matches), 800)
}

func TestMTimeFilter(t *testing.T) {
	mtime := time.Date(2011, time.December, 14, 18, 25, 5, 0, time.UTC)
	fileAge := time.Since(mtime) - (60 * time.Second)

	fc := getNoFilterFileCount()
	fc.MTime = config.Duration(-fileAge)
	matches := []string{"foo", "bar", "qux",
		"subdir/", "subdir/quux", "subdir/quuz",
		"subdir/nested2", "subdir/nested2/qux"}

	fileCountEquals(t, fc, len(matches), 5096)

	fc.MTime = config.Duration(fileAge)
	matches = []string{"baz"}
	fileCountEquals(t, fc, len(matches), 0)
}

// The library dependency karrick/godirwalk completely abstracts out the
// behavior of the FollowSymlinks plugin input option. However, it should at
// least behave identically when enabled on a filesystem with no symlinks.
func TestFollowSymlinks(t *testing.T) {
	fc := getNoFilterFileCount()
	fc.FollowSymlinks = true
	matches := []string{"foo", "bar", "baz", "qux",
		"subdir/", "subdir/quux", "subdir/quuz",
		"subdir/nested2", "subdir/nested2/qux"}

	fileCountEquals(t, fc, len(matches), 5096)
}

// Paths with a trailing slash will not exactly match paths produced during the
// walk as these paths are cleaned before being returned from godirwalk. #6329
func TestDirectoryWithTrailingSlash(t *testing.T) {
	plugin := &FileCount{
		Directories: []string{getTestdataDir() + string(filepath.Separator)},
		Name:        "*",
		Recursive:   true,
		Fs:          getFakeFileSystem(getTestdataDir()),
	}

	var acc testutil.Accumulator
	err := plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"filecount",
			map[string]string{
				"directory": getTestdataDir(),
			},
			map[string]interface{}{
				"count":      9,
				"size_bytes": 5096,
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func getNoFilterFileCount() FileCount {
	return FileCount{
		Log:         testutil.Logger{},
		Directories: []string{getTestdataDir()},
		Name:        "*",
		Recursive:   true,
		RegularOnly: false,
		Size:        config.Size(0),
		MTime:       config.Duration(0),
		fileFilters: nil,
		Fs:          getFakeFileSystem(getTestdataDir()),
	}
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	var chunks []string
	var testDirectory string

	if runtime.GOOS == "windows" {
		chunks = strings.Split(dir, "\\")
		testDirectory = strings.Join(chunks[:], "\\") + "\\testdata"
	} else {
		chunks = strings.Split(dir, "/")
		testDirectory = strings.Join(chunks[:], "/") + "/testdata"
	}
	return testDirectory
}

func getFakeFileSystem(basePath string) fakeFileSystem {
	// create our desired "filesystem" object, complete with an internal map allowing our funcs to return meta data as requested

	mtime := time.Date(2015, time.December, 14, 18, 25, 5, 0, time.UTC)
	olderMtime := time.Date(2010, time.December, 14, 18, 25, 5, 0, time.UTC)

	// set file permissions
	var fmask uint32 = 0666
	var dmask uint32 = 0666

	// set directory bit
	dmask |= 1 << uint(32-1)

	// create a lookup map for getting "files" from the "filesystem"
	fileList := map[string]fakeFileInfo{
		basePath:                         {name: "testdata", size: int64(4096), filemode: dmask, modtime: mtime, isdir: true},
		basePath + "/foo":                {name: "foo", filemode: fmask, modtime: mtime},
		basePath + "/bar":                {name: "bar", filemode: fmask, modtime: mtime},
		basePath + "/baz":                {name: "baz", filemode: fmask, modtime: olderMtime},
		basePath + "/qux":                {name: "qux", size: int64(400), filemode: fmask, modtime: mtime},
		basePath + "/subdir":             {name: "subdir", size: int64(4096), filemode: dmask, modtime: mtime, isdir: true},
		basePath + "/subdir/quux":        {name: "quux", filemode: fmask, modtime: mtime},
		basePath + "/subdir/quuz":        {name: "quuz", filemode: fmask, modtime: mtime},
		basePath + "/subdir/nested2":     {name: "nested2", size: int64(200), filemode: dmask, modtime: mtime, isdir: true},
		basePath + "/subdir/nested2/qux": {name: "qux", filemode: fmask, modtime: mtime, size: int64(400)},
	}

	return fakeFileSystem{files: fileList}
}

func fileCountEquals(t *testing.T, fc FileCount, expectedCount int, expectedSize int) {
	tags := map[string]string{"directory": getTestdataDir()}
	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fc.Gather))
	require.True(t, acc.HasPoint("filecount", tags, "count", int64(expectedCount)))
	require.True(t, acc.HasPoint("filecount", tags, "size_bytes", int64(expectedSize)))
}
