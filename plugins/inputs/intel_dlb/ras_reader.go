//go:build linux
// +build linux

package intel_dlb

import (
	"fmt"
	"os"
	"path/filepath"
)

type rasReader interface {
	gatherPaths(path string) ([]string, error)
	readFromFile(filePath string) ([]byte, error)
}

type rasReaderImpl struct {
}

// gatherPaths gathers all paths based on provided pattern
func (rasReaderImpl) gatherPaths(pattern string) ([]string, error) {
	filePaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob failed for pattern: %s: %v", pattern, err)
	}

	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no candidates for given pattern: %s", pattern)
	}

	return filePaths, nil
}

// readFromFile reads file content.
func (rasReaderImpl) readFromFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
