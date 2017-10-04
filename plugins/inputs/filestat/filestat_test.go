package filestat

import (
	"runtime"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGatherNoMd5(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Files = []string{
		dir + "log1.log",
		dir + "log2.log",
		"/non/existant/file",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": dir + "log1.log",
	}
	fields1 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags1["file"])),
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields1, tags1)

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	fields2 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags2["file"])),
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields2, tags2)

	tags3 := map[string]string{
		"file": "/non/existant/file",
	}
	fields3 := map[string]interface{}{
		"exists": int64(0),
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields3, tags3)
}

func TestGatherExplicitFiles(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "log1.log",
		dir + "log2.log",
		"/non/existant/file",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": dir + "log1.log",
	}
	fields1 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags1["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields1, tags1)

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	fields2 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags2["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields2, tags2)

	tags3 := map[string]string{
		"file": "/non/existant/file",
	}
	fields3 := map[string]interface{}{
		"exists": int64(0),
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields3, tags3)
}

func TestGatherGlob(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "*.log",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": dir + "log1.log",
	}
	fields1 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags1["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields1, tags1)

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	fields2 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags2["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields2, tags2)
}

func TestGatherSuperAsterisk(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "**",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": dir + "log1.log",
	}
	fields1 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags1["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields1, tags1)

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	fields2 := map[string]interface{}{
		"size_bytes":        int64(0),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags2["file"])),
		"md5_sum":           "d41d8cd98f00b204e9800998ecf8427e",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields2, tags2)

	tags3 := map[string]string{
		"file": dir + "test.conf",
	}
	fields3 := map[string]interface{}{
		"size_bytes":        int64(104),
		"exists":            int64(1),
		"modification_time": int64(getModificationTime(fs, tags3["file"])),
		"md5_sum":           "5a7e9b77fa25e7bb411dbd17cf403c1f",
	}
	acc.AssertContainsTaggedFields(t, "filestat", fields3, tags3)
}

func TestGetMd5(t *testing.T) {
	dir := getTestdataDir()
	md5, err := getMd5(dir + "test.conf")
	assert.NoError(t, err)
	assert.Equal(t, "5a7e9b77fa25e7bb411dbd17cf403c1f", md5)

	md5, err = getMd5("/tmp/foo/bar/fooooo")
	assert.Error(t, err)
}

func getModificationTime(f *FileStat, filepath string) int64 {
	// This function is a near copy from the main code,
	// however since the modification time of test files will change based on code pull this seems the only way to solve unit tests
	// All unit tests should call this function for expected value of modification_time

	var err error
	g, ok := f.globs[filepath]
	if !ok {
		if g, err = globpath.Compile(filepath); err != nil {
			return 0
		}
		f.globs[filepath] = g
	}

	files := g.Match()
	if len(files) == 0 {
		return 0
	}

	for _, fileInfo := range files {
		if fileInfo == nil {
			return 0
		} else {
			return fileInfo.ModTime().Unix()
		}
	}

	return 0
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filestat_test.go", "testdata/", 1)
}
