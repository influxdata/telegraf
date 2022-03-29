package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePluginData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	readme := `# plugin

## Configuration

` + "```" + `toml
# test plugin
[[input.plugin]]
  # No configuration
` + "```"
	r, err := os.Create("README.md")
	require.NoError(t, err)
	_, err = r.Write([]byte(readme))
	require.NoError(t, err)

	plugin := `package main
func (*Plugin) SampleConfig() string {
	return ` + "`{{ .SampleConfig }}`" + `
}
`
	sourceFile, err := os.Create("test.go")
	require.NoError(t, err)
	_, err = sourceFile.Write([]byte(plugin))
	require.NoError(t, err)

	defer func() {
		err = os.Remove("test.go")
		require.NoError(t, err)
		err = os.Remove("test.go.tmp")
		require.NoError(t, err)
		err = os.Remove("README.md")
		require.NoError(t, err)
	}()

	s, err := extractPluginData()
	require.NoError(t, err)

	err = generatePluginData("test", s)
	require.NoError(t, err)

	expected := `package main
func (*Plugin) SampleConfig() string {
	return ` + "`" + `# test plugin
[[input.plugin]]
  # No configuration
` + "`" + `
}
`

	newSourceFile, err := os.ReadFile("test.go")
	require.NoError(t, err)

	require.Equal(t, expected, string(newSourceFile))
}

func TestCleanGeneratedFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// Create files that will be cleaned up
	_, err := os.Create("testClean.go")
	require.NoError(t, err)
	_, err = os.Create("testClean.go.tmp")
	require.NoError(t, err)

	err = cleanGeneratedFiles("testClean")
	require.NoError(t, err)

	err = os.Remove("testClean.go")
	require.NoError(t, err)
}
