package aliyuncms

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
)

const (
	description  = "Pull Metric Statistics from Aliyun CMS"
	sampleConfig = `
  ## Aliyun Credentials
  ## Credentials are loaded in the following order
  ## 1) Ram RoleArn credential
  ## 2) AccessKey STS token credential
  ## 3) AccessKey credential
  ## 4) Ecs Ram Role credential
  ## 5) RSA keypair credential
  ## 6) Environment variables credential
  ## 7) Instance metadata credential
  
  # access_key_id = ""
  # access_key_secret = ""
  # access_key_sts_token = ""
  # role_arn = ""
  # role_session_name = ""
  # private_key = ""
  # public_key_id = ""
  # role_name = ""

  ## Specify the ali cloud region list to be queried for metrics and objects discovery
  ## If not set, all supported regions (see below) would be covered, it can provide a significant load on API, so the recommendation here 
  ## is to limit the list as much as possible. Allowed values: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
  ## Default supported regions are:
  ## 21 items: cn-qingdao,cn-beijing,cn-zhangjiakou,cn-huhehaote,cn-hangzhou,cn-shanghai,cn-shenzhen,
  ##           cn-heyuan,cn-chengdu,cn-hongkong,ap-southeast-1,ap-southeast-2,ap-southeast-3,ap-southeast-5,
  ##           ap-south-1,ap-northeast-1,us-west-1,us-east-1,eu-central-1,eu-west-1,me-east-1
  ##
  ## From discovery perspective it set the scope for object discovery, the discovered info can be used to enrich
  ## the metrics with objects attributes/tags. Discovery is supported not for all projects (if not supported, then 
  ## it will be reported on the start - for example for 'acs_cdn' project:
  ## 'E! [inputs.aliyuncms] Discovery tool is not activated: no discovery support for project "acs_cdn"' )
  ## Currently, discovery supported for the following projects:
  ## - acs_ecs_dashboard
  ## - acs_rds_dashboard
  ## - acs_slb_dashboard
  ## - acs_vpc_eip   
  regions = ["cn-hongkong"]

  # The minimum period for AliyunCMS metrics is 1 minute (60s). However not all
  # metrics are made available to the 1 minute period. Some are collected at
  # 3 minute, 5 minute, or larger intervals.
  # See: https://help.aliyun.com/document_detail/51936.html?spm=a2c4g.11186623.2.18.2bc1750eeOw1Pv
  # Note that if a period is configured that is smaller than the minimum for a
  # particular metric, that metric will not be returned by the Aliyun OpenAPI
  # and will not be collected by Telegraf.
  #
  ## Requested AliyunCMS aggregation Period (required - must be a multiple of 60s)
  period = "5m"
  
  ## Collection Delay (required - must account for metrics availability via AliyunCMS API)
  delay = "1m"
  
  ## Recommended: use metric 'interval' that is a multiple of 'period' to avoid
  ## gaps or overlap in pulled data
  interval = "5m"
  
  ## Metric Statistic Project (required)
  project = "acs_slb_dashboard"
  
  ## Maximum requests per second, default value is 200
  ratelimit = 200
  
  ## How often the discovery API call executed (default 1m)
  #discovery_interval = "1m"
  
  ## Metrics to Pull (Required)
  [[inputs.aliyuncms.metrics]]
  ## Metrics names to be requested, 
  ## described here (per project): https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.6.690.1938ad41wg8QSq
  names = ["InstanceActiveConnection", "InstanceNewConnection"]
  
  ## Dimension filters for Metric (these are optional).
  ## This allows to get additional metric dimension. If dimension is not specified it can be returned or
  ## the data can be aggregated - it depends on particular metric, you can find details here: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.6.690.1938ad41wg8QSq
  ##
  ## Note, that by default dimension filter includes the list of discovered objects in scope (if discovery is enabled)
  ## Values specified here would be added into the list of discovered objects.
  ## You can specify either single dimension:      
  #dimensions = '{"instanceId": "p-example"}'
  
  ## Or you can specify several dimensions at once:
  #dimensions = '[{"instanceId": "p-example"},{"instanceId": "q-example"}]'
  
  ## Enrichment tags, can be added from discovery (if supported)
  ## Notation is <measurement_tag_name>:<JMES query path (https://jmespath.org/tutorial.html)>
  ## To figure out which fields are available, consult the Describe<ObjectType> API per project.
  ## For example, for SLB: https://api.aliyun.com/#/?product=Slb&version=2014-05-15&api=DescribeLoadBalancers&params={}&tab=MOCK&lang=GO
  #tag_query_path = [
  #    "address:Address",
  #    "name:LoadBalancerName",
  #    "cluster_owner:Tags.Tag[?TagKey=='cs.cluster.name'].TagValue | [0]"
  #    ]
  ## The following tags added by default: regionId (if discovery enabled), userId, instanceId.
  
  ## Allow metrics without discovery data, if discovery is enabled. If set to true, then metric without discovery
  ## data would be emitted, otherwise dropped. This cane be of help, in case debugging dimension filters, or partial coverage 
  ## of discovery scope vs monitoring scope 
  #allow_dps_without_discovery = false
`
)

type (
	// AliyunCMS is aliyun cms config info.
	AliyunCMS struct {
		AccessKeyID       string `toml:"access_key_id"`
		AccessKeySecret   string `toml:"access_key_secret"`
		AccessKeyStsToken string `toml:"access_key_sts_token"`
		RoleArn           string `toml:"role_arn"`
		RoleSessionName   string `toml:"role_session_name"`
		PrivateKey        string `toml:"private_key"`
		PublicKeyID       string `toml:"public_key_id"`
		RoleName          string `toml:"role_name"`

		Regions           []string        `toml:"regions"`
		DiscoveryInterval config.Duration `toml:"discovery_interval"`
		Period            config.Duration `toml:"period"`
		Delay             config.Duration `toml:"delay"`
		Project           string          `toml:"project"`
		Metrics           []*Metric       `toml:"metrics"`
		RateLimit         int             `toml:"ratelimit"`

		Log telegraf.Logger `toml:"-"`

		client        aliyuncmsClient
		windowStart   time.Time
		windowEnd     time.Time
		dt            *discoveryTool
		dimensionKey  string
		discoveryData map[string]interface{}
		measurement   string
	}

	// Metric describes what metrics to get
	Metric struct {
		ObjectsFilter                 string   `toml:"objects_filter"`
		MetricNames                   []string `toml:"names"`
		Dimensions                    string   `toml:"dimensions"` //String representation of JSON dimensions
		TagsQueryPath                 []string `toml:"tag_query_path"`
		AllowDataPointWODiscoveryData bool     `toml:"allow_dps_without_discovery"` //Allow data points without discovery data (if no discovery data found)

		dtLock               sync.Mutex                   //Guard for discoveryTags & dimensions
		discoveryTags        map[string]map[string]string //Internal data structure that can enrich metrics with tags
		dimensionsUdObj      map[string]string
		dimensionsUdArr      []map[string]string //Parsed Dimesnsions JSON string (unmarshalled)
		requestDimensions    []map[string]string //this is the actual dimensions list that would be used in API request
		requestDimensionsStr string              //String representation of the above

	}

	// Dimension describe how to get metrics
	Dimension struct {
		Value string `toml:"value"`
	}

	aliyuncmsClient interface {
		DescribeMetricList(request *cms.DescribeMetricListRequest) (response *cms.DescribeMetricListResponse, err error)
	}
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

// SampleConfig implements telegraf.Inputs interface
func (s *AliyunCMS) SampleConfig() string {
	return sampleConfig
}

// Description implements telegraf.Inputs interface
func (s *AliyunCMS) Description() string {
	return description
}

// Init perform checks of plugin inputs and initialize internals
func (s *AliyunCMS) Init() error {
	if s.Project == "" {
		return errors.New("project is not set")
	}

	var (
		roleSessionExpiration = 600
		sessionExpiration     = 600
	)
	configuration := &providers.Configuration{
		AccessKeyID:           s.AccessKeyID,
		AccessKeySecret:       s.AccessKeySecret,
		AccessKeyStsToken:     s.AccessKeyStsToken,
		RoleArn:               s.RoleArn,
		RoleSessionName:       s.RoleSessionName,
		RoleSessionExpiration: &roleSessionExpiration,
		PrivateKey:            s.PrivateKey,
		PublicKeyID:           s.PublicKeyID,
		SessionExpiration:     &sessionExpiration,
		RoleName:              s.RoleName,
	}
	credentialProviders := []providers.Provider{
		providers.NewConfigurationCredentialProvider(configuration),
		providers.NewEnvCredentialProvider(),
		providers.NewInstanceMetadataProvider(),
	}
	credential, err := providers.NewChainProvider(credentialProviders).Retrieve()
	if err != nil {
		return errors.Errorf("failed to retrieve credential: %v", err)
	}
	s.client, err = cms.NewClientWithOptions("", sdk.NewConfig(), credential)
	if err != nil {
		return errors.Errorf("failed to create cms client: %v", err)
	}

	//check metrics dimensions consistency
	for _, metric := range s.Metrics {
		if metric.Dimensions != "" {
			metric.dimensionsUdObj = map[string]string{}
			metric.dimensionsUdArr = []map[string]string{}
			err := json.Unmarshal([]byte(metric.Dimensions), &metric.dimensionsUdObj)
			if err != nil {
				err := json.Unmarshal([]byte(metric.Dimensions), &metric.dimensionsUdArr)
				return errors.Errorf("Can't parse dimensions (it is neither obj, nor array) %q :%v", metric.Dimensions, err)
			}
		}
	}

	s.measurement = formatMeasurement(s.Project)

	//Check regions
	if len(s.Regions) == 0 {
		s.Regions = aliyunRegionList
		s.Log.Infof("'regions' is not set. Metrics will be queried across %d regions:\n%s",
			len(s.Regions), strings.Join(s.Regions, ","))
	}

	//Init discovery...
	if s.dt == nil { //Support for tests
		s.dt, err = newDiscoveryTool(s.Regions, s.Project, s.Log, credential, int(float32(s.RateLimit)*0.2), time.Duration(s.DiscoveryInterval))
		if err != nil {
			s.Log.Errorf("Discovery tool is not activated: %v", err)
			s.dt = nil
			return nil
		}
	}

	s.discoveryData, err = s.dt.getDiscoveryDataAcrossRegions(nil)
	if err != nil {
		s.Log.Errorf("Discovery tool is not activated: %v", err)
		s.dt = nil
		return nil
	}

	s.Log.Infof("%d object(s) discovered...", len(s.discoveryData))

	//Special setting for acs_oss project since the API differs
	if s.Project == "acs_oss" {
		s.dimensionKey = "BucketName"
	}

	return nil
}

// Start plugin discovery loop, metrics are gathered through Gather
func (s *AliyunCMS) Start(telegraf.Accumulator) error {
	//Start periodic discovery process
	if s.dt != nil {
		s.dt.start()
	}

	return nil
}

// Gather implements telegraf.Inputs interface
func (s *AliyunCMS) Gather(acc telegraf.Accumulator) error {
	s.updateWindow(time.Now())

	// limit concurrency or we can easily exhaust user connection limit
	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)
	defer lmtr.Stop()

	var wg sync.WaitGroup
	for _, metric := range s.Metrics {
		//Prepare internal structure with data from discovery
		s.prepareTagsAndDimensions(metric)
		wg.Add(len(metric.MetricNames))
		for _, metricName := range metric.MetricNames {
			<-lmtr.C
			go func(metricName string, metric *Metric) {
				defer wg.Done()
				acc.AddError(s.gatherMetric(acc, metricName, metric))
			}(metricName, metric)
		}
		wg.Wait()
	}

	return nil
}

// Stop - stops the plugin discovery loop
func (s *AliyunCMS) Stop() {
	if s.dt != nil {
		s.dt.stop()
	}
}

func (s *AliyunCMS) updateWindow(relativeTo time.Time) {
	//https://help.aliyun.com/document_detail/51936.html?spm=a2c4g.11186623.6.701.54025679zh6wiR
	//The start and end times are executed in the mode of
	//opening left and closing right, and startTime cannot be equal
	//to or greater than endTime.

	windowEnd := relativeTo.Add(-time.Duration(s.Delay))

	if s.windowEnd.IsZero() {
		// this is the first run, no window info, so just get a single period
		s.windowStart = windowEnd.Add(-time.Duration(s.Period))
	} else {
		// subsequent window, start where last window left off
		s.windowStart = s.windowEnd
	}

	s.windowEnd = windowEnd
}

// Gather given metric and emit error
func (s *AliyunCMS) gatherMetric(acc telegraf.Accumulator, metricName string, metric *Metric) error {
	for _, region := range s.Regions {
		req := cms.CreateDescribeMetricListRequest()
		req.Period = strconv.FormatInt(int64(time.Duration(s.Period).Seconds()), 10)
		req.MetricName = metricName
		req.Length = "10000"
		req.Namespace = s.Project
		req.EndTime = strconv.FormatInt(s.windowEnd.Unix()*1000, 10)
		req.StartTime = strconv.FormatInt(s.windowStart.Unix()*1000, 10)
		req.Dimensions = metric.requestDimensionsStr
		req.RegionId = region

		for more := true; more; {
			resp, err := s.client.DescribeMetricList(req)
			if err != nil {
				return errors.Errorf("failed to query metricName list: %v", err)
			}
			if resp.Code != "200" {
				s.Log.Errorf("failed to query metricName list: %v", resp.Message)
				break
			}

			var datapoints []map[string]interface{}
			if err := json.Unmarshal([]byte(resp.Datapoints), &datapoints); err != nil {
				return errors.Errorf("failed to decode response datapoints: %v", err)
			}

			if len(datapoints) == 0 {
				s.Log.Debugf("No metrics returned from CMS, response msg: %s", resp.Message)
				break
			}

		NextDataPoint:
			for _, datapoint := range datapoints {
				fields := map[string]interface{}{}
				datapointTime := int64(0)
				tags := map[string]string{}
				for key, value := range datapoint {
					switch key {
					case "instanceId", "BucketName":
						tags[key] = value.(string)
						if metric.discoveryTags != nil { //discovery can be not activated
							//Skipping data point if discovery data not exist
							_, ok := metric.discoveryTags[value.(string)]
							if !ok &&
								!metric.AllowDataPointWODiscoveryData {
								s.Log.Warnf("Instance %q is not found in discovery, skipping monitoring datapoint...", value.(string))
								continue NextDataPoint
							}

							for k, v := range metric.discoveryTags[value.(string)] {
								tags[k] = v
							}
						}
					case "userId":
						tags[key] = value.(string)
					case "timestamp":
						datapointTime = int64(value.(float64)) / 1000
					default:
						fields[formatField(metricName, key)] = value
					}
				}
				//Log.logW("Datapoint time: %s, now: %s", time.Unix(datapointTime, 0).Format(time.RFC3339), time.Now().Format(time.RFC3339))
				acc.AddFields(s.measurement, fields, tags, time.Unix(datapointTime, 0))
			}

			req.NextToken = resp.NextToken
			more = req.NextToken != ""
		}
	}
	return nil
}

//tag helper
func parseTag(tagSpec string, data interface{}) (tagKey string, tagValue string, err error) {
	var (
		ok        bool
		queryPath = tagSpec
	)
	tagKey = tagSpec

	//Split query path to tagKey and query path
	if splitted := strings.Split(tagSpec, ":"); len(splitted) == 2 {
		tagKey = splitted[0]
		queryPath = splitted[1]
	}

	tagRawValue, err := jmespath.Search(queryPath, data)
	if err != nil {
		return "", "", errors.Errorf("Can't query data from discovery data using query path %q: %v",
			queryPath, err)
	}

	if tagRawValue == nil { //Nothing found
		return "", "", nil
	}

	tagValue, ok = tagRawValue.(string)
	if !ok {
		return "", "", errors.Errorf("Tag value %v parsed by query %q is not a string value",
			tagRawValue, queryPath)
	}

	return tagKey, tagValue, nil
}

func (s *AliyunCMS) prepareTagsAndDimensions(metric *Metric) {
	var (
		newData    bool
		defaulTags = []string{"RegionId:RegionId"}
	)

	if s.dt == nil { //Discovery is not activated
		return
	}

	//Reading all data from buffered channel
L:
	for {
		select {
		case s.discoveryData = <-s.dt.dataChan:
			newData = true
			continue
		default:
			break L
		}
	}

	//new data arrives (so process it) or this is the first call
	if newData || len(metric.discoveryTags) == 0 {
		metric.dtLock.Lock()
		defer metric.dtLock.Unlock()

		if metric.discoveryTags == nil {
			metric.discoveryTags = make(map[string]map[string]string, len(s.discoveryData))
		}

		metric.requestDimensions = nil //erasing
		metric.requestDimensions = make([]map[string]string, 0, len(s.discoveryData))

		//Preparing tags & dims...
		for instanceID, elem := range s.discoveryData {
			//Start filing tags
			//Remove old value if exist
			delete(metric.discoveryTags, instanceID)
			metric.discoveryTags[instanceID] = make(map[string]string, len(metric.TagsQueryPath)+len(defaulTags))

			for _, tagQueryPath := range metric.TagsQueryPath {
				tagKey, tagValue, err := parseTag(tagQueryPath, elem)
				if err != nil {
					s.Log.Errorf("%v", err)
					continue
				}
				if err == nil && tagValue == "" { //Nothing found
					s.Log.Debugf("Data by query path %q: is not found, for instance %q", tagQueryPath, instanceID)
					continue
				}

				metric.discoveryTags[instanceID][tagKey] = tagValue
			}

			//Adding default tags if not already there
			for _, defaultTagQP := range defaulTags {
				tagKey, tagValue, err := parseTag(defaultTagQP, elem)

				if err != nil {
					s.Log.Errorf("%v", err)
					continue
				}

				if err == nil && tagValue == "" { //Nothing found
					s.Log.Debugf("Data by query path %q: is not found, for instance %q",
						defaultTagQP, instanceID)
					continue
				}

				metric.discoveryTags[instanceID][tagKey] = tagValue
			}

			//Preparing dimensions (first adding dimensions that comes from discovery data)
			metric.requestDimensions = append(
				metric.requestDimensions,
				map[string]string{s.dimensionKey: instanceID})
		}

		//Get final dimension (need to get full lis of
		//what was provided in config + what comes from discovery
		if len(metric.dimensionsUdArr) != 0 {
			metric.requestDimensions = append(metric.requestDimensions, metric.dimensionsUdArr...)
		}
		if len(metric.dimensionsUdObj) != 0 {
			metric.requestDimensions = append(metric.requestDimensions, metric.dimensionsUdObj)
		}

		//Unmarshalling to string
		reqDim, err := json.Marshal(metric.requestDimensions)
		if err != nil {
			s.Log.Errorf("Can't marshal metric request dimensions %v :%v",
				metric.requestDimensions, err)
			metric.requestDimensionsStr = ""
		} else {
			metric.requestDimensionsStr = string(reqDim)
		}
	}
}

// Formatting helpers
func formatField(metricName string, statistic string) string {
	if metricName == statistic {
		statistic = "value"
	}
	return fmt.Sprintf("%s_%s", snakeCase(metricName), snakeCase(statistic))
}

func formatMeasurement(project string) string {
	project = strings.Replace(project, "/", "_", -1)
	project = snakeCase(project)
	return fmt.Sprintf("aliyuncms_%s", project)
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.Replace(s, "__", "_", -1)
	return s
}

func init() {
	inputs.Add("aliyuncms", func() telegraf.Input {
		return &AliyunCMS{
			RateLimit:         200,
			DiscoveryInterval: config.Duration(time.Minute),
			dimensionKey:      "instanceId",
		}
	})
}
