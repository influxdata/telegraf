package rotate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWriter_NoRotation(t *testing.T) {
	tempDir := t.TempDir()
	writer, err := NewFileWriter(filepath.Join(tempDir, "test"), 0, 0, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	_, err = writer.Write([]byte("Hello World"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("Hello World 2"))
	require.NoError(t, err)
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 1, len(files))
}

func TestFileWriter_TimeRotation(t *testing.T) {
	tempDir := t.TempDir()
	interval, _ := time.ParseDuration("10ms")
	writer, err := NewFileWriter(filepath.Join(tempDir, "test"), interval, 0, -1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	_, err = writer.Write([]byte("Hello World"))
	require.NoError(t, err)
	time.Sleep(interval)
	_, err = writer.Write([]byte("Hello World 2"))
	require.NoError(t, err)
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestFileWriter_ReopenTimeRotation(t *testing.T) {
	tempDir := t.TempDir()
	interval, _ := time.ParseDuration("10ms")
	filePath := filepath.Join(tempDir, "test.log")
	err := os.WriteFile(filePath, []byte("Hello World"), 0644)
	time.Sleep(interval)
	assert.NoError(t, err)
	writer, err := NewFileWriter(filepath.Join(tempDir, "test.log"), interval, 0, -1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestFileWriter_SizeRotation(t *testing.T) {
	tempDir := t.TempDir()
	maxSize := int64(9)
	writer, err := NewFileWriter(filepath.Join(tempDir, "test.log"), 0, maxSize, -1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	_, err = writer.Write([]byte("Hello World"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("World 2"))
	require.NoError(t, err)
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestFileWriter_ReopenSizeRotation(t *testing.T) {
	tempDir := t.TempDir()
	maxSize := int64(12)
	filePath := filepath.Join(tempDir, "test.log")
	err := os.WriteFile(filePath, []byte("Hello World"), 0644)
	assert.NoError(t, err)
	writer, err := NewFileWriter(filepath.Join(tempDir, "test.log"), 0, maxSize, -1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	_, err = writer.Write([]byte("Hello World Again"))
	require.NoError(t, err)
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestFileWriter_DeleteArchives(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long test in short mode")
	}

	tempDir := t.TempDir()
	maxSize := int64(5)
	writer, err := NewFileWriter(filepath.Join(tempDir, "test.log"), 0, maxSize, 2)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, writer.Close()) })

	_, err = writer.Write([]byte("First file"))
	require.NoError(t, err)
	// File names include the date with second precision
	// So, to force rotation with different file names
	// we need to wait
	time.Sleep(1 * time.Second)
	_, err = writer.Write([]byte("Second file"))
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	_, err = writer.Write([]byte("Third file"))
	require.NoError(t, err)

	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 3, len(files))

	for _, tempFile := range files {
		var bytes []byte
		var err error
		path := filepath.Join(tempDir, tempFile.Name())
		if bytes, err = os.ReadFile(path); err != nil {
			t.Error(err.Error())
			return
		}
		contents := string(bytes)

		if contents != "" && contents != "Second file" && contents != "Third file" {
			t.Error("Should have deleted the eldest log file")
			return
		}
	}
}

func TestFileWriter_CloseRotates(t *testing.T) {
	tempDir := t.TempDir()
	maxSize := int64(9)
	writer, err := NewFileWriter(filepath.Join(tempDir, "test.log"), 0, maxSize, -1)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 1, len(files))
	assert.Regexp(t, "^test\\.[^\\.]+\\.log$", files[0].Name())
}
