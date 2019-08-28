package logger

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	config := createBasicLogConfig(tmpfile.Name())
	SetupLogging(config)
	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestDebugWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	config := createBasicLogConfig(tmpfile.Name())
	config.Debug = true
	SetupLogging(config)
	log.Printf("D! TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z D! TEST\n"))
}

func TestErrorWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	config := createBasicLogConfig(tmpfile.Name())
	config.Quiet = true
	SetupLogging(config)
	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z E! TEST\n"))
}

func TestAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	config := createBasicLogConfig(tmpfile.Name())
	config.Debug = true
	SetupLogging(config)
	log.Printf("TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()
	config := createBasicLogConfig(tmpfile.Name())
	config.Debug = true
	SetupLogging(config)
	log.Printf("TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))

	tmpf, err := os.OpenFile(tmpfile.Name(), os.O_RDWR|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NoError(t, tmpf.Close())

	log.Printf("SHOULD BE FIRST")

	f, err = ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! SHOULD BE FIRST\n"))
}

func TestWriteToFileInRotation(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "LogRotation")
	require.NoError(t, err)
	config := createBasicLogConfig(filepath.Join(tempDir, "test.log"))
	config.LogTarget = LogTargetFile
	config.RotationMaxSize = internal.Size{Size: int64(30)}
	writer := newLogWriter(config)
	// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
	closer, isCloser := writer.(io.Closer)
	assert.True(t, isCloser)
	defer func() { closer.Close(); os.RemoveAll(tempDir) }()

	log.Printf("I! TEST 1") // Writes 31 bytes, will rotate
	log.Printf("I! TEST")   // Writes 29 byes, no rotation expected
	files, _ := ioutil.ReadDir(tempDir)
	assert.Equal(t, 2, len(files))
}

func TestLogTargetSettings(t *testing.T) {
	config := LogConfig{
		LogTarget: "",
		Quiet:     true,
	}
	SetupLogging(config)
	logger, isTelegrafLogger := actualLogger.(*telegrafLog)
	assert.True(t, isTelegrafLogger)
	assert.Equal(t, logger.internalWriter, os.Stderr)

	config = LogConfig{
		LogTarget: "stderr",
		Quiet:     true,
	}
	SetupLogging(config)
	logger, isTelegrafLogger = actualLogger.(*telegrafLog)
	assert.True(t, isTelegrafLogger)
	assert.Equal(t, logger.internalWriter, os.Stderr)
}

func BenchmarkTelegrafLogWrite(b *testing.B) {
	var msg = []byte("test")
	var buf bytes.Buffer
	w := newTelegrafWriter(&buf)
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
