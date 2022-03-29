package gcs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

const (
	singleObjectNotFound = "{\"error\":{\"code\":404,\"message\":\"No such object: test-iteration-bucket/prefix/offset-key.json\",\"errors\":[{\"message\":\"No such object: test-iteration-bucket/prefix/offset-key.json\",\"domain\":\"global\",\"reason\":\"notFound\"}]}}"
	singleFileList       = "{\"kind\":\"storage#objects\",\"items\":[{\"kind\":\"storage#object\",\"id\":\"test-iteration-bucket/1604148850990/1604148851295698\",\"selfLink\":\"https://www.googleapis.com/storage/v1/b/1604148850990/o/1604148850990\",\"mediaLink\":\"https://content-storage.googleapis.com/download/storage/v1/b/test-iteration-bucket/o/1604148850990?generation=1604148851295698&alt=media\",\"name\":\"1604148850990\",\"bucket\":\"test-iteration-bucket\",\"generation\":\"1604148851295698\",\"metageneration\":\"1\",\"contentType\":\"text/plain; charset=utf-8\",\"storageClass\":\"STANDARD\",\"size\":\"161\",\"md5Hash\":\"y59iuRCTpkm7wpvU5YHUYw==\",\"crc32c\":\"y57reA==\",\"etag\":\"CNKLy5Pw3uwCEAE=\",\"timeCreated\":\"2020-10-31T12:54:11.295Z\",\"updated\":\"2020-10-31T12:54:11.295Z\",\"timeStorageClassUpdated\":\"2020-10-31T12:54:11.295Z\"}]}"
	firstFile            = "{\"metrics\":[{\"fields\":{\"cosine\":10,\"sine\":-1.0975806427415925e-12},\"name\":\"cpu\",\"tags\":{\"datacenter\":\"us-east-1\",\"host\":\"localhost\"},\"timestamp\":1604148850991}]}"
	secondFile           = "{\"metrics\":[{\"fields\":{\"cosine\":11,\"sine\":-2.0975806427415925e-12},\"name\":\"cpu\",\"tags\":{\"datacenter\":\"us-east-1\",\"host\":\"localhost\"},\"timestamp\":1604148850992}]}"
	thirdFile            = "{\"metrics\":[{\"fields\":{\"cosine\":12,\"sine\":-3.0975806427415925e-12},\"name\":\"cpu\",\"tags\":{\"datacenter\":\"us-east-1\",\"host\":\"localhost\"},\"timestamp\":1604148850993}]}"
	fourthFile           = "{\"metrics\":[{\"fields\":{\"cosine\":13,\"sine\":-4.0975806427415925e-12},\"name\":\"cpu\",\"tags\":{\"datacenter\":\"us-east-1\",\"host\":\"localhost\"},\"timestamp\":1604148850994}]}"
	firstFileListing     = "{\"kind\":\"storage#object\",\"id\":\"test-iteration-bucket/prefix/1604148850991/1604148851353983\",\"selfLink\":\"https://www.googleapis.com/storage/v1/b/test-iteration-bucket/o/1604148850991\",\"mediaLink\":\"https://content-storage.googleapis.com/download/storage/v1/b/test-iteration-bucket/o/1604148850991?generation=1604148851353983&alt=media\",\"name\":\"prefix/1604148850991\",\"bucket\":\"test-iteration-bucket\",\"generation\":\"1604148851353983\",\"metageneration\":\"1\",\"contentType\":\"text/plain; charset=utf-8\",\"storageClass\":\"STANDARD\",\"size\":\"161\",\"md5Hash\":\"y59iuRCTpkm7wpvU5YHUYw==\",\"crc32c\":\"y57reA==\",\"etag\":\"CP/SzpPw3uwCEAE=\",\"timeCreated\":\"2020-10-31T12:54:11.353Z\",\"updated\":\"2020-10-31T12:54:11.353Z\",\"timeStorageClassUpdated\":\"2020-10-31T12:54:11.353Z\"}"
	secondFileListing    = "{\"kind\":\"storage#object\",\"id\":\"test-iteration-bucket/prefix/1604148850992/1604148851414237\",\"selfLink\":\"https://www.googleapis.com/storage/v1/b/test-iteration-bucket/o/1604148850992\",\"mediaLink\":\"https://content-storage.googleapis.com/download/storage/v1/b/test-iteration-bucket/o/1604148850992?generation=1604148851414237&alt=media\",\"name\":\"prefix/1604148850992\",\"bucket\":\"test-iteration-bucket\",\"generation\":\"1604148851414237\",\"metageneration\":\"1\",\"contentType\":\"text/plain; charset=utf-8\",\"storageClass\":\"STANDARD\",\"size\":\"161\",\"md5Hash\":\"y59iuRCTpkm7wpvU5YHUYw==\",\"crc32c\":\"y57reA==\",\"etag\":\"CN2p0pPw3uwCEAE=\",\"timeCreated\":\"2020-10-31T12:54:11.414Z\",\"updated\":\"2020-10-31T12:54:11.414Z\",\"timeStorageClassUpdated\":\"2020-10-31T12:54:11.414Z\"}"
	thirdFileListing     = "{\"kind\":\"storage#object\",\"id\":\"test-iteration-bucket/prefix/1604148850993/1604148851467554\",\"selfLink\":\"https://www.googleapis.com/storage/v1/b/test-iteration-bucket/o/1604148850993\",\"mediaLink\":\"https://content-storage.googleapis.com/download/storage/v1/b/test-iteration-bucket/o/1604148850993?generation=1604148851467554&alt=media\",\"name\":\"prefix/1604148850993\",\"bucket\":\"test-iteration-bucket\",\"generation\":\"1604148851467554\",\"metageneration\":\"1\",\"contentType\":\"text/plain; charset=utf-8\",\"storageClass\":\"STANDARD\",\"size\":\"161\",\"md5Hash\":\"y59iuRCTpkm7wpvU5YHUYw==\",\"crc32c\":\"y57reA==\",\"etag\":\"CKLK1ZPw3uwCEAE=\",\"timeCreated\":\"2020-10-31T12:54:11.467Z\",\"updated\":\"2020-10-31T12:54:11.467Z\",\"timeStorageClassUpdated\":\"2020-10-31T12:54:11.467Z\"}"
	fourthFileListing    = "{\"kind\":\"storage#object\",\"id\":\"test-iteration-bucket/prefix/1604148850994/1604148851467554\",\"selfLink\":\"https://www.googleapis.com/storage/v1/b/test-iteration-bucket/o/1604148850994\",\"mediaLink\":\"https://content-storage.googleapis.com/download/storage/v1/b/test-iteration-bucket/o/1604148850994?generation=1604148851467554&alt=media\",\"name\":\"prefix/1604148850994\",\"bucket\":\"test-iteration-bucket\",\"generation\":\"1604148851467554\",\"metageneration\":\"1\",\"contentType\":\"text/plain; charset=utf-8\",\"storageClass\":\"STANDARD\",\"size\":\"161\",\"md5Hash\":\"y59iuRCTpkm7wpvU5YHUYw==\",\"crc32c\":\"y57reA==\",\"etag\":\"CKLK1ZPw3uwCEAE=\",\"timeCreated\":\"2020-10-31T12:54:11.467Z\",\"updated\":\"2020-10-31T12:54:11.467Z\",\"timeStorageClassUpdated\":\"2020-10-31T12:54:11.467Z\"}"
	fileListing          = "{\"kind\":\"storage#objects\"}"
	offSetTemplate       = "{\"offSet\":\"%s\"}"
)

var objListing = parseJSONFromText(fileListing)
var firstElement = parseJSONFromText(firstFileListing)
var secondElement = parseJSONFromText(secondFileListing)
var thirdElement = parseJSONFromText(thirdFileListing)
var fourthElement = parseJSONFromText(fourthFileListing)

func TestRunSetUpClient(t *testing.T) {
	gcs := &GCS{
		Project:   "test-project",
		Bucket:    "test-bucket",
		Prefix:    "prefix",
		OffsetKey: "1230405",
		Log:       testutil.Logger{},
	}

	require.Error(t, gcs.setUpClient())
}

func TestRunInit(t *testing.T) {
	srv := startGCSServer(t)
	defer srv.Close()

	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	gcs := &GCS{
		Project:   "test-project",
		Bucket:    "test-bucket",
		Prefix:    "prefix/",
		OffsetKey: "offset.json",
		Log:       testutil.Logger{},
	}

	require.NoError(t, gcs.Init())

	assert.Equal(t, "offsetfile", gcs.offSet.OffSet)
}

func TestRunInitNoOffsetKey(t *testing.T) {
	srv := startGCSServer(t)
	defer srv.Close()

	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	gcs := &GCS{
		Project: "test-project",
		Bucket:  "test-bucket",
		Prefix:  "prefix/",
		Log:     testutil.Logger{},
	}

	require.NoError(t, gcs.Init())

	assert.Equal(t, "offsetfile", gcs.offSet.OffSet)
	assert.Equal(t, "prefix/offset-key.json", gcs.OffsetKey)
}

func TestRunGatherOneItem(t *testing.T) {
	srv := startOneItemGCSServer(t)
	defer srv.Close()
	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	acc := &testutil.Accumulator{}

	gcs := &GCS{
		Project: "test-project",
		Bucket:  "test-iteration-bucket",
		Prefix:  "prefix/",
		Log:     testutil.Logger{},
		parser:  createParser(),
	}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	metric := acc.Metrics[0]
	assert.Equal(t, "cpu", metric.Measurement)
	assert.Equal(t, "us-east-1", metric.Tags["tags_datacenter"])
	assert.Equal(t, "localhost", metric.Tags["tags_host"])
	assert.Equal(t, 10.0, metric.Fields["fields_cosine"])
	assert.Equal(t, -1.0975806427415925e-12, metric.Fields["fields_sine"])
}

func TestRunGatherOneIteration(t *testing.T) {
	srv := startMultipleItemGCSServer(t)
	defer srv.Close()

	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	gcs := &GCS{
		Project:   "test-project",
		Bucket:    "test-iteration-bucket",
		Prefix:    "prefix/",
		OffsetKey: "custom-offset-key.json",
		Log:       testutil.Logger{},
		parser:    createParser(),
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	assert.Equal(t, 3, len(acc.Metrics))
}

func TestRunGatherIteratiosnWithLimit(t *testing.T) {
	srv := startMultipleItemGCSServer(t)
	defer srv.Close()

	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	gcs := &GCS{
		Project:             "test-project",
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

	assert.Equal(t, 1, len(acc.Metrics))
	require.NoError(t, gcs.Gather(acc))

	assert.Equal(t, 2, len(acc.Metrics))
	require.NoError(t, gcs.Gather(acc))

	assert.Equal(t, 3, len(acc.Metrics))
}

func TestRunGatherIterationWithPages(t *testing.T) {
	srv := stateFulGCSServer(t)
	defer srv.Close()

	os.Setenv("STORAGE_EMULATOR_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	gcs := &GCS{
		Project:   "test-project",
		Bucket:    "test-iteration-bucket",
		Prefix:    "prefix/",
		OffsetKey: "custom-offset-key.json",
		Log:       testutil.Logger{},
		parser:    createParser(),
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, gcs.Init())

	require.NoError(t, gcs.Gather(acc))

	assert.Equal(t, 4, len(acc.Metrics))
	assert.Equal(t, true, gcs.offSet.isPresent())
	assert.Equal(t, "prefix/1604148850994", gcs.offSet.OffSet)

	emptyAcc := &testutil.Accumulator{}
	require.NoError(t, gcs.Gather(emptyAcc))

	assert.Equal(t, 0, len(emptyAcc.Metrics))
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

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b/test-iteration-bucket/o":
			serveJSONText(w, singleFileList)
		default:
			serveBlobs(r.URL.Path, "", t, w)
		}
	})

	return srv
}

func startMultipleItemGCSServer(t *testing.T) *httptest.Server {
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
			serveBlobs(r.URL.Path, currentOffSetKey, t, w)
		}
	})

	return srv
}

func stateFulGCSServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.NotFoundHandler())

	currentOffSetKey := fmt.Sprintf(offSetTemplate, "prefix/1604148850990")

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/b/test-iteration-bucket/o":
			offset := r.URL.Query().Get("startOffset")
			objListing := parseJSONFromText(fileListing)

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
			serveBlobs(r.URL.Path, currentOffSetKey, t, w)
		}
	})

	return srv
}

func serveBlobs(urlPath string, offsetKey string, t *testing.T, w http.ResponseWriter) {
	switch urlPath {
	case "/test-iteration-bucket/prefix/offset-key.json":
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(singleObjectNotFound))
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
	bodyBytes, err := ioutil.ReadAll(rc)

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

func serveJSONText(w http.ResponseWriter, jsonText string) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(jsonText)); err != nil {
		fmt.Println(err)
	}
}

func failPath(path string, t *testing.T, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	t.Fatalf("unexpected path: " + path)
}

func parseJSONFromText(jsonText string) map[string]interface{} {
	var element map[string]interface{}
	if err := json.Unmarshal([]byte(jsonText), &element); err != nil {
		fmt.Println(err)
	}

	return element
}
