//go:build !windows

// TODO: Windows - should be enabled for Windows when super asterisk is fixed on Windows
// https://github.com/influxdata/telegraf/issues/6248

package globpath

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testdataDir = getTestdataDir()
)

func TestCompileAndMatch(t *testing.T) {
	type test struct {
		path    string
		matches int
	}

	tests := []test{
		//test super asterisk
		{path: filepath.Join(testdataDir, "**"), matches: 7},
		// test single asterisk
		{path: filepath.Join(testdataDir, "*.log"), matches: 3},
		// test no meta characters (file exists)
		{path: filepath.Join(testdataDir, "log1.log"), matches: 1},
		// test file that doesn't exist
		{path: filepath.Join(testdataDir, "i_dont_exist.log"), matches: 0},
		// test super asterisk that doesn't exist
		{path: filepath.Join(testdataDir, "dir_doesnt_exist", "**"), matches: 0},
		// test exclamation mark creates non-matching list with a range
		{path: filepath.Join(testdataDir, "log[!1-2]*"), matches: 1},
		// test caret creates non-matching list
		{path: filepath.Join(testdataDir, "log[^1-2]*"), matches: 1},
		// test exclamation mark creates non-matching list without a range
		{path: filepath.Join(testdataDir, "log[!2]*"), matches: 2},
		// test exclamation mark creates non-matching list without a range
		//nolint:gocritic // filepathJoin - '\\' used to escape in glob, not path separator
		{path: filepath.Join(testdataDir, "log\\[!*"), matches: 1},
		// test exclamation mark creates non-matching list without a range
		//nolint:gocritic // filepathJoin - '\\' used to escape in glob, not path separator
		{path: filepath.Join(testdataDir, "log\\[^*"), matches: 0},
	}

	for _, tc := range tests {
		g, err := Compile(tc.path)
		require.NoError(t, err)
		matches := g.Match()
		require.Len(t, matches, tc.matches)
	}
}

func TestRootGlob(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{filepath.Join(testdataDir, "**"), filepath.Join(testdataDir, "*")},
		{filepath.Join(testdataDir, "nested?", "**"), filepath.Join(testdataDir, "nested?", "*")},
		{filepath.Join(testdataDir, "ne**", "nest*"), filepath.Join(testdataDir, "ne*")},
		{filepath.Join(testdataDir, "nested?", "*"), ""},
	}

	for _, test := range tests {
		actual, err := Compile(test.input)
		require.NoError(t, err)
		require.Equal(t, actual.rootGlob, test.output)
	}
}

func TestFindNestedTextFile(t *testing.T) {
	// test super asterisk
	g1, err := Compile(filepath.Join(testdataDir, "**.txt"))
	require.NoError(t, err)

	matches := g1.Match()
	require.Len(t, matches, 1)
}

func TestMatch_ErrPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix only test")
	}

	tests := []struct {
		input    string
		expected []string
	}{
		{"/root/foo", []string(nil)},
		{"/root/f*", []string(nil)},
	}

	for _, test := range tests {
		glob, err := Compile(test.input)
		require.NoError(t, err)
		actual := glob.Match()
		require.Equal(t, test.expected, actual)
	}
}

func TestWindowsSeparator(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows only test")
	}

	glob, err := Compile("testdata/nested1")
	require.NoError(t, err)
	ok := glob.MatchString("testdata\\nested1")
	require.True(t, ok)
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
