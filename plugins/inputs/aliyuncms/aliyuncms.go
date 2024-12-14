//go:generate ../../../tools/readme_config_includer/generator
package aliyuncms

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmespath/go-jmespath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type (
	// AliyunCMS is Aliyun cms config info.
	AliyunCMS struct {
		AccessKeyID       string `toml:"access_key_id"`
		AccessKeySecret   string `toml:"access_key_secret"`
		AccessKeyStsToken string `toml:"access_key_sts_token"`
		RoleArn           string `toml:"role_arn"`
		RoleSessionName   string `toml:"role_session_name"`
		PrivateKey        string `toml:"private_key"`
		PublicKeyID       string `toml:"public_key_id"`
		RoleName          string `toml:"role_name"`

		MetricServices    []string        `toml:"metric_services"`
		Regions           []string        `toml:"regions"`
		DiscoveryInterval config.Duration `toml:"discovery_interval"`
		Period            config.Duration `toml:"period"`
		Delay             config.Duration `toml:"delay"`
		Project           string          `toml:"project"`
		Metrics           []*metric       `toml:"metrics"`
		RateLimit         int             `toml:"ratelimit"`

		Log telegraf.Logger `toml:"-"`

		cmsClient aliyuncmsClient
		rdsClient aliyunrdsClient

		windowStart   time.Time
		windowEnd     time.Time
		dt            *discoveryTool
		dimensionKey  string
		discoveryData map[string]interface{}
		measurement   string
	}

	// Metric describes what metrics to get
	metric struct {
		ObjectsFilter                 string   `toml:"objects_filter"`
		MetricNames                   []string `toml:"names"`
		Dimensions                    string   `toml:"dimensions"` // String representation of JSON dimensions
		TagsQueryPath                 []string `toml:"tag_query_path"`
		AllowDataPointWODiscoveryData bool     `toml:"allow_dps_without_discovery"` // Allow data points without discovery data (if no discovery data found)

		dtLock               sync.Mutex                   // Guard for discoveryTags & dimensions
		discoveryTags        map[string]map[string]string // Internal data structure that can enrich metrics with tags
		dimensionsUdObj      map[string]string
		dimensionsUdArr      []map[string]string // Parsed Dimesnsions JSON string (unmarshalled)
		requestDimensions    []map[string]string // this is the actual dimensions list that would be used in API request
		requestDimensionsStr string              // String representation of the above

	}

	// Dimension describe how to get metrics
	Dimension struct {
		Value string `toml:"value"`
	}

	aliyuncmsClient interface {
		DescribeMetricList(request *cms.DescribeMetricListRequest) (response *cms.DescribeMetricListResponse, err error)
	}

	aliyunrdsClient interface {
		DescribeDBInstancePerformance(request *rds.DescribeDBInstancePerformanceRequest) (response *rds.DescribeDBInstancePerformanceResponse, err error)
		DescribeAvailableMetrics(request *rds.DescribeAvailableMetricsRequest) (response *rds.DescribeAvailableMetricsResponse, err error)
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

func (*AliyunCMS) SampleConfig() string {
	return sampleConfig
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
		return fmt.Errorf("failed to retrieve credential: %w", err)
	}
	s.cmsClient, err = cms.NewClientWithOptions("", sdk.NewConfig(), credential)
	if err != nil {
		return fmt.Errorf("failed to create cms client: %w", err)
	}

	// check metrics dimensions consistency
	if s.Project == "acs_rds_dashboard" && slices.Contains(s.MetricServices, "rds") {
		s.rdsClient, err = rds.NewClientWithOptions("", sdk.NewConfig(), credential)
		if err != nil {
			return fmt.Errorf("failed to create rds client: %w", err)
		}
	}

	// check metrics dimensions consistency
	for i := range s.Metrics {
		metric := s.Metrics[i]
		if metric.Dimensions == "" {
			continue
		}
		metric.dimensionsUdObj = make(map[string]string)
		metric.dimensionsUdArr = make([]map[string]string, 0)

		// first try to unmarshal as an object
		if err := json.Unmarshal([]byte(metric.Dimensions), &metric.dimensionsUdObj); err == nil {
			// We were successful, so stop here
			continue
		}

		// then try to unmarshal as an array
		if err := json.Unmarshal([]byte(metric.Dimensions), &metric.dimensionsUdArr); err != nil {
			return fmt.Errorf("cannot parse dimensions (neither obj, nor array) %q: %w", metric.Dimensions, err)
		}
	}

	s.measurement = formatMeasurement(s.Project)

	// Check regions
	if len(s.Regions) == 0 {
		s.Regions = aliyunRegionList
		s.Log.Infof("'regions' is not set. Metrics will be queried across %d regions:\n%s",
			len(s.Regions), strings.Join(s.Regions, ","))
	}

	//Check metric services
	if len(s.MetricServices) == 0 {
		s.MetricServices = []string{"cms"}
		s.Log.Info("'metric_services' is not set. Metrics will be queried from the cms service")
	}

	//Init discovery...
	if s.dt == nil { // Support for tests
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

	// Special setting for acs_oss project since the API differs
	if s.Project == "acs_oss" {
		s.dimensionKey = "BucketName"
	}
	return nil
}

// Start plugin discovery loop, metrics are gathered through Gather
func (s *AliyunCMS) Start(telegraf.Accumulator) error {
	// Start periodic discovery process
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
	for _, m := range s.Metrics {
		// Prepare internal structure with data from discovery
		s.prepareTagsAndDimensions(m)
		wg.Add(len(m.MetricNames))
		for _, metricName := range m.MetricNames {
			<-lmtr.C
			go func(metricName string, m *metric) {
				defer wg.Done()
				acc.AddError(s.gatherMetric(acc, metricName, m))
			}(metricName, m)
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
	// https://help.aliyun.com/document_detail/51936.html?spm=a2c4g.11186623.6.701.54025679zh6wiR
	// The start and end times are executed in the mode of
	// opening left and closing right, and startTime cannot be equal
	// to or greater than endTime.

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
func (s *AliyunCMS) gatherMetric(acc telegraf.Accumulator, metricName string, metric *metric) error {
	for _, region := range s.Regions {
		for more := true; more; {
			var dataPoints []map[string]interface{}
			var err error
			var reqCms *cms.DescribeMetricListRequest
			var respCms *cms.DescribeMetricListResponse

			s.Log.Debugf("METRIC DIMENSIONS: %v\n", metric.requestDimensions)

			if s.rdsClient != nil {
				dataPoints, err = s.getAlicloudRDSPerformanceMetrics(dataPoints, region, metricName, metric)

				if err != nil {
					return fmt.Errorf("failed to convert the performance value string to an float: %w", err)
				}
			} else {
				reqCms, respCms, err = s.getGenericCMSMetrics(region, metricName, metric)

				if err != nil {
					return fmt.Errorf("failed to convert the performance value string to an float: %w", err)
				}

				if err := json.Unmarshal([]byte(respCms.Datapoints), &dataPoints); err != nil {
					return fmt.Errorf("failed to decode response datapoints: %w", err)
				}

				if len(dataPoints) == 0 {
					s.Log.Debugf("No metrics returned from CMS, response msg: %s", respCms.Message)
				}
			}

		NextDataPoint:
			for _, datapoint := range dataPoints {
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
						if reflect.TypeOf(value).String() == "int64" {
							datapointTime = value.(int64)
						} else {
							datapointTime = int64(value.(float64)) / 1000
						}
					default:
						fields[formatField(metricName, key)] = value
					}
				}
				acc.AddFields(s.measurement, fields, tags, time.Unix(datapointTime, 0))
			}
			reqCms.NextToken = respCms.NextToken
			more = reqCms.NextToken != ""
		}
	}
	return nil
}

// Alicloud RDS performance metrics utils
func (s *AliyunMetrics) getAlicloudRDSPerformanceMetrics(dataPoints []map[string]interface{}, region string, metricName string, metric *Metric) ([]map[string]interface{}, error) {
	// TODO Adjust the functionality to call beforehand the Ali API and list all available metrics

	for _, instanceID := range metric.requestDimensions {
		test := rds.CreateDescribeAvailableMetricsRequest()
		test.DBInstanceName = instanceID["instanceId"]
		test.RegionId = region

		// TODO Check the SDK if there are dedicated functions available for the Performance metrics. We should list it here to save api requests and not blow the api rate limit. Maybe we need to reverse engineer the Aliyun API, again.
		respTest, err := s.rdsClient.DescribeAvailableMetrics(test)
		if err != nil {
			return nil, fmt.Errorf("failed to get the database instance performance metrics: %w", err)
		}
		fmt.Println(respTest.BaseResponse.GetHttpContentString())

		req := rds.CreateDescribeDBInstancePerformanceRequest()
		req.DBInstanceId = instanceID["instanceId"]
		req.Key = metricName
		startTime := s.windowStart.UTC()
		req.StartTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02dZ", startTime.Year(), startTime.Month(),
			startTime.Day(), startTime.Hour(), startTime.Minute())
		endTime := s.windowEnd.UTC()
		req.EndTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02dZ", endTime.Year(), endTime.Month(),
			endTime.Day(), endTime.Hour(), endTime.Minute())
		req.RegionId = region

		resp, err := s.rdsClient.DescribeDBInstancePerformance(req)

		if err != nil {
			return nil, fmt.Errorf("failed to get the database instance performance metrics: %w", err)
		}

		if resp.GetHttpStatus() != 200 {
			s.Log.Errorf("failed to get the database instance performance metrics: %v", resp.BaseResponse.GetHttpContentString())
			break
		}

		for _, performanceKey := range resp.PerformanceKeys.PerformanceKey {
			for _, performanceValue := range performanceKey.Values.PerformanceValue {
				parsedTime, err := time.Parse(time.RFC3339, performanceValue.Date)
				if err != nil {
					return nil, fmt.Errorf("failed to parse response performance time datapoints: %w", err)
				}

				if strings.Contains(performanceValue.Value, "&") {
					performanceKeys := strings.Split(performanceKey.ValueFormat, "&")
					performanceValues := strings.Split(performanceValue.Value, "&")

					for i, value := range performanceValues {
						valueAsFloat, err := strconv.ParseFloat(value, 32)
						if err != nil {
							return nil, fmt.Errorf("failed to convert the performance value string to an float: %w", err)
						}
						dataPoints = append(dataPoints,
							map[string]interface{}{
								"instanceId":       instanceID["instanceId"],
								performanceKeys[i]: valueAsFloat,
								"timestamp":        parsedTime.Unix(),
							})
					}
				} else {
					valueAsFloat, err := strconv.ParseFloat(performanceValue.Value, 32)
					if err != nil {
						return nil, fmt.Errorf("failed to convert the performance value string to an float: %w", err)
					}
					dataPoints = append(dataPoints,
						map[string]interface{}{
							"instanceId":               instanceID["instanceId"],
							performanceKey.ValueFormat: valueAsFloat,
							"timestamp":                parsedTime.Unix(),
						})
				}
			}
		}

		if len(dataPoints) == 0 {
			s.Log.Debugf("No rds performance metrics returned from RDS, response msg: %s", resp.GetHttpContentString())
			break
		}
	}
	return dataPoints, nil
}

func (s *AliyunMetrics) getGenericCMSMetrics(region, metricName string, metric *Metric) (*cms.DescribeMetricListRequest, *cms.DescribeMetricListResponse, error) {
	reqCms := cms.CreateDescribeMetricListRequest()
	reqCms.Period = strconv.FormatInt(int64(time.Duration(s.Period).Seconds()), 10)
	reqCms.MetricName = metricName
	reqCms.Length = "10000"
	reqCms.Namespace = s.Project
	reqCms.EndTime = strconv.FormatInt(s.windowEnd.Unix()*1000, 10)
	reqCms.StartTime = strconv.FormatInt(s.windowStart.Unix()*1000, 10)
	reqCms.Dimensions = metric.requestDimensionsStr
	reqCms.RegionId = region

	respCms, err := s.cmsClient.DescribeMetricList(reqCms)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to query metricName list: %w", err)
	}
	if respCms.Code != "200" {
		s.Log.Errorf("failed to query metric name list: %v", respCms.Message)
	}

	return reqCms, respCms, nil
}

// tag helper
func parseTag(tagSpec string, data interface{}) (tagKey, tagValue string, err error) {
	var (
		ok        bool
		queryPath = tagSpec
	)
	tagKey = tagSpec

	// Split query path to tagKey and query path
	if splitted := strings.Split(tagSpec, ":"); len(splitted) == 2 {
		tagKey = splitted[0]
		queryPath = splitted[1]
	}

	tagRawValue, err := jmespath.Search(queryPath, data)
	if err != nil {
		return "", "", fmt.Errorf("can't query data from discovery data using query path %q: %w", queryPath, err)
	}

	if tagRawValue == nil { // Nothing found
		return "", "", nil
	}

	tagValue, ok = tagRawValue.(string)
	if !ok {
		return "", "", fmt.Errorf("tag value %q parsed by query %q is not a string value", tagRawValue, queryPath)
	}

	return tagKey, tagValue, nil
}

func (s *AliyunCMS) prepareTagsAndDimensions(metric *metric) {
	var (
		newData     bool
		defaultTags = []string{"RegionId:RegionId"}
	)

	if s.dt == nil { // Discovery is not activated
		return
	}

	// Reading all data from buffered channel
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

	// new data arrives (so process it) or this is the first call
	if newData || len(metric.discoveryTags) == 0 {
		metric.dtLock.Lock()
		defer metric.dtLock.Unlock()

		if metric.discoveryTags == nil {
			metric.discoveryTags = make(map[string]map[string]string, len(s.discoveryData))
		}

		metric.requestDimensions = nil // erasing
		metric.requestDimensions = make([]map[string]string, 0, len(s.discoveryData))

		// Preparing tags & dims...
		for instanceID, elem := range s.discoveryData {
			// Start filing tags
			// Remove old value if exist
			delete(metric.discoveryTags, instanceID)
			metric.discoveryTags[instanceID] = make(map[string]string, len(metric.TagsQueryPath)+len(defaultTags))

			for _, tagQueryPath := range metric.TagsQueryPath {
				tagKey, tagValue, err := parseTag(tagQueryPath, elem)
				if err != nil {
					s.Log.Errorf("%v", err)
					continue
				}
				if err == nil && tagValue == "" { // Nothing found
					s.Log.Debugf("Data by query path %q: is not found, for instance %q", tagQueryPath, instanceID)
					continue
				}

				metric.discoveryTags[instanceID][tagKey] = tagValue
			}

			// Adding default tags if not already there
			for _, defaultTagQP := range defaultTags {
				tagKey, tagValue, err := parseTag(defaultTagQP, elem)

				if err != nil {
					s.Log.Errorf("%v", err)
					continue
				}

				if err == nil && tagValue == "" { // Nothing found
					s.Log.Debugf("Data by query path %q: is not found, for instance %q",
						defaultTagQP, instanceID)
					continue
				}

				metric.discoveryTags[instanceID][tagKey] = tagValue
			}

			// if no dimension configured in config file, use discovery data
			if len(metric.dimensionsUdArr) == 0 && len(metric.dimensionsUdObj) == 0 {
				metric.requestDimensions = append(
					metric.requestDimensions,
					map[string]string{s.dimensionKey: instanceID})
			}
		}

		// add dimensions filter from config file
		if len(metric.dimensionsUdArr) != 0 {
			metric.requestDimensions = append(metric.requestDimensions, metric.dimensionsUdArr...)
		}
		if len(metric.dimensionsUdObj) != 0 {
			metric.requestDimensions = append(metric.requestDimensions, metric.dimensionsUdObj)
		}

		// Unmarshalling to string
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
func formatField(metricName, statistic string) string {
	if metricName == statistic {
		statistic = "value"
	}
	return fmt.Sprintf("%s_%s", snakeCase(metricName), snakeCase(statistic))
}

func formatMeasurement(project string) string {
	project = strings.ReplaceAll(project, "/", "_")
	project = snakeCase(project)
	return "aliyuncms_" + project
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.ReplaceAll(s, "__", "_")
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
