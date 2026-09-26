package directory_monitor

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestCreator(t *testing.T) {
	creator, found := inputs.Inputs["directory_monitor"]
	require.True(t, found)

	expected := &DirectoryMonitor{
		MaxBufferedMetrics:         defaultMaxBufferedMetrics,
		DirectoryDurationThreshold: defaultDirectoryDurationThreshold,
		FileQueueSize:              defaultFileQueueSize,
		ParseMethod:                defaultParseMethod,
		FileAction:                 defaultFileAction,
	}

	require.Equal(t, expected, creator())
}

func TestDeleteFileAfterProcessing(t *testing.T) {
	acc := testutil.Accumulator{}
	testCsvFile := "test.csv"

	// Establish process, finished, and error directories.
	// Finished is configured so we can prove delete mode does not write a copy there.
	processDirectory := t.TempDir()
	finishedDirectory := t.TempDir()
	errorDirectory := t.TempDir()

	// Init plugin in delete mode.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		ErrorDirectory:     errorDirectory,
		FileAction:         "delete",
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (telegraf.Parser, error) {
		parser := csv.Parser{
			HeaderRowCount: 1,
		}
		err := parser.Init()
		return &parser, err
	})
	r.Log = testutil.Logger{}

	// Write a valid csv file into the monitored directory.
	f, err := os.Create(filepath.Join(processDirectory, testCsvFile))
	require.NoError(t, err)
	_, err = f.WriteString("thing,color\nsky,blue\ngrass,green\n")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	// Start plugin and process the file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(2)
	r.Stop()

	// Both data rows should have been parsed.
	require.Len(t, acc.Metrics, 2)

	// The original file is removed so it will not be scanned again.
	_, err = os.Stat(filepath.Join(processDirectory, testCsvFile))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	// Delete mode does not copy the file into the finished directory.
	_, err = os.Stat(filepath.Join(finishedDirectory, testCsvFile))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	// A successful file is not moved to the error directory either.
	_, err = os.Stat(filepath.Join(errorDirectory, testCsvFile))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestDeleteModeMovesFailedFile(t *testing.T) {
	acc := testutil.Accumulator{}
	testJSONFile := "broken.json"

	// Establish process and error directories. Finished is set but unused for failures.
	processDirectory := t.TempDir()
	finishedDirectory := t.TempDir()
	errorDirectory := t.TempDir()

	// Init plugin in delete mode with an error directory.
	r := DirectoryMonitor{
		Directory:          processDirectory,
		FinishedDirectory:  finishedDirectory,
		ErrorDirectory:     errorDirectory,
		FileAction:         "delete",
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	err := r.Init()
	require.NoError(t, err)

	r.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{NameKey: "Name"}
		err := p.Init()
		return p, err
	})
	r.Log = testutil.Logger{}

	// Write JSON that the parser cannot consume.
	err = os.WriteFile(filepath.Join(processDirectory, testJSONFile), []byte("this is not json"), 0640)
	require.NoError(t, err)

	// Start plugin and attempt to process the file.
	// Parsing fails, so no metrics arrive. Wait until the file is moved aside.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		_, statErr := os.Stat(filepath.Join(errorDirectory, testJSONFile))
		return statErr == nil
	}, 5*time.Second, 10*time.Millisecond)
	r.Stop()

	// Nothing was ingested.
	require.Empty(t, acc.Metrics)

	// The original is removed from the monitored directory.
	_, err = os.Stat(filepath.Join(processDirectory, testJSONFile))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	// Failed files are still copied to the error directory.
	_, err = os.Stat(filepath.Join(errorDirectory, testJSONFile))
	require.NoError(t, err)

	// Failures are not copied to the finished directory.
	_, err = os.Stat(filepath.Join(finishedDirectory, testJSONFile))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestDeleteModeDoesNotRequireFinishedDirectory(t *testing.T) {
	// Delete mode ignores finished_directory, so it may be left empty.
	r := DirectoryMonitor{
		Directory:          t.TempDir(),
		FileAction:         "delete",
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	require.NoError(t, r.Init())
}

func TestMoveModeRequiresFinishedDirectory(t *testing.T) {
	// The default action is move, which still needs a destination directory.
	r := DirectoryMonitor{
		Directory:          t.TempDir(),
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	require.Error(t, r.Init())
}

func TestInvalidFileAction(t *testing.T) {
	r := DirectoryMonitor{
		Directory:          t.TempDir(),
		FinishedDirectory:  t.TempDir(),
		FileAction:         "archive",
		MaxBufferedMetrics: defaultMaxBufferedMetrics,
		FileQueueSize:      defaultFileQueueSize,
		ParseMethod:        defaultParseMethod,
	}
	require.Error(t, r.Init())
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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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
	err = os.WriteFile(filepath.Join(processDirectory, testCsvGzFile), b.Bytes(), 0640)
	require.NoError(t, err)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(6)
	r.Stop()

	// Verify that we read both files once.
	require.Len(t, acc.Metrics, 6)

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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
	err = os.WriteFile(filepath.Join(processDirectory, testCsvGzFile), b.Bytes(), 0640)
	require.NoError(t, err)

	// Start plugin before adding file.
	err = r.Start(&acc)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(6)
	r.Stop()

	// Verify that we read both files once.
	require.Len(t, acc.Metrics, 6)

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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
	require.Len(t, acc.Metrics, 5)
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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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
	require.Len(t, acc.Metrics, 1)
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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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

	expectedFields := map[string]any{
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
	require.Len(t, acc.Metrics, 1)

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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

	expectedFields := map[string]any{
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
	require.Len(t, acc.Metrics, 1)

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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

	expectedFields := map[string]any{
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
	require.Len(t, acc.Metrics, 1)

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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
	f, err := os.CreateTemp(processDirectory, "test.json")
	require.NoError(t, err)
	_, err = f.WriteString(testJSON)
	require.NoError(t, err)
	f.Close()

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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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

	// Write json file to process into a subdirectory in the 'process' directory.
	err = os.Mkdir(filepath.Join(processDirectory, "sub"), 0750)
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

	r.SetParserFunc(func() (telegraf.Parser, error) {
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

	// Write json file to process into a subdirectory in the 'process' directory.
	err = os.Mkdir(filepath.Join(processDirectory, "sub"), 0750)
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

func TestPreserveTimestamps(t *testing.T) {
	testJSONFile := "test.json"
	originalTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	runMonitor := func(t *testing.T, preserve bool) os.FileInfo {
		t.Helper()
		acc := testutil.Accumulator{}
		finishedDirectory := t.TempDir()
		processDirectory := t.TempDir()

		r := DirectoryMonitor{
			Directory:          processDirectory,
			FinishedDirectory:  finishedDirectory,
			PreserveTimestamps: preserve,
			MaxBufferedMetrics: defaultMaxBufferedMetrics,
			FileQueueSize:      defaultFileQueueSize,
			ParseMethod:        defaultParseMethod,
		}
		require.NoError(t, r.Init())
		r.SetParserFunc(func() (telegraf.Parser, error) {
			p := &json.Parser{NameKey: "Name"}
			err := p.Init()
			return p, err
		})

		srcPath := filepath.Join(processDirectory, testJSONFile)
		require.NoError(t, os.WriteFile(srcPath, []byte(`{"Name": "event1", "Speed": 100.1}`), 0640))
		require.NoError(t, os.Chtimes(srcPath, originalTime, originalTime))

		r.Log = testutil.Logger{}
		require.NoError(t, r.Start(&acc))
		require.NoError(t, r.Gather(&acc))
		acc.Wait(1)
		r.Stop()

		info, err := os.Stat(filepath.Join(finishedDirectory, testJSONFile))
		require.NoError(t, err)
		return info
	}

	t.Run("enabled keeps the original modification time", func(t *testing.T) {
		info := runMonitor(t, true)
		require.True(t, info.ModTime().Equal(originalTime), "expected %s, got %s", originalTime, info.ModTime())
	})

	t.Run("disabled uses the move time", func(t *testing.T) {
		info := runMonitor(t, false)
		require.False(t, info.ModTime().Equal(originalTime), "expected the move time, got the original timestamp")
	})
}
