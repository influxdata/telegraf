package aliyun

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/limiter"
)

// DiscoveryRequest is a marker interface for Aliyun discovery requests
type DiscoveryRequest interface{}

// AliyunSdkClient is an interface for Aliyun SDK clients
type AliyunSdkClient interface {
	ProcessCommonRequest(req *requests.CommonRequest) (response *responses.CommonResponse, err error)
}

// DiscoveryTool provides automatic discovery of Aliyun resources
type DiscoveryTool struct {
	Req                map[string]DiscoveryRequest // Discovery request per region
	RateLimit          int                         // Rate limit for API queries
	ReqDefaultPageSize int                         // Default page size for pagination
	Cli                map[string]AliyunSdkClient  // API client per region

	RespRootKey     string // Root key in JSON response for discovery data
	RespObjectIDKey string // Key for object ID in discovery data

	wg       sync.WaitGroup              // WaitGroup for discovery goroutine
	Interval time.Duration               // Discovery interval
	done     chan bool                   // Channel to stop discovery
	DataChan chan map[string]interface{} // Channel for discovery data
	Lg       telegraf.Logger             // Logger
}

// ParsedDiscoveryResponse contains parsed discovery response data
type ParsedDiscoveryResponse struct {
	Data       []interface{}
	TotalCount int
	PageSize   int
	PageNumber int
}

// getRPCReqFromDiscoveryRequest extracts RpcRequest from a discovery request
func getRPCReqFromDiscoveryRequest(req DiscoveryRequest) (*requests.RpcRequest, error) {
	if reflect.ValueOf(req).Type().Kind() != reflect.Ptr ||
		reflect.ValueOf(req).IsNil() {
		return nil, fmt.Errorf("unexpected type of the discovery request object: %q, %q",
			reflect.ValueOf(req).Type(), reflect.ValueOf(req).Kind())
	}

	ptrV := reflect.Indirect(reflect.ValueOf(req))

	for i := 0; i < ptrV.NumField(); i++ {
		if ptrV.Field(i).Type().String() == "*requests.RpcRequest" {
			if !ptrV.Field(i).CanInterface() {
				return nil, fmt.Errorf("can't get interface of %q", ptrV.Field(i))
			}

			rpcReq, ok := ptrV.Field(i).Interface().(*requests.RpcRequest)
			if !ok {
				return nil, fmt.Errorf("can't convert interface of %q to '*requests.RpcRequest' type",
					ptrV.Field(i).Interface())
			}

			return rpcReq, nil
		}
	}
	return nil, fmt.Errorf("didn't find *requests.RpcRequest embedded struct in %q", ptrV.Type())
}

// parseDiscoveryResponse parses the discovery API response
func (dt *DiscoveryTool) parseDiscoveryResponse(resp *responses.CommonResponse) (*ParsedDiscoveryResponse, error) {
	var (
		fullOutput    = make(map[string]interface{})
		data          []byte
		foundDataItem bool
		foundRootKey  bool
		pdResp        = &ParsedDiscoveryResponse{}
	)

	data = resp.GetHttpContentBytes()
	if data == nil {
		return nil, errors.New("no data in response to be parsed")
	}

	if err := json.Unmarshal(data, &fullOutput); err != nil {
		return nil, fmt.Errorf("can't parse JSON from discovery response: %w", err)
	}

	for key, val := range fullOutput {
		switch key {
		case dt.RespRootKey:
			foundRootKey = true
			rootKeyVal, ok := val.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("content of root key %q, is not an object: %q", key, val)
			}

			// Find the array with discovered data
			for _, item := range rootKeyVal {
				if pdResp.Data, foundDataItem = item.([]interface{}); foundDataItem {
					break
				}
			}
			if !foundDataItem {
				return nil, fmt.Errorf("didn't find array item in root key %q", key)
			}
		case "TotalCount", "TotalRecordCount":
			pdResp.TotalCount = int(val.(float64))
		case "PageSize", "PageRecordCount":
			pdResp.PageSize = int(val.(float64))
		case "PageNumber":
			pdResp.PageNumber = int(val.(float64))
		}
	}
	if !foundRootKey {
		return nil, fmt.Errorf("didn't find root key %q in discovery response", dt.RespRootKey)
	}

	return pdResp, nil
}

// getDiscoveryData retrieves discovery data from a single region with pagination
func (dt *DiscoveryTool) getDiscoveryData(cli AliyunSdkClient, req *requests.CommonRequest, lmtr chan bool) (map[string]interface{}, error) {
	var (
		err           error
		resp          *responses.CommonResponse
		pDResp        *ParsedDiscoveryResponse
		discoveryData []interface{}
		totalCount    int
		pageNumber    int
	)
	defer delete(req.QueryParams, "PageNumber")

	for {
		if lmtr != nil {
			<-lmtr // Rate limiting
		}

		resp, err = cli.ProcessCommonRequest(req)
		if err != nil {
			return nil, err
		}

		pDResp, err = dt.parseDiscoveryResponse(resp)
		if err != nil {
			return nil, err
		}
		discoveryData = append(discoveryData, pDResp.Data...)
		pageNumber = pDResp.PageNumber
		totalCount = pDResp.TotalCount

		// Pagination
		pageNumber++
		req.QueryParams["PageNumber"] = strconv.Itoa(pageNumber)

		if len(discoveryData) == totalCount {
			// Map data to the appropriate shape before return
			preparedData := make(map[string]interface{}, len(discoveryData))

			for _, raw := range discoveryData {
				elem, ok := raw.(map[string]interface{})
				if !ok {
					return nil, errors.New("can't parse input data element, not a map[string]interface{} type")
				}
				if objectID, ok := elem[dt.RespObjectIDKey].(string); ok {
					preparedData[objectID] = elem
				}
			}
			return preparedData, nil
		}
	}
}

// GetDiscoveryDataAcrossRegions retrieves discovery data from all configured regions
func (dt *DiscoveryTool) GetDiscoveryDataAcrossRegions(lmtr chan bool) (map[string]interface{}, error) {
	resultData := make(map[string]interface{})

	for region, cli := range dt.Cli {
		dscReq, ok := dt.Req[region]
		if !ok {
			return nil, fmt.Errorf("error building common discovery request: not valid region %q", region)
		}

		rpcReq, err := getRPCReqFromDiscoveryRequest(dscReq)
		if err != nil {
			return nil, err
		}

		commonRequest := requests.NewCommonRequest()
		commonRequest.Method = rpcReq.GetMethod()
		commonRequest.Product = rpcReq.GetProduct()
		commonRequest.Domain = rpcReq.GetDomain()
		commonRequest.Version = rpcReq.GetVersion()
		commonRequest.Scheme = rpcReq.GetScheme()
		commonRequest.ApiName = rpcReq.GetActionName()
		commonRequest.QueryParams = rpcReq.QueryParams
		commonRequest.QueryParams["PageSize"] = strconv.Itoa(dt.ReqDefaultPageSize)
		commonRequest.TransToAcsRequest()

		// Get discovery data using a common request
		data, discoveryDataErr := dt.getDiscoveryData(cli, commonRequest, lmtr)
		if discoveryDataErr != nil {
			return nil, discoveryDataErr
		}

		for k, v := range data {
			resultData[k] = v
		}
	}
	return resultData, nil
}

// Start begins the discovery polling loop
func (dt *DiscoveryTool) Start() {
	var (
		err      error
		data     map[string]interface{}
		lastData map[string]interface{}
	)

	dt.done = make(chan bool)

	dt.wg.Add(1)
	go func() {
		defer dt.wg.Done()

		ticker := time.NewTicker(dt.Interval)
		defer ticker.Stop()

		lmtr := limiter.NewRateLimiter(dt.RateLimit, time.Second)
		defer lmtr.Stop()

		for {
			select {
			case <-dt.done:
				return
			case <-ticker.C:
				data, err = dt.GetDiscoveryDataAcrossRegions(lmtr.C)
				if err != nil {
					dt.Lg.Errorf("Can't get discovery data: %v", err)
					continue
				}

				if !reflect.DeepEqual(data, lastData) {
					lastData = make(map[string]interface{}, len(data))
					for k, v := range data {
						lastData[k] = v
					}

					// Send discovery data in blocking mode
					dt.DataChan <- data
				}
			}
		}
	}()
}

// Stop stops the discovery loop
func (dt *DiscoveryTool) Stop() {
	close(dt.done)

	// Shutdown timer
	timer := time.NewTimer(time.Second * 3)
	defer timer.Stop()
L:
	for {
		select {
		case <-timer.C:
			break L
		case <-dt.DataChan:
		}
	}

	dt.wg.Wait()
}
