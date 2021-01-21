// +build !windows

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
	// test super asterisk
	g1, err := Compile(filepath.Join(testdataDir, "**"))
	require.NoError(t, err)
	// test single asterisk
	g2, err := Compile(filepath.Join(testdataDir, "*.log"))
	require.NoError(t, err)
	// test no meta characters (file exists)
	g3, err := Compile(filepath.Join(testdataDir, "log1.log"))
	require.NoError(t, err)
	// test file that doesn't exist
	g4, err := Compile(filepath.Join(testdataDir, "i_dont_exist.log"))
	require.NoError(t, err)
	// test super asterisk that doesn't exist
	g5, err := Compile(filepath.Join(testdataDir, "dir_doesnt_exist", "**"))
	require.NoError(t, err)

	matches := g1.Match()
	require.Len(t, matches, 6)
	matches = g2.Match()
	require.Len(t, matches, 2)
	matches = g3.Match()
	require.Len(t, matches, 1)
	matches = g4.Match()
	require.Len(t, matches, 1)
	matches = g5.Match()
	require.Len(t, matches, 0)
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
		actual, _ := Compile(test.input)
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
		{"/root/foo", []string{"/root/foo"}},
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
