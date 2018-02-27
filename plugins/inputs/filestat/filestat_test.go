package filestat

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))

	tags3 := map[string]string{
		"file": "/non/existant/file",
	}
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(0)))
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
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags3 := map[string]string{
		"file": "/non/existant/file",
	}
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(0)))
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
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))
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
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": dir + "log2.log",
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags3 := map[string]string{
		"file": dir + "test.conf",
	}
	require.True(t, acc.HasPoint("filestat", tags3, "size_bytes", int64(104)))
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags3, "md5_sum", "5a7e9b77fa25e7bb411dbd17cf403c1f"))
}

func TestModificationTime(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Files = []string{
		dir + "log1.log",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": dir + "log1.log",
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasInt64Field("filestat", "modification_time"))
}

func TestNoModificationTime(t *testing.T) {
	fs := NewFileStat()
	fs.Files = []string{
		"/non/existant/file",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	tags1 := map[string]string{
		"file": "/non/existant/file",
	}
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(0)))
	require.False(t, acc.HasInt64Field("filestat", "modification_time"))
}

func TestGetMd5(t *testing.T) {
	dir := getTestdataDir()
	md5, err := getMd5(dir + "test.conf")
	assert.NoError(t, err)
	assert.Equal(t, "5a7e9b77fa25e7bb411dbd17cf403c1f", md5)

	md5, err = getMd5("/tmp/foo/bar/fooooo")
	assert.Error(t, err)
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filestat_test.go", "testdata/", 1)
}
