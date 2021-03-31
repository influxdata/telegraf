package aliyuncms

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/pkg/errors"
)

// https://www.alibabacloud.com/help/doc-detail/40654.htm?gclid=Cj0KCQjw4dr0BRCxARIsAKUNjWTAMfyVUn_Y3OevFBV3CMaazrhq0URHsgE7c0m0SeMQRKlhlsJGgIEaAviyEALw_wcB
var aliyunRegionList = []string{
	"cn-qingdao",
	"cn-beijing",
	"cn-zhangjiakou",
	"cn-huhehaote",
	"cn-hangzhou",
	"cn-shanghai",
	"cn-shenzhen",
	"cn-heyuan",
	"cn-chengdu",
	"cn-hongkong",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-5",
	"ap-south-1",
	"ap-northeast-1",
	"us-west-1",
	"us-east-1",
	"eu-central-1",
	"eu-west-1",
	"me-east-1",
}

type discoveryRequest interface {
}

type aliyunSdkClient interface {
	ProcessCommonRequest(req *requests.CommonRequest) (response *responses.CommonResponse, err error)
}

type discoveryTool struct {
	req                map[string]discoveryRequest //Discovery request (specific per object type)
	rateLimit          int                         //Rate limit for API query, as it is limited by API backend
	reqDefaultPageSize int                         //Default page size while querying data from API (how many objects per request)
	cli                map[string]aliyunSdkClient  //API client, which perform discovery request

	respRootKey     string //Root key in JSON response where to look for discovery data
	respObjectIDKey string //Key in element of array under root key, that stores object ID
	//for ,majority of cases it would be InstanceId, for OSS it is BucketName. This key is also used in dimension filtering// )
	wg       sync.WaitGroup              //WG for primary discovery goroutine
	interval time.Duration               //Discovery interval
	done     chan bool                   //Done channel to stop primary discovery goroutine
	dataChan chan map[string]interface{} //Discovery data
	lg       telegraf.Logger             //Telegraf logger (should be provided)
}

//getRPCReqFromDiscoveryRequest - utility function to map between aliyun request primitives
//discoveryRequest represents different type of discovery requests
func getRPCReqFromDiscoveryRequest(req discoveryRequest) (*requests.RpcRequest, error) {
	if reflect.ValueOf(req).Type().Kind() != reflect.Ptr ||
		reflect.ValueOf(req).IsNil() {
		return nil, errors.Errorf("Not expected type of the discovery request object: %q, %q", reflect.ValueOf(req).Type(), reflect.ValueOf(req).Kind())
	}

	ptrV := reflect.Indirect(reflect.ValueOf(req))

	for i := 0; i < ptrV.NumField(); i++ {
		if ptrV.Field(i).Type().String() == "*requests.RpcRequest" {
			if !ptrV.Field(i).CanInterface() {
				return nil, errors.Errorf("Can't get interface of %v", ptrV.Field(i))
			}

			rpcReq, ok := ptrV.Field(i).Interface().(*requests.RpcRequest)

			if !ok {
				return nil, errors.Errorf("Cant convert interface of %v to '*requests.RpcRequest' type", ptrV.Field(i).Interface())
			}

			return rpcReq, nil
		}
	}
	return nil, errors.Errorf("Didn't find *requests.RpcRequest embedded struct in %q", ptrV.Type())
}

//NewDiscoveryTool function returns discovery tool object.
//The object is used to periodically get data about aliyun objects and send this
//data into channel. The intention is to enrich reported metrics with discovery data.
//Discovery is supported for a limited set of object types (defined by project) and can be extended in future.
//Discovery can be limited by region if not set, then all regions is queried.
//Request against API can inquire additional costs, consult with aliyun API documentation.
func NewDiscoveryTool(regions []string, project string, lg telegraf.Logger, credential auth.Credential, rateLimit int, discoveryInterval time.Duration) (*discoveryTool, error) {
	var (
		dscReq                = map[string]discoveryRequest{}
		cli                   = map[string]aliyunSdkClient{}
		parseRootKey          = regexp.MustCompile(`Describe(.*)`)
		responseRootKey       string
		responseObjectIDKey   string
		err                   error
		noDiscoverySupportErr = errors.Errorf("no discovery support for project %q", project)
	)

	if len(regions) == 0 {
		regions = aliyunRegionList
		lg.Warnf("Discovery regions are not provided! Data will be queried across %d regions!", len(aliyunRegionList))
	}

	if rateLimit == 0 { //Can be a rounding case
		rateLimit = 1
	}

	for _, region := range regions {
		switch project {
		case "acs_ecs_dashboard":
			dscReq[region] = ecs.CreateDescribeInstancesRequest()
			responseObjectIDKey = "InstanceId"
		case "acs_rds_dashboard":
			dscReq[region] = rds.CreateDescribeDBInstancesRequest()
			responseObjectIDKey = "DBInstanceId"
		case "acs_slb_dashboard":
			dscReq[region] = slb.CreateDescribeLoadBalancersRequest()
			responseObjectIDKey = "LoadBalancerId"
		case "acs_memcache":
			return nil, noDiscoverySupportErr
		case "acs_ocs":
			return nil, noDiscoverySupportErr
		case "acs_oss":
			//oss is really complicated
			//it is on it's own format
			return nil, noDiscoverySupportErr

			//As a possible solution we can
			//mimic to request format supported by oss

			//req := DescribeLOSSRequest{
			//	RpcRequest: &requests.RpcRequest{},
			//}
			//req.InitWithApiInfo("oss", "2014-08-15", "DescribeDBInstances", "oss", "openAPI")
		case "acs_vpc_eip":
			dscReq[region] = vpc.CreateDescribeEipAddressesRequest()
			responseObjectIDKey = "AllocationId"
		case "acs_kvstore":
			return nil, noDiscoverySupportErr
		case "acs_mns_new":
			return nil, noDiscoverySupportErr
		case "acs_cdn":
			//API replies are in its own format.
			return nil, noDiscoverySupportErr
		case "acs_polardb":
			return nil, noDiscoverySupportErr
		case "acs_gdb":
			return nil, noDiscoverySupportErr
		case "acs_ads":
			return nil, noDiscoverySupportErr
		case "acs_mongodb":
			return nil, noDiscoverySupportErr
		case "acs_express_connect":
			return nil, noDiscoverySupportErr
		case "acs_fc":
			return nil, noDiscoverySupportErr
		case "acs_nat_gateway":
			return nil, noDiscoverySupportErr
		case "acs_sls_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_containerservice_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_vpn":
			return nil, noDiscoverySupportErr
		case "acs_bandwidth_package":
			return nil, noDiscoverySupportErr
		case "acs_cen":
			return nil, noDiscoverySupportErr
		case "acs_ens":
			return nil, noDiscoverySupportErr
		case "acs_opensearch":
			return nil, noDiscoverySupportErr
		case "acs_scdn":
			return nil, noDiscoverySupportErr
		case "acs_drds":
			return nil, noDiscoverySupportErr
		case "acs_iot":
			return nil, noDiscoverySupportErr
		case "acs_directmail":
			return nil, noDiscoverySupportErr
		case "acs_elasticsearch":
			return nil, noDiscoverySupportErr
		case "acs_ess_dashboard":
			return nil, noDiscoverySupportErr
		case "acs_streamcompute":
			return nil, noDiscoverySupportErr
		case "acs_global_acceleration":
			return nil, noDiscoverySupportErr
		case "acs_hitsdb":
			return nil, noDiscoverySupportErr
		case "acs_kafka":
			return nil, noDiscoverySupportErr
		case "acs_openad":
			return nil, noDiscoverySupportErr
		case "acs_pcdn":
			return nil, noDiscoverySupportErr
		case "acs_dcdn":
			return nil, noDiscoverySupportErr
		case "acs_petadata":
			return nil, noDiscoverySupportErr
		case "acs_videolive":
			return nil, noDiscoverySupportErr
		case "acs_hybriddb":
			return nil, noDiscoverySupportErr
		case "acs_adb":
			return nil, noDiscoverySupportErr
		case "acs_mps":
			return nil, noDiscoverySupportErr
		case "acs_maxcompute_prepay":
			return nil, noDiscoverySupportErr
		case "acs_hdfs":
			return nil, noDiscoverySupportErr
		case "acs_ddh":
			return nil, noDiscoverySupportErr
		case "acs_hbr":
			return nil, noDiscoverySupportErr
		case "acs_hdr":
			return nil, noDiscoverySupportErr
		case "acs_cds":
			return nil, noDiscoverySupportErr
		default:
			return nil, errors.Errorf("project %q is not recognized by discovery...", project)
		}

		cli[region], err = sdk.NewClientWithOptions(region, sdk.NewConfig(), credential)
		if err != nil {
			return nil, err
		}
	}

	if len(dscReq) == 0 || len(cli) == 0 {
		return nil, errors.Errorf("Can't build discovery request for project: %q,\nregions: %v", project, regions)
	}

	//Getting response root key (if not set already). This is to be able to parse discovery responses
	//As they differ per object type
	//Discovery requests are of the same type per every region, so pick the first one
	rpcReq, err := getRPCReqFromDiscoveryRequest(dscReq[regions[0]])
	//This means that the discovery request is not of proper type/kind
	if err != nil {
		return nil, errors.Errorf("Can't parse rpc request object from  discovery request %v", dscReq[regions[0]])
	}

	/*
		The action name is of the following format Describe<Project related title for managed instances>,
		For example: DescribeLoadBalancers -> for SLB project, or DescribeInstances for ECS project
		We will use it to construct root key name in the discovery API response.
		It follows the following logic: for 'DescribeLoadBalancers' action in discovery request we get the response
		in json of the following structure:
		{
			 ...
			 "LoadBalancers": {
				 "LoadBalancer": [ here comes objects, one per every instance]
			}
		}
		As we can see, the root key is a part of action name, except first word (part) 'Describe'
	*/
	result := parseRootKey.FindStringSubmatch(rpcReq.GetActionName())
	if result == nil || len(result) != 2 {
		return nil, errors.Errorf("Can't parse the discovery response root key from request action name %q", rpcReq.GetActionName())
	}
	responseRootKey = result[1]

	return &discoveryTool{
		req:                dscReq,
		cli:                cli,
		respRootKey:        responseRootKey,
		respObjectIDKey:    responseObjectIDKey,
		rateLimit:          rateLimit,
		interval:           discoveryInterval,
		reqDefaultPageSize: 20,
		dataChan:           make(chan map[string]interface{}, 1),
		lg:                 lg,
	}, nil
}

func (dt *discoveryTool) parseDiscoveryResponse(resp *responses.CommonResponse) (discData []interface{}, totalCount int, pageSize int, pageNumber int, err error) {
	var (
		fullOutput    = map[string]interface{}{}
		data          []byte
		foundDataItem bool
		foundRootKey  bool
	)

	data = resp.GetHttpContentBytes()
	if data == nil { //No data
		return nil, 0, 0, 0, errors.Errorf("No data in response to be parsed")
	}

	err = json.Unmarshal(data, &fullOutput)
	if err != nil {
		return nil, 0, 0, 0, errors.Errorf("Can't parse JSON from discovery response: %v", err)
	}

	for key, val := range fullOutput {
		switch key {
		case dt.respRootKey:
			foundRootKey = true
			rootKeyVal, ok := val.(map[string]interface{})
			if !ok {
				return nil, 0, 0, 0, errors.Errorf("Content of root key %q, is not an object: %v", key, val)
			}

			//It should contain the array with discovered data
			for _, item := range rootKeyVal {
				if discData, foundDataItem = item.([]interface{}); foundDataItem {
					break
				}
			}
			if !foundDataItem {
				return nil, 0, 0, 0, errors.Errorf("Didn't find array item in root key %q", key)
			}
		case "TotalCount":
			totalCount = int(val.(float64))
		case "PageSize":
			pageSize = int(val.(float64))
		case "PageNumber":
			pageNumber = int(val.(float64))
		}
	}
	if !foundRootKey {
		return nil, 0, 0, 0, errors.Errorf("Didn't find root key %q in discovery response", dt.respRootKey)
	}

	return
}

func (dt *discoveryTool) getDiscoveryData(cli aliyunSdkClient, req *requests.CommonRequest, limiter chan bool) (map[string]interface{}, error) {
	var (
		err           error
		resp          *responses.CommonResponse
		data          []interface{}
		discoveryData []interface{}
		totalCount    int
		pageNumber    int
	)
	defer delete(req.QueryParams, "PageNumber")

	for {
		if limiter != nil {
			<-limiter //Rate limiting
		}

		resp, err = cli.ProcessCommonRequest(req)
		if err != nil {
			return nil, err
		}

		data, totalCount, _, pageNumber, err = dt.parseDiscoveryResponse(resp)
		if err != nil {
			return nil, err
		}
		discoveryData = append(discoveryData, data...)

		//Pagination
		pageNumber++
		req.QueryParams["PageNumber"] = strconv.Itoa(pageNumber)

		if len(discoveryData) == totalCount { //All data received
			//Map data to appropriate shape before return
			preparedData := map[string]interface{}{}

			for _, raw := range discoveryData {
				if elem, ok := raw.(map[string]interface{}); ok {
					if objectID, ok := elem[dt.respObjectIDKey].(string); ok {
						preparedData[objectID] = elem
					}
				} else {
					return nil, errors.Errorf("Can't parse input data element, not a map[string]interface{} type")
				}
			}

			return preparedData, nil
		}
	}
}

func (dt *discoveryTool) getDiscoveryDataAllRegions(limiter chan bool) (map[string]interface{}, error) {
	var (
		data       map[string]interface{}
		resultData = map[string]interface{}{}
	)

	for region, cli := range dt.cli {
		//Building common request, as the code below is the same no matter
		//which aliyun object type (project) is used
		dscReq, ok := dt.req[region]
		if !ok {
			return nil, errors.Errorf("Error building common discovery request: not valid region %q", region)
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
		commonRequest.QueryParams["PageSize"] = strconv.Itoa(dt.reqDefaultPageSize)
		commonRequest.TransToAcsRequest()

		//Get discovery data using common request
		data, err = dt.getDiscoveryData(cli, commonRequest, limiter)
		if err != nil {
			return nil, err
		}

		for k, v := range data {
			resultData[k] = v
		}
	}
	return resultData, nil
}

func (dt *discoveryTool) Start() {
	var (
		err      error
		data     map[string]interface{}
		lastData map[string]interface{}
	)

	//Initializing channel
	dt.done = make(chan bool)

	dt.wg.Add(1)
	go func() {
		defer dt.wg.Done()

		ticker := time.NewTicker(dt.interval)
		defer ticker.Stop()

		lmtr := limiter.NewRateLimiter(dt.rateLimit, time.Second)
		defer lmtr.Stop()

		for {
			select {
			case <-dt.done:
				return
			case <-ticker.C:
				data, err = dt.getDiscoveryDataAllRegions(lmtr.C)
				if err != nil {
					dt.lg.Errorf("Can't get discovery data: %v", err)
					continue
				}

				if !reflect.DeepEqual(data, lastData) {
					lastData = nil
					lastData = map[string]interface{}{}
					for k, v := range data {
						lastData[k] = v
					}

					//send discovery data in blocking mode
					dt.dataChan <- data
				}
			}
		}
	}()
}

func (dt *discoveryTool) Stop() {
	close(dt.done)

	//Shutdown timer
	timer := time.NewTimer(time.Second * 3)
	defer timer.Stop()
L:
	for { //Unblock go routine by reading from dt.dataChan
		select {
		case <-timer.C:
			break L
		case <-dt.dataChan:
		}
	}

	dt.wg.Wait()
}
