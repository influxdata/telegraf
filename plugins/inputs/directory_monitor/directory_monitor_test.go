package directory_monitor

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/testutil"
)

func TestCSVGZImport(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"
	testCsvGzFile := "test.csv.gz"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      100000,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 1,
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString("thing,color\nsky,blue\ngrass,green\nclifford,red\n")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Write csv.gz file to process into the 'process' directory.
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err = w.Write([]byte("thing,color\nsky,blue\ngrass,green\nclifford,red\n"))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(processDirectory, testCsvGzFile), b.Bytes(), 0666)
	require.NoError(t, err)

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
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvGzFile))
	require.NoError(t, err)
}

func TestMultipleJSONFileImports(t *testing.T) {
	acc := testutil.Accumulator{}
	testJSONFile := "test.json"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      1000,
	}
	err := r.Init()
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
	f, err := os.Create(filepath.Join(processDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString("{\"Name\": \"event1\",\"Speed\": 100.1,\"Length\": 20.1}\n{\"Name\": \"event2\",\"Speed\": 500,\"Length\": 1.4}\n{\"Name\": \"event3\",\"Speed\": 200,\"Length\": 10.23}\n{\"Name\": \"event4\",\"Speed\": 80,\"Length\": 250}\n{\"Name\": \"event5\",\"Speed\": 120.77,\"Length\": 25.97}")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

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

func TestFileTag(t *testing.T) {
	acc := testutil.Accumulator{}
	testJSONFile := "test.json"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		FileTag:            "filename",
		MaxBufferedMetrics: 1000,
		FileQueueSize:      1000,
	}
	err := r.Init()
	require.NoError(t, err)

	parserConfig := parsers.Config{
		DataFormat:  "json",
		JSONNameKey: "Name",
	}

	r.SetParserFunc(func() (parsers.Parser, error) {
		return parsers.NewParser(&parserConfig)
	})

	// Let's drop a 1-line LINE-DELIMITED json.
	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString("{\"Name\": \"event1\",\"Speed\": 100.1,\"Length\": 20.1}")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	err = r.Start(&acc)
	r.Log = testutil.Logger{}
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	// Verify that we read each JSON line once to a single metric.
	require.Equal(t, len(acc.Metrics), 1)
	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			require.Equal(t, r.FileTag, key)
			require.Equal(t, filepath.Base(testJSONFile), value)
		}
	}
}

func TestCSVNoSkipRows(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      100000,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 1,
			SkipRows:       0,
			TagColumns:     []string{"line1"},
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	testCSV := `line1,line2,line3
hello,80,test_name2`

	expectedFields := map[string]interface{}{
		"line2": int64(80),
		"line3": "test_name2",
	}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString(testCSV)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	// Verify that we read both files once.
	require.Equal(t, len(acc.Metrics), 1)

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvFile))
	require.NoError(t, err)
	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			require.Equal(t, "line1", key)
			require.Equal(t, "hello", value)
		}
		require.Equal(t, expectedFields, m.Fields)
	}
}

func TestCSVSkipRows(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      100000,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 1,
			SkipRows:       2,
			TagColumns:     []string{"line1"},
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	testCSV := `garbage nonsense 1
garbage,nonsense,2
line1,line2,line3
hello,80,test_name2`

	expectedFields := map[string]interface{}{
		"line2": int64(80),
		"line3": "test_name2",
	}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString(testCSV)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	// Verify that we read both files once.
	require.Equal(t, len(acc.Metrics), 1)

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvFile))
	require.NoError(t, err)
	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			require.Equal(t, "line1", key)
			require.Equal(t, "hello", value)
		}
		require.Equal(t, expectedFields, m.Fields)
	}
}

func TestCSVMultiHeader(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: 1000,
		FileQueueSize:      100000,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 2,
			TagColumns:     []string{"line1"},
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	testCSV := `line,line,line
1,2,3
hello,80,test_name2`

	expectedFields := map[string]interface{}{
		"line2": int64(80),
		"line3": "test_name2",
	}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString(testCSV)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	// Verify that we read both files once.
	require.Equal(t, len(acc.Metrics), 1)

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvFile))
	require.NoError(t, err)
	for _, m := range acc.Metrics {
		for key, value := range m.Tags {
			require.Equal(t, "line1", key)
			require.Equal(t, "hello", value)
		}
		require.Equal(t, expectedFields, m.Fields)
	}
}
