package globpath

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompileAndMatch(t *testing.T) {
	dir := getTestdataDir()
	// test super asterisk
	g1, err := Compile(dir + "/**")
	require.NoError(t, err)
	// test single asterisk
	g2, err := Compile(dir + "/*.log")
	require.NoError(t, err)
	// test no meta characters (file exists)
	g3, err := Compile(dir + "/log1.log")
	require.NoError(t, err)
	// test file that doesn't exist
	g4, err := Compile(dir + "/i_dont_exist.log")
	require.NoError(t, err)
	// test super asterisk that doesn't exist
	g5, err := Compile(dir + "/dir_doesnt_exist/**")
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
	dir := getTestdataDir()
	tests := []struct {
		input  string
		output string
	}{
		{dir + "/**", dir + "/*"},
		{dir + "/nested?/**", dir + "/nested?/*"},
		{dir + "/ne**/nest*", dir + "/ne*"},
		{dir + "/nested?/*", ""},
	}

	for _, test := range tests {
		actual, _ := Compile(test.input)
		require.Equal(t, actual.rootGlob, test.output)
	}
}

func TestFindNestedTextFile(t *testing.T) {
	dir := getTestdataDir()
	// test super asterisk
	g1, err := Compile(dir + "/**.txt")
	require.NoError(t, err)

	matches := g1.Match()
	require.Len(t, matches, 1)
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "globpath_test.go", "testdata", 1)
}

func TestMatch_ErrPermission(t *testing.T) {
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
