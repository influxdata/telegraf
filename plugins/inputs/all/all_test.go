package all

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AllPlugins(t *testing.T) {
	filepath.WalkDir("./", func(path string, d fs.DirEntry, err1 error) error {
		if d.IsDir() || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, d.Name(), nil, parser.ParseComments)
		require.NoError(t, err)
		for _, cg := range node.Comments {
			for _, comm := range cg.List {
				if !strings.HasPrefix(comm.Text, "//go:build") {
					continue
				}
				t.Run(d.Name(), func(t *testing.T) {
					tags := strings.Split(comm.Text, "||")
					// tags might contain spaces and hence trim
					tags = func(elems []string, transFormFunc func(string) string) []string {
						result := make([]string, len(elems))
						for i, t := range elems {
							result[i] = strings.TrimPrefix(transFormFunc(t), "//go:build ")
						}
						return result
					}(tags, strings.TrimSpace)

					assert.Len(t, tags, 3)
					assert.Contains(t, tags, "!custom")
					assert.Contains(t, tags, "inputs")
					plugin := getPlugin(tags)
					assert.Greater(t, len(plugin), 0)

					// should contain only one import statement
					require.Equal(t, 1, len(node.Imports))

					// trim the path surrounded by quotes
					importPath := strings.Trim(node.Imports[0].Path.Value, "\"")
					check := strings.TrimSuffix(importPath, plugin)
					// validate if check changed(Success), else fail
					assert.NotEqual(t, importPath, check, fmt.Sprintf("build tag is invalid for %v", d.Name()))
				})

			}
		}
		return nil
	})
}

func getPlugin(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "inputs.") {
			return strings.TrimPrefix(tag, "inputs.")
		}
	}
	return ""
}
