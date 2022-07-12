package t128_metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

// Retriever implements a way to retrieve metrics
type Retriever interface {
	RequestCount() int
	Describe(index int) string
	CreateRequest(index int, baseURL string) (*http.Request, error)
	PopulateResponse(index int, acc telegraf.Accumulator, responseMetrics []ResponseMetric, timestamp time.Time)
}

// NewIndividualRetriever produces an individual retriever
func NewIndividualRetriever(useIntegerConversion bool, configuredMetrics []ConfiguredMetric) Retriever {
	requestMetrics := make([]RequestMetric, 0)

	for _, configMetric := range configuredMetrics {
		parameters := toRequestParameters(configMetric.Parameters)

		for fieldName, fieldPath := range configMetric.Fields {
			requestMetrics = append(requestMetrics, RequestMetric{
				ID:             fieldPath,
				Parameters:     parameters,
				OutMeasurement: configMetric.Name,
				OutField:       fieldName,
			})
		}
	}

	return &individualRetriever{
		metrics:              requestMetrics,
		useIntegerConversion: useIntegerConversion,
	}
}

// NewBulkRetriever creates a new bulk retriever
func NewBulkRetriever(useIntegerConversion bool, configuredMetrics []ConfiguredMetric) (Retriever, error) {
	requestMetrics := make([]BulkRequestMetrics, 0)

	for _, configMetric := range configuredMetrics {
		bulkRequest := BulkRequestMetrics{
			IDs:            make([]string, 0, len(configMetric.Fields)),
			Parameters:     toRequestParameters(configMetric.Parameters),
			OutMeasurement: configMetric.Name,
			OutFields:      make(map[string]string, len(configMetric.Fields)),
		}

		for fieldName, fieldPath := range configMetric.Fields {
			requestID := "/" + fieldPath
			bulkRequest.IDs = append(bulkRequest.IDs, requestID)
			bulkRequest.OutFields[requestID] = fieldName
		}

		sort.Strings(bulkRequest.IDs)

		var err error
		bulkRequest.RequestBody, err = json.Marshal(bulkRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to create request body for measurment '%s': %w", configMetric.Name, err)
		}

		requestMetrics = append(requestMetrics, bulkRequest)
	}

	return &bulkRetriever{
		metrics:              requestMetrics,
		useIntegerConversion: useIntegerConversion,
	}, nil
}

type individualRetriever struct {
	metrics              []RequestMetric
	useIntegerConversion bool
}

func (r *individualRetriever) RequestCount() int {
	return len(r.metrics)
}

func (r *individualRetriever) Describe(index int) string {
	return fmt.Sprintf("metric %s", r.metrics[index].ID)
}

func (r *individualRetriever) CreateRequest(index int, baseURL string) (*http.Request, error) {
	metric := r.metrics[index]

	content := struct {
		Parameters []RequestParameter `json:"parameters,omitempty"`
	}{
		metric.Parameters,
	}

	body, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body for metric '%s': %w", metric.ID, err)
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", baseURL, metric.ID), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for metric '%s': %w", metric.ID, err)
	}

	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func (r *individualRetriever) PopulateResponse(index int, acc telegraf.Accumulator, responseMetrics []ResponseMetric, timestamp time.Time) {
	metric := r.metrics[index]
	for _, responseMetric := range responseMetrics {
		for _, permutation := range responseMetric.Permutations {
			if permutation.Value == nil {
				continue
			}

			tags := make(map[string]string)
			for _, parameter := range permutation.Parameters {
				tags[parameter.Name] = parameter.Value
			}

			acc.AddFields(
				metric.OutMeasurement,
				map[string]interface{}{metric.OutField: tryNumericConversion(
					r.useIntegerConversion,
					*permutation.Value),
				},
				tags,
				timestamp,
			)
		}
	}
}

type bulkRetriever struct {
	metrics              []BulkRequestMetrics
	useIntegerConversion bool
}

func (r *bulkRetriever) RequestCount() int {
	return len(r.metrics)
}

func (r *bulkRetriever) Describe(index int) string {
	return fmt.Sprintf("measurement %s", r.metrics[index].OutMeasurement)
}

func (r *bulkRetriever) CreateRequest(index int, baseURL string) (*http.Request, error) {
	metrics := r.metrics[index]
	request, err := http.NewRequest("POST", baseURL, bytes.NewReader(metrics.RequestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for measurement '%s': %w", metrics.OutMeasurement, err)
	}

	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func (r *bulkRetriever) PopulateResponse(index int, acc telegraf.Accumulator, responseMetrics []ResponseMetric, timestamp time.Time) {
	metrics := r.metrics[index]
	for _, responseMetric := range responseMetrics {
		outField, ok := metrics.OutFields[responseMetric.ID]
		if !ok {
			acc.AddError(fmt.Errorf("response contained unexpected metric: %s", responseMetric.ID))
			continue
		}

		for _, permutation := range responseMetric.Permutations {
			if permutation.Value == nil {
				continue
			}

			tags := make(map[string]string)
			for _, parameter := range permutation.Parameters {
				tags[parameter.Name] = parameter.Value
			}

			acc.AddFields(
				metrics.OutMeasurement,
				map[string]interface{}{outField: tryNumericConversion(
					r.useIntegerConversion,
					*permutation.Value),
				},
				tags,
				timestamp,
			)
		}
	}
}

func toRequestParameters(configParameters map[string][]string) []RequestParameter {
	// Sort names for consistency in testing. It's not free, but only happens during startup.
	parameterNames := make([]string, 0, len(configParameters))
	for parameterName := range configParameters {
		parameterNames = append(parameterNames, parameterName)
	}
	sort.Strings(parameterNames)

	parameters := make([]RequestParameter, 0, len(configParameters))
	for _, parameterName := range parameterNames {
		values := configParameters[parameterName]
		parameters = append(parameters, RequestParameter{
			Name:    parameterName,
			Values:  values,
			Itemize: true,
		})
	}

	return parameters
}

func tryNumericConversion(useIntegerConversion bool, value string) interface{} {
	if useIntegerConversion {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}
