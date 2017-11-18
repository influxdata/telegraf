package burrow

import (
	"encoding/json"
	"errors"
	"net/http"

	"fmt"
	"github.com/influxdata/telegraf"
)

type (
	// burrow api client
	apiClient struct {
		client      http.Client
		acc         telegraf.Accumulator
		apiPrefix   string
		baseURL     string
		workerCount int

		limitClusters []string
		limitGroups   []string
		limitTopics   []string

		requestUser string
		requestPass string

		guardChan chan bool
		errorChan chan<- error
	}

	// burrow api response
	apiResponse struct {
		Error   bool       `json:"error"`
		Request apiRequest `json:"request"`
		Message string     `json:"message"`

		// all possible possible answers
		Clusters []string          `json:"clusters"`
		Groups   []string          `json:"consumers"`
		Topics   []string          `json:"topics"`
		Offsets  []int64           `json:"offsets"`
		Status   apiStatusResponse `json:"status"`
	}

	// burrow api response: request field
	apiRequest struct {
		URI     string `json:"uri"`
		Host    string `json:"host"`
		Cluster string `json:"cluster"`
		Group   string `json:"group"`
		Topic   string `json:"topic"`
	}

	// burrow api response: status field
	apiStatusResponse struct {
		Partitions []apiStatusResponseLag `json:"partitions"`
	}

	// buttor api response: lag field
	apiStatusResponseLag struct {
		Topic     string                   `json:"topic"`
		Partition int32                    `json:"partition"`
		Status    string                   `json:"status"`
		Start     apiStatusResponseLagItem `json:"start"`
		End       apiStatusResponseLagItem `json:"end"`
	}

	// buttor api response: lag field item
	apiStatusResponseLagItem struct {
		Offset    int64 `json:"offset"`
		Timestamp int64 `json:"timestamp"`
		Lag       int64 `json:"lag"`
	}
)

// construct endpoint request
func (api *apiClient) getRequest(uri string) (*http.Request, error) {
	// create new request
	endpoint := fmt.Sprintf("%s%s", api.baseURL, uri)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// add support for http basic authorization
	if api.requestUser != "" {
		req.SetBasicAuth(api.requestUser, api.requestPass)
	}

	return req, nil
}

// perform synchronous http request
func (api *apiClient) call(uri string) (apiResponse, error) {
	var br apiResponse

	// acquire concurrent lock
	api.guardChan <- true
	defer func() {
		<-api.guardChan
	}()

	// get request
	req, err := api.getRequest(uri)
	if err != nil {
		return br, err
	}

	// do request
	res, err := api.client.Do(req)
	if err != nil {
		return br, err
	}

	// decode response
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return br, fmt.Errorf("endpoint: '%s', invalid response code: '%d'", uri, res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&br); err != nil {
		return br, err
	}

	// if error is raised, respond with error
	if br.Error {
		return br, errors.New(br.Message)
	}

	return br, err
}
