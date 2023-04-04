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

// exceptionMap holds those plugins which differ in conventions when defining plugins.
// Most plugins follow the convention of pkg-name to plugin-name.
// For ex, Pivot processor pkg github.com/influxdata/telegraf/plugins/processors/pivot maps directly to
// the last element of the pkg i.e "pivot"
// But in case of "aws_ec2" processor, the pkg is defined as "github.com/influxdata/telegraf/plugins/processors/aws/ec2".
// This ensure package names are not tied with plugin names.
// it should be of the form <pkg-name>: <plugin-name>
var exceptionMap = map[string]string{
	"github.com/influxdata/telegraf/plugins/processors/aws/ec2": "aws_ec2",
}

func Test_AllPlugins(t *testing.T) {
	pluginDirs := []string{"aggregators", "inputs", "outputs", "parsers", "processors", "secretstores"}
	for _, dir := range pluginDirs {
		testPluginDirectory(t, dir)
	}
}

func testPluginDirectory(t *testing.T, directory string) {
	allDir := filepath.Join(directory, "all")
	pluginCategory := directory
	err := filepath.WalkDir(allDir, func(path string, d fs.DirEntry, walkErr error) error {
		require.NoError(t, walkErr)
		if d.IsDir() || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		require.NoError(t, err)

		for _, cg := range node.Comments {
			for _, comm := range cg.List {
				if !strings.HasPrefix(comm.Text, "//go:build") {
					continue
				}
				testName := fmt.Sprintf("%v-%v", pluginCategory, strings.TrimSuffix(d.Name(), ".go"))
				t.Run(testName, func(t *testing.T) {
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
					assert.Contains(t, tags, pluginCategory)
					plugin := getPlugin(tags, pluginCategory)
					assert.Greater(t, len(plugin), 0)

					// should contain one or more import statements
					assert.GreaterOrEqual(t, len(node.Imports), 1)
					// trim the path surrounded by quotes
					importPath := strings.Trim(node.Imports[0].Path.Value, "\"")

					// check if present in exceptionMap
					exception, ok := exceptionMap[importPath]
					if ok {
						assert.Equal(t, plugin, exception)
						return
					}
					check := strings.TrimSuffix(importPath, plugin)
					// validate if check changed(Success), else fail
					assert.NotEqual(t, importPath, check, fmt.Sprintf("build tag is invalid for %v", d.Name()))
				})
			}
		}
		return nil
	})
	require.NoError(t, err)
}

func getPlugin(tags []string, pluginCategory string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, fmt.Sprintf("%v.", pluginCategory)) {
			return strings.TrimPrefix(tag, fmt.Sprintf("%v.", pluginCategory))
		}
	}
	return ""
}
