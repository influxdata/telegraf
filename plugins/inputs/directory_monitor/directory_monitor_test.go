package directory_monitor

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestCreator(t *testing.T) {
	creator, found := inputs.Inputs["directory_monitor"]
	require.True(t, found)

	expected := &DirectoryMonitor{
		FilesToMonitor:             defaultFilesToMonitor,
		FilesToIgnore:              defaultFilesToIgnore,
		MaxBufferedMetrics:         defaultMaxBufferedMetrics,
		DirectoryDurationThreshold: defaultDirectoryDurationThreshold,
		FileQueueSize:              defaultFileQueueSize,
		ParseMethod:                defaultParseMethod,
	}

	require.Equal(t, expected, creator())
}

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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
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

func TestCSVGZImportWithHeader(t *testing.T) {
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 1,
			SkipRows:       1,
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString("This is some garbage to be skipped\n")
	require.NoError(t, err)
	_, err = f.WriteString("thing,color\nsky,blue\ngrass,green\nclifford,red\n")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Write csv.gz file to process into the 'process' directory.
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err = w.Write([]byte("This is some garbage to be skipped\n"))
	require.NoError(t, err)
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		p := &json.Parser{NameKey: "Name"}
		err := p.Init()
		return p, err
	})

	// Let's drop a 5-line LINE-DELIMITED json.
	// Write csv file to process into the 'process' directory.
	f, err := os.Create(filepath.Join(processDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString(
		"{\"Name\": \"event1\",\"Speed\": 100.1,\"Length\": 20.1}\n{\"Name\": \"event2\",\"Speed\": 500,\"Length\": 1.4}\n" +
			"{\"Name\": " + "\"event3\",\"Speed\": 200,\"Length\": 10.23}\n{\"Name\": \"event4\",\"Speed\": 80,\"Length\": 250}\n" +
			"{\"Name\": \"event5\",\"Speed\": 120.77,\"Length\": 25.97}",
	)
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (parsers.Parser, error) {
		p := &json.Parser{NameKey: "Name"}
		err := p.Init()
		return p, err
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
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
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
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

func TestParseCompleteFile(t *testing.T) {
	acc := testutil.Accumulator{}

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        "at-once",
	}
	err := r.Init()
	require.NoError(t, err)
	r.Log = testutil.Logger{}

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := &json.Parser{
			NameKey: "name",
			TagKeys: []string{"tag1"},
		}
		err := parser.Init()
		return parser, err
	})

	testJSON := `{
		"name": "test1",
		"value": 100.1,
		"tag1": "value1"
	}`

	// Write json file to process into the 'process' directory.
	f, _ := os.CreateTemp(processDirectory, "test.json")
	_, _ = f.WriteString(testJSON)
	_ = f.Close()

	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	require.NoError(t, acc.FirstError())
	require.Len(t, acc.Metrics, 1)
	testutil.RequireMetricEqual(t, testutil.TestMetric(100.1), acc.GetTelegrafMetrics()[0], testutil.IgnoreTime())
}

func TestParseSubdirectories(t *testing.T) {
	acc := testutil.Accumulator{}

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		Recursive:          true,
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        "at-once",
	}
	err := r.Init()
	require.NoError(t, err)
	r.Log = testutil.Logger{}

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := &json.Parser{
			NameKey: "name",
			TagKeys: []string{"tag1"},
		}
		err := parser.Init()
		return parser, err
	})

	testJSON := `{
		"name": "test1",
		"value": 100.1,
		"tag1": "value1"
	}`

	// Write json file to process into the 'process' directory.
	testJSONFile := "test.json"
	f, err := os.Create(filepath.Join(processDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString(testJSON)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Write json file to process into a subdirectory in the the 'process' directory.
	err = os.Mkdir(filepath.Join(processDirectory, "sub"), os.ModePerm)
	require.NoError(t, err)
	f, err = os.Create(filepath.Join(processDirectory, "sub", testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString(testJSON)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(2)
	r.Stop()

	require.NoError(t, acc.FirstError())
	require.Len(t, acc.Metrics, 2)
	testutil.RequireMetricEqual(t, testutil.TestMetric(100.1), acc.GetTelegrafMetrics()[0], testutil.IgnoreTime())

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(finishedDirectory, "sub", testJSONFile))
	require.NoError(t, err)
}

func TestParseSubdirectoriesFilesIgnore(t *testing.T) {
	acc := testutil.Accumulator{}

	// Establish process directory and finished directory.
	finishedDirectory := t.TempDir()
	processDirectory := t.TempDir()

	filesToIgnore := `sub/test.json`
	if runtime.GOOS == "windows" {
		filesToIgnore = `\\sub\\test.json`
	}

	// Init plugin.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		Recursive:          true,
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        "at-once",
		FilesToIgnore:      []string{filesToIgnore},
	}
	err := r.Init()
	require.NoError(t, err)
	r.Log = testutil.Logger{}

	r.SetParserFunc(func() (parsers.Parser, error) {
		parser := &json.Parser{
			NameKey: "name",
			TagKeys: []string{"tag1"},
		}
		err := parser.Init()
		return parser, err
	})

	testJSON := `{
		"name": "test1",
		"value": 100.1,
		"tag1": "value1"
	}`

	// Write json file to process into the 'process' directory.
	testJSONFile := "test.json"
	f, err := os.Create(filepath.Join(processDirectory, testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString(testJSON)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Write json file to process into a subdirectory in the the 'process' directory.
	err = os.Mkdir(filepath.Join(processDirectory, "sub"), os.ModePerm)
	require.NoError(t, err)
	f, err = os.Create(filepath.Join(processDirectory, "sub", testJSONFile))
	require.NoError(t, err)
	_, err = f.WriteString(testJSON)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(1)
	r.Stop()

	require.NoError(t, acc.FirstError())
	require.Len(t, acc.Metrics, 1)
	testutil.RequireMetricEqual(t, testutil.TestMetric(100.1), acc.GetTelegrafMetrics()[0], testutil.IgnoreTime())

	// File should have gone back to the test directory, as we configured.
	_, err = os.Stat(filepath.Join(finishedDirectory, testJSONFile))
	require.NoError(t, err)
}
