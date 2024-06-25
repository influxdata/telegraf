package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/wlog"
)

func TestWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := createBasicConfig(tmpfile.Name())
	err = SetupLogging(cfg)
	require.NoError(t, err)
	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	f, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestDebugWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := createBasicConfig(tmpfile.Name())
	cfg.Debug = true
	err = SetupLogging(cfg)
	require.NoError(t, err)
	log.Printf("D! TEST")

	f, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z D! TEST\n"))
}

func TestErrorWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := createBasicConfig(tmpfile.Name())
	cfg.Quiet = true
	err = SetupLogging(cfg)
	require.NoError(t, err)
	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	f, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z E! TEST\n"))
}

func TestAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := createBasicConfig(tmpfile.Name())
	cfg.Debug = true
	err = SetupLogging(cfg)
	require.NoError(t, err)
	log.Printf("TEST")

	f, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := createBasicConfig(tmpfile.Name())
	cfg.Debug = true
	err = SetupLogging(cfg)
	require.NoError(t, err)
	log.Printf("TEST")

	f, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z I! TEST\n"))

	tmpf, err := os.OpenFile(tmpfile.Name(), os.O_RDWR|os.O_TRUNC, 0640)
	require.NoError(t, err)
	require.NoError(t, tmpf.Close())

	log.Printf("SHOULD BE FIRST")

	f, err = os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, f[19:], []byte("Z I! SHOULD BE FIRST\n"))
}

func TestWriteToFileInRotation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := createBasicConfig(filepath.Join(tempDir, "test.log"))
	cfg.RotationMaxSize = 30
	require.NoError(t, SetupLogging(cfg))
	// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
	t.Cleanup(func() { require.NoError(t, instance.Close()) })

	log.Printf("I! TEST 1") // Writes 31 bytes, will rotate
	log.Printf("I! TEST")   // Writes 29 byes, no rotation expected
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestLogTargetSettings(t *testing.T) {
	instance = nil
	cfg := &Config{
		LogTarget: "",
		Quiet:     true,
	}
	require.NoError(t, SetupLogging(cfg))
	logger, isTelegrafLogger := instance.(*defaultLogger)
	require.True(t, isTelegrafLogger)
	require.Equal(t, logger.internalWriter, os.Stderr)

	cfg = &Config{
		LogTarget: "stderr",
		Quiet:     true,
	}
	require.NoError(t, SetupLogging(cfg))
	logger, isTelegrafLogger = instance.(*defaultLogger)
	require.True(t, isTelegrafLogger)
	require.Equal(t, logger.internalWriter, os.Stderr)
}

func BenchmarkTelegrafLogWrite(b *testing.B) {
	l, err := createDefaultLogger(&Config{})
	require.NoError(b, err)

	// Discard all logging output
	dl := l.(*defaultLogger)
	dl.writer = wlog.NewWriter(io.Discard)
	dl.internalWriter = io.Discard

	for i := 0; i < b.N; i++ {
		dl.Info("test")
	}
}

func createBasicConfig(filename string) *Config {
	return &Config{
		Logfile:             filename,
		LogTarget:           "file",
		RotationMaxArchives: -1,
	}
}
