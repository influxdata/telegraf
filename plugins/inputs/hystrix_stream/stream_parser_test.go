package hystrix_stream

import (
	"io/ioutil"
	"testing"
	"os"
	"net/http"
)

func Test_parse_from_file(t *testing.T) {
	bytes, readErr := ioutil.ReadFile("testdata/dummy_stream_entries.txt")

	if readErr != nil {
		t.Fatal("Could not read testdata")
	}

	data := string(bytes)

	entries, err := parseChunk(data)

	if err != nil {
		t.Errorf("Got error when parsing multiple lines from chunk: %v", err)
	}

	if len(entries) != 181 {
		t.Errorf("Expected 181 entries, got %d", len(entries))
	}
}

func Test_stream_entries(t *testing.T) {

	file, fileErr := os.Open("testdata/dummy_stream_entries.txt")

	if fileErr != nil {
		t.Fatal("Could not open testfile")
	}

	defer file.Close()

	entryChan, stop := entryStream(file, 181)


	stopped := false
	entryCount := 0
	for ; !stopped; {
		select {
		case <-stop:
			stopped = true;
		case <-entryChan:
			entryCount++
		}
	}

	if entryCount != 181 {
		t.Errorf("Expected to have read 181 entries, read %d", entryCount)
	}
}

func local_Test_stream_entries_locally(t *testing.T) {

	response, httpGetError := http.Get("http://localhost:8090/hystrix")

	if httpGetError != nil {
		t.Fatalf("Could not open url %v", httpGetError)
	}

	defer response.Body.Close()

	entryChan, stop := entryStream(response.Body, 10)


	stopped := false
	entryCount := 0
	for ; !stopped; {
		select {
		case <-stop:
			stopped = true;
		case <-entryChan:
			entryCount++
			println(entryCount)
		}
	}

}
