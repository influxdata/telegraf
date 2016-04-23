package globpath

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileAndMatch(t *testing.T) {
	dir := getTestdataDir()
	g1, err := Compile(dir + "/**")
	require.NoError(t, err)
	g2, err := Compile(dir + "/*.log")
	require.NoError(t, err)
	g3, err := Compile(dir + "/log1.log")
	require.NoError(t, err)

	matches := g1.Match()
	assert.Len(t, matches, 3)
	matches = g2.Match()
	assert.Len(t, matches, 2)
	matches = g3.Match()
	assert.Len(t, matches, 1)
}

func TestFindRootDir(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{"/var/log/telegraf.conf", "/var/log"},
		{"/home/**", "/home"},
		{"/home/*/**", "/home"},
		{"/lib/share/*/*/**.txt", "/lib/share"},
	}

	for _, test := range tests {
		actual := findRootDir(test.input)
		assert.Equal(t, test.output, actual)
	}
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "globpath_test.go", "testdata", 1)
}
