package file

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"
)

const (
	expNewFile   = "test1,tag1=value1 value=1 1257894000000000000\n"
	expExistFile = "cpu,cpu=cpu0 value=100 1455312810012459582\n" +
		"test1,tag1=value1 value=1 1257894000000000000\n"
)

func TestFileExistingFile(t *testing.T) {
	fh := createFile(t)
	s, _ := serializers.NewInfluxSerializer()
	f := File{
		Files:      []string{fh.Name()},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	validateFile(t, fh.Name(), expExistFile)

	err = f.Close()
	require.NoError(t, err)
}

func TestFileNewFile(t *testing.T) {
	s, _ := serializers.NewInfluxSerializer()
	fh := tmpFile(t)
	f := File{
		Files:      []string{fh},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	validateFile(t, fh, expNewFile)

	err = f.Close()
	require.NoError(t, err)
}

func TestFileExistingFiles(t *testing.T) {
	fh1 := createFile(t)
	fh2 := createFile(t)
	fh3 := createFile(t)

	s, _ := serializers.NewInfluxSerializer()
	f := File{
		Files:      []string{fh1.Name(), fh2.Name(), fh3.Name()},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	validateFile(t, fh1.Name(), expExistFile)
	validateFile(t, fh2.Name(), expExistFile)
	validateFile(t, fh3.Name(), expExistFile)

	err = f.Close()
	require.NoError(t, err)
}

func TestFileNewFiles(t *testing.T) {
	s, _ := serializers.NewInfluxSerializer()
	fh1 := tmpFile(t)
	fh2 := tmpFile(t)
	fh3 := tmpFile(t)
	f := File{
		Files:      []string{fh1, fh2, fh3},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	validateFile(t, fh1, expNewFile)
	validateFile(t, fh2, expNewFile)
	validateFile(t, fh3, expNewFile)

	err = f.Close()
	require.NoError(t, err)
}

func TestFileBoth(t *testing.T) {
	fh1 := createFile(t)
	fh2 := tmpFile(t)

	s, _ := serializers.NewInfluxSerializer()
	f := File{
		Files:      []string{fh1.Name(), fh2},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	validateFile(t, fh1.Name(), expExistFile)
	validateFile(t, fh2, expNewFile)

	err = f.Close()
	require.NoError(t, err)
}

func TestFileStdout(t *testing.T) {
	// keep backup of the real stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	s, _ := serializers.NewInfluxSerializer()
	f := File{
		Files:      []string{"stdout"},
		serializer: s,
	}

	err := f.Connect()
	require.NoError(t, err)

	err = f.Write(testutil.MockMetrics())
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		outC <- buf.String()
	}()

	// back to normal state
	err = w.Close()
	require.NoError(t, err)

	// restoring the real stdout
	os.Stdout = old
	out := <-outC

	require.Equal(t, expNewFile, out)
}

func createFile(t *testing.T) *os.File {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	_, err = f.WriteString("cpu,cpu=cpu0 value=100 1455312810012459582\n")
	require.NoError(t, err)
	return f
}

func tmpFile(t *testing.T) string {
	return t.TempDir() + internal.RandomString(10)
}

func validateFile(t *testing.T, fileName, expS string) {
	buf, err := os.ReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, expS, string(buf))
}
