//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package filestat

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

var (
	testdataDir = getTestdataDir()
)

func TestGatherNoMd5(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Files = []string{
		filepath.Join(testdataDir, "log1.log"),
		filepath.Join(testdataDir, "log2.log"),
		filepath.Join(testdataDir, "non_existent_file"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "log1.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))

	tags2 := map[string]string{
		"file": filepath.Join(testdataDir, "log2.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))

	tags3 := map[string]string{
		"file": filepath.Join(testdataDir, "non_existent_file"),
	}
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(0)))
}

func TestGatherExplicitFiles(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Md5 = true
	fs.Files = []string{
		filepath.Join(testdataDir, "log1.log"),
		filepath.Join(testdataDir, "log2.log"),
		filepath.Join(testdataDir, "non_existent_file"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "log1.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": filepath.Join(testdataDir, "log2.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags3 := map[string]string{
		"file": filepath.Join(testdataDir, "non_existent_file"),
	}
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(0)))
}

func TestNonExistentFile(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Md5 = true
	fs.Files = []string{
		"/non/existant/file",
	}
	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	acc.AssertContainsFields(t, "filestat", map[string]interface{}{"exists": int64(0)})
	require.False(t, acc.HasField("filestat", "error"))
	require.False(t, acc.HasField("filestat", "md5_sum"))
	require.False(t, acc.HasField("filestat", "size_bytes"))
	require.False(t, acc.HasField("filestat", "modification_time"))
}

func TestGatherGlob(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Md5 = true
	fs.Files = []string{
		filepath.Join(testdataDir, "*.log"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "log1.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": filepath.Join(testdataDir, "log2.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))
}

func TestGatherSuperAsterisk(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Md5 = true
	fs.Files = []string{
		filepath.Join(testdataDir, "**"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "log1.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags1, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags2 := map[string]string{
		"file": filepath.Join(testdataDir, "log2.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags2, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags2, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags2, "md5_sum", "d41d8cd98f00b204e9800998ecf8427e"))

	tags3 := map[string]string{
		"file": filepath.Join(testdataDir, "test.conf"),
	}
	require.True(t, acc.HasPoint("filestat", tags3, "size_bytes", int64(104)))
	require.True(t, acc.HasPoint("filestat", tags3, "exists", int64(1)))
	require.True(t, acc.HasPoint("filestat", tags3, "md5_sum", "5a7e9b77fa25e7bb411dbd17cf403c1f"))
}

func TestModificationTime(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Files = []string{
		filepath.Join(testdataDir, "log1.log"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "log1.log"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "size_bytes", int64(0)))
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(1)))
	require.True(t, acc.HasInt64Field("filestat", "modification_time"))
}

func TestNoModificationTime(t *testing.T) {
	fs := NewFileStat()
	fs.Log = testutil.Logger{}
	fs.Files = []string{
		filepath.Join(testdataDir, "non_existent_file"),
	}

	acc := testutil.Accumulator{}
	require.NoError(t, acc.GatherError(fs.Gather))

	tags1 := map[string]string{
		"file": filepath.Join(testdataDir, "non_existent_file"),
	}
	require.True(t, acc.HasPoint("filestat", tags1, "exists", int64(0)))
	require.False(t, acc.HasInt64Field("filestat", "modification_time"))
}

func TestGetMd5(t *testing.T) {
	md5, err := getMd5(filepath.Join(testdataDir, "test.conf"))
	require.NoError(t, err)
	require.Equal(t, "5a7e9b77fa25e7bb411dbd17cf403c1f", md5)

	_, err = getMd5("/tmp/foo/bar/fooooo")
	require.Error(t, err)
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
