package logger

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
)

func TestStructuredStderr(t *testing.T) {
	instance = defaultHandler()
	cfg := &Config{
		LogFormat: "structured",
		Quiet:     true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	logger, ok := instance.impl.(*structuredLogger)
	require.Truef(t, ok, "logging instance is not a structured-logger but %T", instance.impl)
	require.Equal(t, logger.output, os.Stderr)
}

func TestStructuredFile(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level": "INFO",
		"msg":   "TEST",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredFileDebug(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("D! TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level": "DEBUG",
		"msg":   "TEST",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredFileError(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Quiet:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)

	expected := map[string]interface{}{
		"level": "ERROR",
		"msg":   "TEST",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level": "INFO",
		"msg":   "TEST",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredDerivedLogger(t *testing.T) {
	instance = defaultHandler()

	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	l := New("testing", "test", "")
	l.Info("TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level":    "INFO",
		"msg":      "TEST",
		"category": "testing",
		"plugin":   "test",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredDerivedLoggerWithAttributes(t *testing.T) {
	instance = defaultHandler()

	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	l := New("testing", "test", "myalias")
	l.AddAttribute("alias", "foo") // Should be ignored
	l.AddAttribute("device_id", 123)

	l.Info("TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level":     "INFO",
		"msg":       "TEST",
		"category":  "testing",
		"plugin":    "test",
		"alias":     "myalias",
		"device_id": float64(123),
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:             filename,
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level": "INFO",
		"msg":   "TEST",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)

	require.NoError(t, os.Truncate(filename, 0))

	log.Printf("SHOULD BE FIRST")

	buf, err = os.ReadFile(filename)
	require.NoError(t, err)

	expected = map[string]interface{}{
		"level": "INFO",
		"msg":   "SHOULD BE FIRST",
	}

	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func TestStructuredWriteToFileInRotation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		Logfile:             filepath.Join(tempDir, "test.log"),
		LogFormat:           "structured",
		RotationMaxArchives: -1,
		RotationMaxSize:     70,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	log.Printf("I! TEST 1") // Writes 70 bytes in structured format, will rotate
	log.Printf("I! TEST")   // Writes 68 bytes in structured format, no rotation expected

	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestStructuredLogMessageKey(t *testing.T) {
	instance = defaultHandler()

	tmpfile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	filename := tmpfile.Name()
	require.NoError(t, tmpfile.Close())

	cfg := &Config{
		Logfile:                 filename,
		LogFormat:               "structured",
		RotationMaxArchives:     -1,
		Debug:                   true,
		StructuredLogMessageKey: "message",
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	l := New("testing", "test", "")
	l.Info("TEST")

	buf, err := os.ReadFile(filename)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"level":    "INFO",
		"message":  "TEST",
		"category": "testing",
		"plugin":   "test",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))

	require.Contains(t, actual, "time")
	require.NotEmpty(t, actual["time"])
	delete(actual, "time")
	require.Equal(t, expected, actual)
}

func BenchmarkTelegrafStructuredLogWrite(b *testing.B) {
	// Discard all logging output
	l := &structuredLogger{
		handler: slog.NewJSONHandler(io.Discard, defaultStructuredHandlerOptions),
		output:  io.Discard,
		errlog:  log.New(os.Stderr, "", 0),
	}

	ts := time.Now()
	for i := 0; i < b.N; i++ {
		l.Print(telegraf.Debug, ts, "", nil, "test")
	}
}
