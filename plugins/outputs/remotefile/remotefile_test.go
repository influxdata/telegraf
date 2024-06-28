package remotefile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestStaticFileCreation(t *testing.T) {
	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"source": "localhost"},
			map[string]interface{}{"value": 42},
			time.Unix(1719410485, 0),
		),
	}
	expected := "test,source=localhost value=42i 1719410485000000000\n"

	tmpdir, err := os.MkdirTemp("", "telegraf-remotefile-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Setup the plugin including the serializer
	plugin := &File{
		Remote:            config.NewSecret([]byte("local:" + tmpdir)),
		Files:             []string{"test"},
		WriteBackInterval: config.Duration(100 * time.Millisecond),
		Log:               &testutil.Logger{},
	}

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Write the input metrics and close the plugin. This is required to
	// actually flush the data to disk
	require.NoError(t, plugin.Write(input))
	plugin.Close()

	// Check the result
	require.FileExists(t, filepath.Join(tmpdir, "test"))

	actual, err := os.ReadFile(filepath.Join(tmpdir, "test"))
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

func TestStaticFileAppend(t *testing.T) {
	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"source": "localhost"},
			map[string]interface{}{"value": 42},
			time.Unix(1719410485, 0),
		),
	}
	expected := "test,source=remotehost value=23i 1719410465000000000\n"
	expected += "test,source=localhost value=42i 1719410485000000000\n"

	tmpdir, err := os.MkdirTemp("", "telegraf-remotefile-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Create a file where we want to append to
	f, err := os.OpenFile(filepath.Join(tmpdir, "test"), os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)
	defer f.Close()
	_, err = f.WriteString("test,source=remotehost value=23i 1719410465000000000\n")
	require.NoError(t, err)
	f.Close()

	// Setup the plugin including the serializer
	plugin := &File{
		Remote:            config.NewSecret([]byte("local:" + tmpdir)),
		Files:             []string{"test"},
		WriteBackInterval: config.Duration(100 * time.Millisecond),
		Log:               &testutil.Logger{},
	}

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Write the input metrics and close the plugin. This is required to
	// actually flush the data to disk
	require.NoError(t, plugin.Write(input))
	plugin.Close()

	// Check the result
	require.FileExists(t, filepath.Join(tmpdir, "test"))

	actual, err := os.ReadFile(filepath.Join(tmpdir, "test"))
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

func TestDynamicFiles(t *testing.T) {
	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"source": "localhost"},
			map[string]interface{}{"value": 23},
			time.Unix(1719410465, 0),
		),
		metric.New(
			"test",
			map[string]string{"source": "remotehost"},
			map[string]interface{}{"value": 21},
			time.Unix(1719410465, 0),
		),
		metric.New(
			"test",
			map[string]string{"source": "localhost"},
			map[string]interface{}{"value": 42},
			time.Unix(1719410485, 0),
		),
		metric.New(
			"test",
			map[string]string{"source": "remotehost"},
			map[string]interface{}{"value": 66},
			time.Unix(1719410485, 0),
		),
		metric.New(
			"test",
			map[string]string{"source": "remotehost"},
			map[string]interface{}{"value": 55},
			time.Unix(1716310124, 0),
		),
		metric.New(
			"test",
			map[string]string{"source": "remotehost"},
			map[string]interface{}{"value": 1},
			time.Unix(1716310174, 0),
		),
	}
	expected := map[string][]string{
		"localhost-2024-06-26": {
			"test,source=localhost value=23i 1719410465000000000\n",
			"test,source=localhost value=42i 1719410485000000000\n",
		},
		"remotehost-2024-06-26": {
			"test,source=remotehost value=21i 1719410465000000000\n",
			"test,source=remotehost value=66i 1719410485000000000\n",
		},
		"remotehost-2024-05-21": {
			"test,source=remotehost value=55i 1716310124000000000\n",
			"test,source=remotehost value=1i 1716310174000000000\n",
		},
	}

	tmpdir, err := os.MkdirTemp("", "telegraf-remotefile-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Setup the plugin including the serializer
	plugin := &File{
		Remote:            config.NewSecret([]byte("local:" + tmpdir)),
		Files:             []string{`{{.Tag "source"}}-{{.Time.Format "2006-01-02"}}`},
		WriteBackInterval: config.Duration(100 * time.Millisecond),
		Log:               &testutil.Logger{},
	}

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Write the first batch of metrics wait for the data to settle to disk
	require.NoError(t, plugin.Write(input[:2]))
	require.Eventually(t, func() bool {
		_, err1 := os.Stat(filepath.Join(tmpdir, "localhost-2024-06-26"))
		_, err2 := os.Stat(filepath.Join(tmpdir, "remotehost-2024-06-26"))
		return err1 == nil && err2 == nil
	}, 5*time.Second, 100*time.Millisecond)

	// Check the result
	for _, fn := range []string{"localhost-2024-06-26", "remotehost-2024-06-26"} {
		tmpfn := filepath.Join(tmpdir, fn)
		require.FileExists(t, tmpfn)

		actual, err := os.ReadFile(tmpfn)
		require.NoError(t, err)
		require.Equal(t, expected[fn][0], string(actual))
	}

	require.NoError(t, plugin.Write(input[2:]))
	plugin.Close()

	// Check the result
	for fn, lines := range expected {
		expectedContent := strings.Join(lines, "")
		tmpfn := filepath.Join(tmpdir, fn)
		require.FileExists(t, tmpfn)

		actual, err := os.ReadFile(tmpfn)
		require.NoError(t, err)
		require.Equal(t, expectedContent, string(actual))
	}
}

func TestCustomTemplateFunctions(t *testing.T) {
	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"source": "localhost"},
			map[string]interface{}{"value": 42},
			time.Unix(1587686400, 0),
		),
	}
	expected := "test,source=localhost value=42i 1587686400000000000\n"

	expectedFilename := fmt.Sprintf("test-%d", time.Now().Year())

	tmpdir, err := os.MkdirTemp("", "telegraf-remotefile-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Setup the plugin including the serializer
	plugin := &File{
		Remote:            config.NewSecret([]byte("local:" + tmpdir)),
		Files:             []string{"test-{{now.Year}}"},
		WriteBackInterval: config.Duration(100 * time.Millisecond),
		Log:               &testutil.Logger{},
	}

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Write the input metrics and close the plugin. This is required to
	// actually flush the data to disk
	require.NoError(t, plugin.Write(input))
	plugin.Close()

	// Check the result
	require.FileExists(t, filepath.Join(tmpdir, expectedFilename))

	actual, err := os.ReadFile(filepath.Join(tmpdir, expectedFilename))
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}
