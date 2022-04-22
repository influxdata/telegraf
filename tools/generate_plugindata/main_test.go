package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var originalPlugin = `package main
func (*Plugin) SampleConfig() string {
	return ` + "`{{ .SampleConfig }}`" + `
}

`

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
	err = r.Close()
	require.NoError(t, err)

	sourceFile, err := os.Create("test_sample_config.go")
	require.NoError(t, err)
	_, err = sourceFile.Write([]byte(originalPlugin))
	require.NoError(t, err)
	err = sourceFile.Close()
	require.NoError(t, err)

	defer func() {
		err = os.Remove("test_sample_config.go")
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

	newSourceFile, err := os.ReadFile("test_sample_config.go")
	require.NoError(t, err)

	require.Equal(t, expected, string(newSourceFile))
}

func TestGeneratePluginDataNoConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	readme := `# plugin`

	r, err := os.Create("README.md")
	require.NoError(t, err)
	_, err = r.Write([]byte(readme))
	require.NoError(t, err)
	err = r.Close()
	require.NoError(t, err)

	defer func() {
		err = os.Remove("README.md")
		require.NoError(t, err)
	}()

	s, err := extractPluginData()
	require.NoError(t, err)
	require.Empty(t, s)
}

func setupGeneratedPluginFile(t *testing.T, fileName string) {
	// Create files that will be cleaned up
	r, err := os.Create(fileName)
	require.NoError(t, err)
	defer r.Close()

	updatePlugin := `package main
func (*Plugin) SampleConfig() string {
	return "I am a sample config"
}

`
	_, err = r.Write([]byte(updatePlugin))
	require.NoError(t, err)
}

func TestCleanGeneratedFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	filename := "testClean_sample_config.go"

	setupGeneratedPluginFile(t, filename)

	err := cleanGeneratedFiles("testClean")
	require.NoError(t, err)

	b, err := os.ReadFile(filename)
	require.NoError(t, err)

	require.Equal(t, originalPlugin, string(b))

	err = os.Remove(filename)
	require.NoError(t, err)
}
