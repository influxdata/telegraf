package directory_monitor

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCSVGZImport(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"
	testCsvGzFile := "test.csv.gz"

	// Establish process directory and finished directory.
	finishedDirectory, err := ioutil.TempDir("", "finished")
	require.NoError(t, err)
	processDirectory, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(processDirectory)
	defer os.RemoveAll(finishedDirectory)

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      100000,
	}
	err = r.Init()
	require.NoError(t, err)

	parserConfig := parsers.Config{
		DataFormat:        "csv",
		CSVHeaderRowCount: 1,
	}
	require.NoError(t, err)
	r.SetParserFunc(func() (parsers.Parser, error) {
		return parsers.NewParser(&parserConfig)
	})
	r.Log = testutil.Logger{}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	f.WriteString("thing,color\nsky,blue\ngrass,green\nclifford,red\n")
	f.Close()

	// Write csv.gz file to process into the 'process' directory.
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte("thing,color\nsky,blue\ngrass,green\nclifford,red\n"))
	w.Close()
	err = ioutil.WriteFile(filepath.Join(processDirectory, testCsvGzFile), b.Bytes(), 0666)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(6)
	r.Stop()

	// Verify that we read both files once.
	require.Equal(t, len(acc.Metrics), 6)

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvFile))
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvGzFile))

	require.NoError(t, err)
}

// For JSON data.
type event struct {
	Name   string
	Speed  float64
	Length float64
}

func TestMultipleJSONFileImports(t *testing.T) {
	acc := testutil.Accumulator{}
	testJsonFile := "test.json"

	// Establish process directory and finished directory.
	finishedDirectory, err := ioutil.TempDir("", "finished")
	require.NoError(t, err)
	processDirectory, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(processDirectory)
	defer os.RemoveAll(finishedDirectory)

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      1000,
	}
	err = r.Init()
	require.NoError(t, err)

	parserConfig := parsers.Config{
		DataFormat:  "json",
		JSONNameKey: "Name",
	}

	r.SetParserFunc(func() (parsers.Parser, error) {
		return parsers.NewParser(&parserConfig)
	})

	// Let's drop a 5-line LINE-DELIMITED json.
	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testJsonFile))
	require.NoError(t, err)
	f.WriteString("{\"Name\": \"event1\",\"Speed\": 100.1,\"Length\": 20.1}\n{\"Name\": \"event2\",\"Speed\": 500,\"Length\": 1.4}\n{\"Name\": \"event3\",\"Speed\": 200,\"Length\": 10.23}\n{\"Name\": \"event4\",\"Speed\": 80,\"Length\": 250}\n{\"Name\": \"event5\",\"Speed\": 120.77,\"Length\": 25.97}")
	f.Close()

	err = r.Start(&acc)
	r.Log = testutil.Logger{}
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(5)
	r.Stop()

	// Verify that we read each JSON line once to a single metric.
	require.Equal(t, len(acc.Metrics), 5)
}
