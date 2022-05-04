//go:build !windows
// +build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package filecount

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMTime(t *testing.T) {
	//this is the time our foo file should have
	mtime := time.Date(2015, time.December, 14, 18, 25, 5, 0, time.UTC)

	fs := getTestFileSystem()
	fileInfo, err := fs.Stat("/testdata/foo")
	require.NoError(t, err)
	require.Equal(t, mtime, fileInfo.ModTime())
}

func TestSize(t *testing.T) {
	//this is the time our foo file should have
	size := int64(4096)
	fs := getTestFileSystem()
	fileInfo, err := fs.Stat("/testdata")
	require.NoError(t, err)
	require.Equal(t, size, fileInfo.Size())
}

func TestIsDir(t *testing.T) {
	//this is the time our foo file should have
	dir := true
	fs := getTestFileSystem()
	fileInfo, err := fs.Stat("/testdata")
	require.NoError(t, err)
	require.Equal(t, dir, fileInfo.IsDir())
}

func TestRealFS(t *testing.T) {
	//test that the default (non-test) empty FS causes expected behaviour
	var fs fileSystem = osFS{}
	//the following file exists on disk - and not in our fake fs
	fileInfo, err := fs.Stat(getTestdataDir() + "/qux")
	require.NoError(t, err)
	require.Equal(t, false, fileInfo.IsDir())
	require.Equal(t, int64(446), fileInfo.Size())

	// now swap out real, for fake filesystem
	fs = getTestFileSystem()
	// now, the same test as above will return an error as the file doesn't exist in our fake fs
	expectedError := "Stat " + getTestdataDir() + "/qux: No such file or directory"
	_, err = fs.Stat(getTestdataDir() + "/qux")
	require.Error(t, err, expectedError)
	// and verify that what we DO expect to find, we do
	fileInfo, err = fs.Stat("/testdata/foo")
	require.NoError(t, err)
	require.NotNil(t, fileInfo)
}

func getTestFileSystem() fakeFileSystem {
	/*
		create our desired "filesystem" object, complete with an internal map allowing our funcs to return meta data as requested

		type FileInfo interface {
			Name() string       // base name of the file
			Size() int64        // length in bytes of file
			Mode() FileMode     // file mode bits
			ModTime() time.Time // modification time
			IsDir() bool        // returns bool indicating if a Dir or not
			Sys() interface{}   // underlying data source. always nil (in this case)
		}

	*/

	mtime := time.Date(2015, time.December, 14, 18, 25, 5, 0, time.UTC)

	// set file permissions
	var fmask uint32 = 0666
	var dmask uint32 = 0666

	// set directory bit
	dmask |= 1 << uint(32-1)

	fileList := map[string]fakeFileInfo{
		"/testdata":     {name: "testdata", size: int64(4096), filemode: dmask, modtime: mtime, isdir: true},
		"/testdata/foo": {name: "foo", filemode: fmask, modtime: mtime},
	}

	return fakeFileSystem{files: fileList}
}
