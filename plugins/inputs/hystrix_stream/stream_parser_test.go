package hystrix_stream

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
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

func Test_fill_cache(t *testing.T) {

	file, fileErr := os.Open("testdata/dummy_stream_entries.txt")

	if fileErr != nil {
		t.Fatal("Could not open testfile")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	fillCacheForeverMax(scanner, 181)

	if len(cachedEntries) != 181 {
		t.Errorf("Expected to have read 181 entries, read %d", len(cachedEntries))
	}
}

func local_Test_stream_entries_locally(t *testing.T) {

	_, err := latestEntries("http://localhost:8090/hystrix")

	if err != nil {
		t.Fatalf("Error on first read : %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	entries, err2 := latestEntries("http://localhost:8090/hystrix")

	if err2 != nil {
		t.Fatalf("Error on first read : %v", err2)
	}

	if len(entries) == 0 {
		t.Error("Expected more than zero entries")
	}

	fmt.Printf("Got %d entries, cached is %d", len(entries), len(cachedEntries))
}
