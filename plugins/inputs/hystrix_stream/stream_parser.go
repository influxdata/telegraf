package hystrix_stream

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"strings"
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

func (s *HystrixData) latestEntries() ([]HystrixStreamEntry, error) {

	if !s.healthy {
		resp, err := http.Get(s.Url)
		if err != nil {
			return make([]HystrixStreamEntry, 0), err
		}
		s.scanner = bufio.NewScanner(resp.Body)
		s.reader = resp.Body
		s.cachedEntries = make([]HystrixStreamEntry, 0)
		go s.fillCacheForever(s.scanner)
		s.healthy = true
	}

	if s.scanner.Err() != nil {
		log.Printf("E! Error scanning hystrix-servlet: [%v]", s.scanner.Err())
		s.reader.Close()
		s.healthy = false
		return make([]HystrixStreamEntry, 0), s.scanner.Err()
	}

	defer s.clearCache()
	return s.cachedEntries, nil
}

func (s *HystrixData) clearCache() {
	s.cacheLock.Lock()
	s.cachedEntries = s.cachedEntries[:0]
	s.cacheLock.Unlock()
}

func (s *HystrixData) fillCacheForever(scanner *bufio.Scanner) {
	s.fillCacheForeverMax(scanner, 100000)
}

func (s *HystrixData) fillCacheForeverMax(scanner *bufio.Scanner, maxEntries int) {
	newEntryCounter := 0

	for scanner.Err() == nil {
		entry := firstNonEmptyLine(scanner)
		entries, err := parseChunk(entry)
		if err == nil {
			for _, entry := range entries {
				s.cacheLock.Lock()
				s.cachedEntries = append(s.cachedEntries, entry)
				newEntryCounter++
				s.cacheLock.Unlock()
			}
		}
		if newEntryCounter >= maxEntries {
			return
		}
	}
}

func firstNonEmptyLine(scanner *bufio.Scanner) string {
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
