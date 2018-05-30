package hystrix_stream

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

// HystrixStreamEntry is 1 entry in the stream from the metrics-stream-servlet
type HystrixStreamEntry struct {
	Type                               string `json:"type"`
	Name                               string `json:"name"`
	Group                              string `json:"group"`
	CurrentTime                        int64  `json:"currentTime"`
	IsCircuitBreakerOpen               bool   `json:"isCircuitBreakerOpen"`
	ErrorPercentage                    int    `json:"errorPercentage"`
	ErrorCount                         int    `json:"errorCount"`
	RequestCount                       int    `json:"requestCount"`
	RollingCountBadRequests            int    `json:"rollingCountBadRequests"`
	RollingCountCollapsedRequests      int    `json:"rollingCountCollapsedRequests"`
	RollingCountEmit                   int    `json:"rollingCountEmit"`
	RollingCountExceptionsThrown       int    `json:"rollingCountExceptionsThrown"`
	RollingCountFailure                int    `json:"rollingCountFailure"`
	RollingCountFallbackFailure        int    `json:"rollingCountFallbackFailure"`
	RollingCountFallbackRejection      int    `json:"rollingCountFallbackRejection"`
	RollingCountFallbackSuccess        int    `json:"rollingCountFallbackSuccess"`
	RollingCountResponsesFromCache     int    `json:"rollingCountResponsesFromCache"`
	RollingCountSemaphoreRejected      int    `json:"rollingCountSemaphoreRejected"`
	RollingCountShortCircuited         int    `json:"rollingCountShortCircuited"`
	RollingCountSuccess                int    `json:"rollingCountSuccess"`
	RollingCountThreadPoolRejected     int    `json:"rollingCountThreadPoolRejected"`
	RollingCountTimeout                int    `json:"rollingCountTimeout"`
	CurrentConcurrentExecutionCount    int    `json:"currentConcurrentExecutionCount"`
	RollingMaxConcurrentExecutionCount int    `json:"rollingMaxConcurrentExecutionCount"`
	LatencyExecute                     struct {
		Num0   int `json:"0"`
		Num25  int `json:"25"`
		Num50  int `json:"50"`
		Num75  int `json:"75"`
		Num90  int `json:"90"`
		Num95  int `json:"95"`
		Num99  int `json:"99"`
		Num100 int `json:"100"`
		Nine95 int `json:"99.5"`
	} `json:"latencyExecute"`
	LatencyTotal struct {
		Num0   int `json:"0"`
		Num25  int `json:"25"`
		Num50  int `json:"50"`
		Num75  int `json:"75"`
		Num90  int `json:"90"`
		Num95  int `json:"95"`
		Num99  int `json:"99"`
		Num100 int `json:"100"`
		Nine95 int `json:"99.5"`
	} `json:"latencyTotal"`
	ReportingHosts int    `json:"reportingHosts"`
	ThreadPool     string `json:"threadPool"`
}

var (
	healthy       = false
	scanner       *bufio.Scanner
	cachedEntries []HystrixStreamEntry
	reader        io.ReadCloser
	cacheLock     sync.Mutex
)

func latestEntries(url string) ([]HystrixStreamEntry, error) {

	if !healthy {
		resp, err := http.Get(url)
		if err != nil {
			return make([]HystrixStreamEntry, 0), err
		}
		scanner = bufio.NewScanner(resp.Body)
		reader = resp.Body
		cachedEntries = make([]HystrixStreamEntry, 0)
		go fillCacheForever(scanner)
		healthy = true
	}

	if scanner.Err() != nil {
		log.Printf("E! Error scanning hystrix-servlet: [%v]", scanner.Err())
		reader.Close()
		healthy = false
		return make([]HystrixStreamEntry, 0), scanner.Err()
	}

	defer clearCache()
	return cachedEntries, nil
}

func clearCache() {
	cacheLock.Lock()
	cachedEntries = cachedEntries[:0]
	cacheLock.Unlock()
}

func fillCacheForever(scanner *bufio.Scanner) {
	fillCacheForeverMax(scanner, 100000)
}

func fillCacheForeverMax(scanner *bufio.Scanner, maxEntries int) {
	newEntryCounter := 0

	for scanner.Err() == nil {
		chunks := streamToStrings(scanner)
		entries, err := parseChunk(chunks)
		if err == nil {
			for _, entry := range entries {
				cacheLock.Lock()
				cachedEntries = append(cachedEntries, entry)
				newEntryCounter++
				cacheLock.Unlock()
			}
		}
		if newEntryCounter >= maxEntries {
			return
		}
	}
}

func streamToStrings(scanner *bufio.Scanner) string {
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			return text
		}
	}
	return ""
}

func parseChunk(streamChunk string) ([]HystrixStreamEntry, error) {

	entries := make([]HystrixStreamEntry, 0)
	for _, line := range strings.Split(streamChunk, "\n") {
		if strings.Contains(line, "data:") {
			entryPartOfLine := strings.SplitAfter(line, "data:")
			if len(entryPartOfLine) == 2 {
				entry := HystrixStreamEntry{}
				jsonErr := json.Unmarshal([]byte(entryPartOfLine[1]), &entry)
				if jsonErr != nil {
					return entries, jsonErr
				} else if entry.Type == "HystrixCommand" {
					entries = append(entries, entry)
				}
			}
		}
	}

	return entries, nil
}
