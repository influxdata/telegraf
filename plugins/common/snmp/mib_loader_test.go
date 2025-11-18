package snmp

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

type TestingMibLoader struct {
	folders []string
	files   []string
}

func (t *TestingMibLoader) appendPath(path string) {
	t.folders = append(t.folders, path)
}

func (t *TestingMibLoader) loadModule(path string) error {
	t.files = append(t.files, path)
	return nil
}
func TestFolderLookup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on windows")
	}

	tests := []struct {
		name    string
		mibPath [][]string
		paths   [][]string
		files   []string
	}{
		{
			name:    "loading folders",
			mibPath: [][]string{{"testdata", "loadMibsFromPath", "root"}},
			paths: [][]string{
				{"testdata", "loadMibsFromPath", "root"},
				{"testdata", "loadMibsFromPath", "root", "dirOne"},
				{"testdata", "loadMibsFromPath", "root", "dirOne", "dirTwo"},
				{"testdata", "loadMibsFromPath", "linkTarget"},
			},
			files: []string{"empty", "emptyFile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := TestingMibLoader{}

			var givenPath []string
			for _, paths := range tt.mibPath {
				rootPath := filepath.Join(paths...)
				givenPath = append(givenPath, rootPath)
			}

			err := LoadMibsFromPath(givenPath, testutil.Logger{}, &loader)
			require.NoError(t, err)

			var folders []string
			for _, pathSlice := range tt.paths {
				path := filepath.Join(pathSlice...)
				folders = append(folders, path)
			}
			require.Equal(t, folders, loader.folders)

			require.Equal(t, tt.files, loader.files)
		})
	}
}

func TestMissingMibPath(t *testing.T) {
	log := testutil.Logger{}
	path := []string{"non-existing-directory"}
	require.NoError(t, LoadMibsFromPath(path, log, &GosmiMibLoader{}))
}

func BenchmarkMibLoading(b *testing.B) {
	log := testutil.Logger{}
	path := []string{"testdata/gosmi"}
	for i := 0; i < b.N; i++ {
		require.NoError(b, LoadMibsFromPath(path, log, &GosmiMibLoader{}))
	}
}
