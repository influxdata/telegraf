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

func TestTextStderr(t *testing.T) {
	instance = defaultHandler()
	cfg := &Config{
		LogFormat: "text",
		Quiet:     true,
	}
	require.NoError(t, SetupLogging(cfg))
	logger, ok := instance.impl.(*textLogger)
	require.Truef(t, ok, "logging instance is not a text-logger but %T", instance.impl)
	require.Equal(t, logger.logger.Writer(), os.Stderr)
}

func TestTextFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z I! TEST\n", string(buf[19:]))
}

func TestTextFileDebug(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("D! TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z D! TEST\n", string(buf[19:]))
}

func TestTextFileError(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Quiet:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z E! TEST\n", string(buf[19:]))
}

func TestTextAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z I! TEST\n", string(buf[19:]))
}

func TestTextWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	log.Printf("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z I! TEST\n", string(buf[19:]))

	require.NoError(t, os.Truncate(tmpfile.Name(), 0))

	log.Printf("SHOULD BE FIRST")

	buf, err = os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Equal(t, "Z I! SHOULD BE FIRST\n", string(buf[19:]))
}

func TestTextWriteToFileInRotation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		Logfile:             filepath.Join(tempDir, "test.log"),
		LogFormat:           "text",
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

func TestTextWriteDerivedLogger(t *testing.T) {
	instance = defaultHandler()

	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	l := New("testing", "test", "")
	l.Info("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z I! [testing.test] TEST\n", string(buf[19:]))
}

func TestTextWriteDerivedLoggerWithAttributes(t *testing.T) {
	instance = defaultHandler()

	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		Logfile:             tmpfile.Name(),
		LogFormat:           "text",
		RotationMaxArchives: -1,
		Debug:               true,
	}
	require.NoError(t, SetupLogging(cfg))

	l := New("testing", "test", "myalias")

	// All attributes should be ignored
	l.AddAttribute("alias", "foo")
	l.AddAttribute("device_id", 123)

	l.Info("TEST")

	buf, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	require.Greater(t, len(buf), 19)
	require.Equal(t, "Z I! [testing.test::myalias] TEST\n", string(buf[19:]))
}

func BenchmarkTelegrafTextLogWrite(b *testing.B) {
	l, err := createTextLogger(&Config{})
	require.NoError(b, err)

	// Discard all logging output
	dl := l.(*textLogger)
	dl.logger.SetOutput(io.Discard)

	ts := time.Now()
	for i := 0; i < b.N; i++ {
		dl.Print(telegraf.Debug, ts, "", nil, "test")
	}
}
