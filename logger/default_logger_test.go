package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

func TestLogTargetDefault(t *testing.T) {
	instance = defaultHandler()
	cfg := &Config{
		Quiet: true,
	}
	require.NoError(t, SetupLogging(cfg))
	logger, ok := instance.impl.(*defaultLogger)
	require.True(t, ok, "logging instance is not a default-logger")
	require.Equal(t, logger.logger.Writer(), os.Stderr)
}

func TestLogTargetStderr(t *testing.T) {
	instance = defaultHandler()
	cfg := &Config{
		LogTarget: "stderr",
		Quiet:     true,
	}
	require.NoError(t, SetupLogging(cfg))
	logger, ok := instance.impl.(*defaultLogger)
	require.True(t, ok, "logging instance is not a default-logger")
	require.Equal(t, logger.logger.Writer(), os.Stderr)
}

func TestLogTargetFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogTarget:           "file",
		RotationMaxArchives: -1,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, buf[19:], []byte("Z I! TEST\n"))
}

func TestLogTargetFileDebug(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogTarget:           "file",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("D! TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, buf[19:], []byte("Z D! TEST\n"))
}

func TestLogTargetFileError(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogTarget:           "file",
		RotationMaxArchives: -1,
		Quiet:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, buf[19:], []byte("Z E! TEST\n"))
}

func TestAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogTarget:           "file",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, buf[19:], []byte("Z I! TEST\n"))
}

func TestWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogTarget:           "file",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, buf[19:], []byte("Z I! TEST\n"))

	tmpf, err := os.OpenFile(tmpfile.Name(), os.O_RDWR|os.O_TRUNC, 0640)
	require.NoError(t, err)
	require.NoError(t, tmpf.Close())

	log.Printf("SHOULD BE FIRST")

	buf, err = os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, buf[19:], []byte("Z I! SHOULD BE FIRST\n"))
}

func TestWriteToFileInRotation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		Logfile:             filepath.Join(tempDir, "test.log"),
		LogTarget:           "file",
		RotationMaxArchives: -1,
		RotationMaxSize:     30,
	}
	require.NoError(t, SetupLogging(cfg))

	// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
	defer CloseLogging() //nolint:errcheck // We cannot do anything if this fails

	log.Printf("I! TEST 1") // Writes 31 bytes, will rotate
	log.Printf("I! TEST")   // Writes 29 byes, no rotation expected

	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func BenchmarkTelegrafLogWrite(b *testing.B) {
	l, err := createDefaultLogger(&Config{})
	require.NoError(b, err)

	// Discard all logging output
	dl := l.(*defaultLogger)
	dl.SetOutput(io.Discard)

	ts := time.Now()
	for i := 0; i < b.N; i++ {
		dl.Print(telegraf.Debug, ts, "", "test")
	}
}
