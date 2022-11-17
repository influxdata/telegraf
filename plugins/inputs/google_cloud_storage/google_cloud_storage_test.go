package gcs

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/influxdata/telegraf/testutil"
)

const offSetTemplate = "{\"offSet\":\"%s\"}"

func TestRunSetUpClient(t *testing.T) {
	gcs := &GCS{
		Bucket:    "test-bucket",
		Prefix:    "prefix",
		OffsetKey: "1230405",
		Log:       testutil.Logger{},
	}

	if err := gcs.setUpClient(); err != nil {
		t.Log(err)
	}
}

func TestRunInit(t *testing.T) {
	srv := startGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	gcs := &GCS{
		Bucket:    "test-bucket",
		Prefix:    "prefix/",
		OffsetKey: "offset.json",
		Log:       testutil.Logger{},
	}

	require.NoError(t, gcs.Init())

	require.Equal(t, "offsetfile", gcs.offSet.OffSet)
}

func TestRunInitNoOffsetKey(t *testing.T) {
	srv := startGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	gcs := &GCS{
		Bucket: "test-bucket",
		Prefix: "prefix/",
		Log:    testutil.Logger{},
	}

	require.NoError(t, gcs.Init())

	require.Equal(t, "offsetfile", gcs.offSet.OffSet)
	require.Equal(t, "prefix/offset-key.json", gcs.OffsetKey)
}

func TestRunGatherOneItem(t *testing.T) {
	srv := startOneItemGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	acc := &testutil.Accumulator{}

	gcs := &GCS{
		Bucket: "test-iteration-bucket",
		Prefix: "prefix/",
		Log:    testutil.Logger{},
		parser: createParser(),
	}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	metric := acc.Metrics[0]
	require.Equal(t, "cpu", metric.Measurement)
	require.Equal(t, "us-east-1", metric.Tags["tags_datacenter"])
	require.Equal(t, "localhost", metric.Tags["tags_host"])
	require.Equal(t, 10.0, metric.Fields["fields_cosine"])
	require.Equal(t, -1.0975806427415925e-12, metric.Fields["fields_sine"])
}

func TestRunGatherOneIteration(t *testing.T) {
	srv := startMultipleItemGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	gcs := &GCS{
		Bucket:    "test-iteration-bucket",
		Prefix:    "prefix/",
		OffsetKey: "custom-offset-key.json",
		Log:       testutil.Logger{},
		parser:    createParser(),
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	require.Equal(t, 3, len(acc.Metrics))
}

func TestRunGatherIteratiosnWithLimit(t *testing.T) {
	srv := startMultipleItemGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	gcs := &GCS{
		Bucket:              "test-iteration-bucket",
		Prefix:              "prefix/",
		ObjectsPerIteration: 1,
		OffsetKey:           "custom-offset-key.json",
		Log:                 testutil.Logger{},
		parser:              createParser(),
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	require.Equal(t, 1, len(acc.Metrics))
	require.NoError(t, gcs.Gather(acc))

	require.Equal(t, 2, len(acc.Metrics))
	require.NoError(t, gcs.Gather(acc))

	require.Equal(t, 3, len(acc.Metrics))
}

func TestRunGatherIterationWithPages(t *testing.T) {
	srv := stateFulGCSServer(t)
	defer srv.Close()

	emulatorSetEnv(t, srv)

	gcs := &GCS{
		Bucket:    "test-iteration-bucket",
		Prefix:    "prefix/",
		OffsetKey: "custom-offset-key.json",
		Log:       testutil.Logger{},
		parser:    createParser(),
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	require.Equal(t, 4, len(acc.Metrics))
	require.Equal(t, true, gcs.offSet.isPresent())
	require.Equal(t, "prefix/1604148850994", gcs.offSet.OffSet)

	emptyAcc := &testutil.Accumulator{}
	require.NoError(t, gcs.Gather(emptyAcc))

	require.Equal(t, 0, len(emptyAcc.Metrics))
}

func createParser() parsers.Parser {
	testParser, _ := parsers.NewParser(&parsers.Config{
		DataFormat:     "json",
		MetricName:     "cpu",
		JSONQuery:      "metrics",
		TagKeys:        []string{"tags_datacenter", "tags_host"},
		JSONTimeKey:    "timestamp",
		JSONTimeFormat: "unix_ms",
	})

	return testParser
}

func startGCSServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.NotFoundHandler())

	currentOffSetKey := fmt.Sprintf(offSetTemplate, "offsetfile")

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test-bucket/prefix/offset.json":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(currentOffSetKey))
			require.NoError(t, err)
		case "/test-bucket/prefix/offset-key.json":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("{\"offSet\":\"offsetfile\"}"))
			require.NoError(t, err)
		default:
			failPath(r.URL.Path, t, w)
		}
	})

	return srv
}

func startOneItemGCSServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.NotFoundHandler())
	singleFileList := readJSON(t, "testdata/single_file_list.json")

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b/test-iteration-bucket/o":
			serveJSONText(w, singleFileList)
		default:
			serveBlobs(t, w, r.URL.Path, "")
		}
	})

	return srv
}

func startMultipleItemGCSServer(t *testing.T) *httptest.Server {
	objListing := parseJSONFromFile(t, "testdata/file_listing.json")
	firstElement := parseJSONFromFile(t, "testdata/first_file_listing.json")
	secondElement := parseJSONFromFile(t, "testdata/second_file_listing.json")
	thirdElement := parseJSONFromFile(t, "testdata/third_file_listing.json")
	fourthElement := parseJSONFromFile(t, "testdata/fourth_file_listing.json")

	srv := httptest.NewServer(http.NotFoundHandler())

	currentOffSetKey := fmt.Sprintf(offSetTemplate, "prefix/1604148850991")

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b/test-iteration-bucket/o":

			offset := r.URL.Query().Get("startOffset")

			if offset == "prefix/1604148850990" {
				objListing["items"] = []interface{}{firstElement, secondElement, thirdElement, fourthElement}
			} else if offset == "prefix/1604148850991" {
				objListing["items"] = []interface{}{secondElement, thirdElement, fourthElement}
			} else if offset == "prefix/16041488509912" {
				objListing["items"] = []interface{}{thirdElement, fourthElement}
			} else if offset == "prefix/16041488509913" {
				objListing["items"] = []interface{}{thirdElement, fourthElement}
			} else {
				objListing["items"] = []interface{}{firstElement, secondElement, thirdElement, fourthElement}
			}

			if data, err := json.Marshal(objListing); err == nil {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(data)
				require.NoError(t, err)
			} else {
				w.WriteHeader(http.StatusNotFound)
				t.Fatalf("unexpected path: " + r.URL.Path)
			}

		default:
			serveBlobs(t, w, r.URL.Path, currentOffSetKey)
		}
	})

	return srv
}

func stateFulGCSServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.NotFoundHandler())

	firstElement := parseJSONFromFile(t, "testdata/first_file_listing.json")
	secondElement := parseJSONFromFile(t, "testdata/second_file_listing.json")
	thirdElement := parseJSONFromFile(t, "testdata/third_file_listing.json")
	fourthElement := parseJSONFromFile(t, "testdata/fourth_file_listing.json")
	currentOffSetKey := fmt.Sprintf(offSetTemplate, "prefix/1604148850990")

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b/test-iteration-bucket/o":
			offset := r.URL.Query().Get("startOffset")
			objListing := parseJSONFromFile(t, "testdata/file_listing.json")

			pageToken := r.URL.Query().Get("pageToken")

			if pageToken == "page2" {
				objListing["items"] = []interface{}{secondElement}
				objListing["nextPageToken"] = "page3"
			} else if pageToken == "page3" {
				objListing["items"] = []interface{}{thirdElement}
				objListing["nextPageToken"] = "page4"
			} else if pageToken == "page4" {
				objListing["items"] = []interface{}{fourthElement}
			} else if offset == "prefix/1604148850994" {
				objListing["items"] = []interface{}{}
			} else {
				objListing["items"] = []interface{}{firstElement}
				objListing["nextPageToken"] = "page2"
			}

			if data, err := json.Marshal(objListing); err == nil {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(data)
				require.NoError(t, err)
			} else {
				failPath(r.URL.Path, t, w)
			}
		case "/upload/storage/v1/b/test-iteration-bucket/o":
			_, params, _ := mime.ParseMediaType(r.Header["Content-Type"][0])
			boundary := params["boundary"]
			currentOffSetKey, _ = fetchJSON(t, boundary, r.Body)
		default:
			serveBlobs(t, w, r.URL.Path, currentOffSetKey)
		}
	})

	return srv
}

func serveBlobs(t *testing.T, w http.ResponseWriter, urlPath string, offsetKey string) {
	singleObjectNotFound := readJSON(t, "testdata/single_object_not_found.json")
	firstFile := readJSON(t, "testdata/first_file.json")
	secondFile := readJSON(t, "testdata/second_file.json")
	thirdFile := readJSON(t, "testdata/third_file.json")
	fourthFile := readJSON(t, "testdata/fourth_file.json")

	switch urlPath {
	case "/test-iteration-bucket/prefix/offset-key.json":
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write(singleObjectNotFound)
		require.NoError(t, err)
	case "/test-bucket/prefix/offset.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(offsetKey))
		require.NoError(t, err)
	case "/test-bucket/prefix/offset-key.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("{\"offSet\":\"offsetfile\"}"))
		require.NoError(t, err)
	case "/test-iteration-bucket/prefix/custom-offset-key.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(offsetKey))
		require.NoError(t, err)
	case "/test-iteration-bucket/1604148850990":
		serveJSONText(w, firstFile)
	case "/test-iteration-bucket/prefix/1604148850991":
		serveJSONText(w, firstFile)
	case "/test-iteration-bucket/prefix/1604148850992":
		serveJSONText(w, secondFile)
	case "/test-iteration-bucket/prefix/1604148850993":
		serveJSONText(w, thirdFile)
	case "/test-iteration-bucket/prefix/1604148850994":
		serveJSONText(w, fourthFile)
	case "/upload/storage/v1/b/test-iteration-bucket/o":
		w.WriteHeader(http.StatusOK)
	default:
		failPath(urlPath, t, w)
	}
}

func fetchJSON(t *testing.T, boundary string, rc io.ReadCloser) (string, error) {
	defer rc.Close()
	bodyBytes, err := io.ReadAll(rc)

	if err != nil {
		t.Fatalf("Could not read bytes from offset action")
		return "", err
	}

	splits := strings.Split(string(bodyBytes), boundary)
	offsetPart := splits[2]
	offsets := strings.Split(offsetPart, "\n")
	fmt.Printf("%s", offsets[3])
	return offsets[3], nil
}

func serveJSONText(w http.ResponseWriter, jsonText []byte) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonText); err != nil {
		fmt.Println(err)
	}
}

func failPath(path string, t *testing.T, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	t.Fatalf("unexpected path: " + path)
}

func parseJSONFromFile(t *testing.T, jsonFilePath string) map[string]interface{} {
	data := readJSON(t, jsonFilePath)

	var element map[string]interface{}
	if err := json.Unmarshal(data, &element); err != nil {
		require.NoErrorf(t, err, "could not parse from data file %s", jsonFilePath)
	}

	return element
}

func readJSON(t *testing.T, jsonFilePath string) []byte {
	data, err := os.ReadFile(jsonFilePath)
	require.NoErrorf(t, err, "could not read from data file %s", jsonFilePath)

	return data
}

func emulatorSetEnv(t *testing.T, srv *httptest.Server) {
	if err := os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", "")); err != nil {
		t.Error(err)
	}
}
