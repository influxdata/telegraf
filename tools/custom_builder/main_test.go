package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCases(t *testing.T) {
	// Silence the output
	log.SetOutput(io.Discard)

	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		configFilename := filepath.Join("testcases", f.Name(), "telegraf.conf")
		expecedFilename := filepath.Join("testcases", f.Name(), "expected.tags")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected output
			file, err := os.Open(expecedFilename)
			require.NoError(t, err)
			defer file.Close()

			var expected []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				expected = append(expected, scanner.Text())
			}
			require.NoError(t, scanner.Err())

			// Configure the command
			cfg := &cmdConfig{
				dryrun:      true,
				quiet:       true,
				configFiles: []string{configFilename},
				root:        "../..",
			}

			actual, err := process(cfg)
			require.NoError(t, err)
			require.EqualValues(t, expected, actual)
		})
	}
}
