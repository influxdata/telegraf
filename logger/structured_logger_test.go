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

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

func TestStructuredLogger(t *testing.T) {
	tempDir := t.TempDir()
	anotherTempDir := t.TempDir()

	t.Run("TestStructuredStderr", func(t *testing.T) {
		instance = defaultHandler()
		cfg := &Config{
			LogFormat: "structured",
			Quiet:     true,
		}
		require.NoError(t, SetupLogging(cfg))
		logger, ok := instance.impl.(*structuredLogger)
		require.Truef(t, ok, "logging instance is not a structured-logger but %T", instance.impl)
		require.Equal(t, logger.output, os.Stderr)
	})

	t.Run("TestStructuredFile", func(t *testing.T) {
		tmpFile := filepath.Join(tempDir, "TestStructuredFile.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
		}
		require.NoError(t, SetupLogging(cfg))

		log.Printf("I! TEST")
		log.Printf("D! TEST") // <- should be ignored

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredFileDebug", func(t *testing.T) {
		tmpFile := filepath.Join(tempDir, "TestStructuredFileDebug.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Debug:               true,
		}
		require.NoError(t, SetupLogging(cfg))

		log.Printf("D! TEST")

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredFileError", func(t *testing.T) {
		tmpFile := filepath.Join(tempDir, "TestStructuredFileError.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Quiet:               true,
		}
		require.NoError(t, SetupLogging(cfg))

		log.Printf("E! TEST")
		log.Printf("I! TEST") // <- should be ignored

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredAddDefaultLogLevel", func(t *testing.T) {
		tmpFile := filepath.Join(tempDir, "TestStructuredAddDefaultLogLevel.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Debug:               true,
		}
		require.NoError(t, SetupLogging(cfg))
		// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
		defer CloseLogging() //nolint:errcheck // We cannot do anything if this fails

		log.Printf("TEST")

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredDerivedLogger", func(t *testing.T) {
		instance = defaultHandler()
		tmpFile := filepath.Join(tempDir, "TestStructuredDerivedLogger.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Debug:               true,
		}
		require.NoError(t, SetupLogging(cfg))
		// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
		defer CloseLogging() //nolint:errcheck // We cannot do anything if this fails

		l := New("testing", "test", "")
		l.Info("TEST")

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredDerivedLoggerWithAttributes", func(t *testing.T) {
		instance = defaultHandler()
		tmpFile := filepath.Join(tempDir, "TestStructuredDerivedLoggerWithAttributes.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Debug:               true,
		}
		require.NoError(t, SetupLogging(cfg))

		l := New("testing", "test", "myalias")
		l.AddAttribute("alias", "foo") // Should be ignored
		l.AddAttribute("device_id", 123)

		l.Info("TEST")

		buf, err := os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredWriteToTruncatedFile", func(t *testing.T) {
		tmpFile := filepath.Join(tempDir, "TestStructuredWriteToTruncatedFile.log")

		cfg := &Config{
			Logfile:             tmpFile,
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			Debug:               true,
		}
		require.NoError(t, SetupLogging(cfg))

		log.Printf("TEST")

		buf, err := os.ReadFile(tmpFile)
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

		require.NoError(t, os.Truncate(tmpFile, 0))

		log.Printf("SHOULD BE FIRST")

		buf, err = os.ReadFile(tmpFile)
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
	})

	t.Run("TestStructuredWriteToFileInRotation", func(t *testing.T) {
		cfg := &Config{
			Logfile:             filepath.Join(anotherTempDir, "TestStructuredWriteToFileInRotation.log"),
			LogFormat:           "structured",
			RotationMaxArchives: -1,
			RotationMaxSize:     30,
		}
		require.NoError(t, SetupLogging(cfg))
		// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
		defer CloseLogging() //nolint:errcheck // We cannot do anything if this fails

		log.Printf("I! TEST 1") // Writes 31 bytes, will rotate
		log.Printf("I! TEST")   // Writes 29 byes, no rotation expected

		files, err := os.ReadDir(anotherTempDir)
		require.NoError(t, err)
		require.Len(t, files, 2)
	})

	t.Run("TestStructuredLogMessageKey", func(t *testing.T) {
		instance = defaultHandler()
		tmpFile := filepath.Join(tempDir, "TestStructuredLogMessageKey.log")

		cfg := &Config{
			Logfile:                 tmpFile,
			LogFormat:               "structured",
			RotationMaxArchives:     -1,
			Debug:                   true,
			StructuredLogMessageKey: "message",
		}
		require.NoError(t, SetupLogging(cfg))
		// Close the writer here, otherwise the temp folder cannot be deleted because the current log file is in use.
		defer CloseLogging() //nolint:errcheck // We cannot do anything if this fails

		l := New("testing", "test", "")
		l.Info("TEST")

		buf, err := os.ReadFile(tmpFile)
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
	})
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
