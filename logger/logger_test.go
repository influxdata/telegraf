package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	cfg := createBasicLogConfig(tmpfile.Name())
	SetupLogging(cfg)
	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	f, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestDebugWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	cfg := createBasicLogConfig(tmpfile.Name())
	cfg.Debug = true
	SetupLogging(cfg)
	log.Printf("D! TEST")

	f, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z D! TEST\n"))
}

func TestErrorWriteLogToFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	cfg := createBasicLogConfig(tmpfile.Name())
	cfg.Quiet = true
	SetupLogging(cfg)
	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	f, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z E! TEST\n"))
}

func TestAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	cfg := createBasicLogConfig(tmpfile.Name())
	cfg.Debug = true
	SetupLogging(cfg)
	log.Printf("TEST")

	f, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	cfg := createBasicLogConfig(tmpfile.Name())
	cfg.Debug = true
	SetupLogging(cfg)
	log.Printf("TEST")

	f, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))

	tmpf, err := os.OpenFile(tmpfile.Name(), os.O_RDWR|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NoError(t, tmpf.Close())

	log.Printf("SHOULD BE FIRST")

	f, err = os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! SHOULD BE FIRST\n"))
}

func TestWriteToFileInRotation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := createBasicLogConfig(filepath.Join(tempDir, "test.log"))
	cfg.LogTarget = LogTargetFile
	cfg.RotationMaxSize = config.Size(30)
	writer := newLogWriter(cfg)
	// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
	closer, isCloser := writer.(io.Closer)
	assert.True(t, isCloser)
	t.Cleanup(func() { require.NoError(t, closer.Close()) })

	log.Printf("I! TEST 1") // Writes 31 bytes, will rotate
	log.Printf("I! TEST")   // Writes 29 byes, no rotation expected
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestLogTargetSettings(t *testing.T) {
	cfg := LogConfig{
		LogTarget: "",
		Quiet:     true,
	}
	SetupLogging(cfg)
	logger, isTelegrafLogger := actualLogger.(*telegrafLog)
	assert.True(t, isTelegrafLogger)
	assert.Equal(t, logger.internalWriter, os.Stderr)

	cfg = LogConfig{
		LogTarget: "stderr",
		Quiet:     true,
	}
	SetupLogging(cfg)
	logger, isTelegrafLogger = actualLogger.(*telegrafLog)
	assert.True(t, isTelegrafLogger)
	assert.Equal(t, logger.internalWriter, os.Stderr)
}

func BenchmarkTelegrafLogWrite(b *testing.B) {
	var msg = []byte("test")
	var buf bytes.Buffer
	w, err := newTelegrafWriter(&buf, LogConfig{})
	if err != nil {
		panic("Unable to create log writer.")
	}
	for i := 0; i < b.N; i++ {
		buf.Reset()
		w.Write(msg)
	}
}

func createBasicLogConfig(filename string) LogConfig {
	return LogConfig{
		Logfile:             filename,
		LogTarget:           LogTargetFile,
		RotationMaxArchives: -1,
	}
}
