package hystrix_stream

import (
	"bufio"
	"encoding/json"
	"io"
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
	LatencyExecuteMean                 int    `json:"latencyExecute_mean"`
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
	LatencyTotalMean int `json:"latencyTotal_mean"`
	LatencyTotal     struct {
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
	PropertyValueCircuitBreakerRequestVolumeThreshold             int         `json:"propertyValue_circuitBreakerRequestVolumeThreshold"`
	PropertyValueCircuitBreakerSleepWindowInMilliseconds          int         `json:"propertyValue_circuitBreakerSleepWindowInMilliseconds"`
	PropertyValueCircuitBreakerErrorThresholdPercentage           int         `json:"propertyValue_circuitBreakerErrorThresholdPercentage"`
	PropertyValueCircuitBreakerForceOpen                          bool        `json:"propertyValue_circuitBreakerForceOpen"`
	PropertyValueCircuitBreakerForceClosed                        bool        `json:"propertyValue_circuitBreakerForceClosed"`
	PropertyValueCircuitBreakerEnabled                            bool        `json:"propertyValue_circuitBreakerEnabled"`
	PropertyValueExecutionIsolationStrategy                       string      `json:"propertyValue_executionIsolationStrategy"`
	PropertyValueExecutionIsolationThreadTimeoutInMilliseconds    int         `json:"propertyValue_executionIsolationThreadTimeoutInMilliseconds"`
	PropertyValueExecutionTimeoutInMilliseconds                   int         `json:"propertyValue_executionTimeoutInMilliseconds"`
	PropertyValueExecutionIsolationThreadInterruptOnTimeout       bool        `json:"propertyValue_executionIsolationThreadInterruptOnTimeout"`
	PropertyValueExecutionIsolationThreadPoolKeyOverride          interface{} `json:"propertyValue_executionIsolationThreadPoolKeyOverride"`
	PropertyValueExecutionIsolationSemaphoreMaxConcurrentRequests int         `json:"propertyValue_executionIsolationSemaphoreMaxConcurrentRequests"`
	PropertyValueFallbackIsolationSemaphoreMaxConcurrentRequests  int         `json:"propertyValue_fallbackIsolationSemaphoreMaxConcurrentRequests"`
	PropertyValueMetricsRollingStatisticalWindowInMilliseconds    int         `json:"propertyValue_metricsRollingStatisticalWindowInMilliseconds"`
	PropertyValueRequestCacheEnabled                              bool        `json:"propertyValue_requestCacheEnabled"`
	PropertyValueRequestLogEnabled                                bool        `json:"propertyValue_requestLogEnabled"`
	ReportingHosts                                                int         `json:"reportingHosts"`
	ThreadPool                                                    string      `json:"threadPool"`
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

func entryStream(reader io.ReadCloser, maxEntries int) (chan HystrixStreamEntry, chan error) {
	entryChannel := make(chan HystrixStreamEntry)
	stopChannel := make(chan error)

	scanner := bufio.NewScanner(reader)
	entryCounter := 0

	go func() {
	forever:
		for {
			chunks := streamToStrings(scanner)
			for _, chunk := range chunks {
				if entries, err := parseChunk(chunk); err == nil {
					for _, entry := range entries {
						entryChannel <- entry
						entryCounter++
						if maxEntries > 0 && entryCounter >= maxEntries {
							stopChannel <- io.EOF
							break forever
						}
					}
				} else {
					stopChannel <- err
					break forever
				}
			}

		}
		reader.Close()
	}()

	return entryChannel, stopChannel
}

func streamToStrings(scanner *bufio.Scanner) []string {
	result := make([]string, 0)
	for scanner.Scan() {
		text := scanner.Text()
		if isData(text) {
			result = append(result, scanner.Text())
			break
		}
	}
	return result
}
func isData(i string) bool {
	return len(i) > 0
}
