//go:generate ../../../tools/readme_config_includer/generator
package aliyunrds

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"
	common_aliyun "github.com/influxdata/telegraf/plugins/common/aliyun"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type (
	AliyunRDS struct {
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
		Metrics           []*metric       `toml:"metrics"`
		RateLimit         int             `toml:"ratelimit"`

		Log telegraf.Logger `toml:"-"`

		rdsClient aliyunrdsClient

		windowStart time.Time
		windowEnd   time.Time
		dt          *discoveryTool
	}

	metric struct {
		MetricNames []string `toml:"names"`
		Instances   []string `toml:"instances"` // List of RDS instance IDs to monitor

		dtLock        sync.Mutex
		instanceList  []string                     // Combined list of instances (discovered + configured)
		discoveryTags map[string]map[string]string // Tags from discovery
	}

	aliyunrdsClient interface {
		DescribeDBInstancePerformance(request *rds.DescribeDBInstancePerformanceRequest) (response *rds.DescribeDBInstancePerformanceResponse, err error)
	}
)

// https://www.alibabacloud.com/help/doc-detail/40654.htm
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

func (*AliyunRDS) SampleConfig() string {
	return sampleConfig
}

// Init performs checks of plugin inputs and initializes internals
func (s *AliyunRDS) Init() error {
	credential, err := common_aliyun.GetCredentials(common_aliyun.CredentialConfig{
		AccessKeyID:       s.AccessKeyID,
		AccessKeySecret:   s.AccessKeySecret,
		AccessKeyStsToken: s.AccessKeyStsToken,
		RoleArn:           s.RoleArn,
		RoleSessionName:   s.RoleSessionName,
		PrivateKey:        s.PrivateKey,
		PublicKeyID:       s.PublicKeyID,
		RoleName:          s.RoleName,
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve credential: %w", err)
	}

	s.rdsClient, err = rds.NewClientWithOptions("", sdk.NewConfig(), credential)
	if err != nil {
		return fmt.Errorf("failed to create rds client: %w", err)
	}

	// Check regions
	if len(s.Regions) == 0 {
		s.Regions = common_aliyun.DefaultRegions()
		s.Log.Infof("'regions' is not set. Metrics will be queried across %d regions:\n%s",
			len(s.Regions), strings.Join(s.Regions, ","))
	}

	// Check metrics
	if len(s.Metrics) == 0 {
		return errors.New("at least one metric must be configured")
	}

	// Init discovery...
	s.dt, err = newDiscoveryTool(s.Regions, s.Log, credential, int(float32(s.RateLimit)*0.2), time.Duration(s.DiscoveryInterval))
	if err != nil {
		s.Log.Warnf("Discovery tool is not activated: %v", err)
		s.dt = nil
	} else {
		discoveryData, err := s.dt.getDiscoveryDataAcrossRegions(nil)
		if err != nil {
			s.Log.Warnf("Discovery tool is not activated: %v", err)
			s.dt = nil
		} else {
			s.Log.Infof("%d RDS instance(s) discovered...", len(discoveryData))
		}
	}

	return nil
}

// Start a plugin discovery loop, metrics are gathered through Gather
func (s *AliyunRDS) Start(telegraf.Accumulator) error {
	// Start a periodic discovery process
	if s.dt != nil {
		s.dt.start()
	}

	return nil
}

// Gather implements telegraf.Inputs interface
func (s *AliyunRDS) Gather(acc telegraf.Accumulator) error {
	s.updateWindow(time.Now())
	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)
	defer lmtr.Stop()

	var wg sync.WaitGroup
	for _, m := range s.Metrics {
		// Prepare instance list from discovery and configuration
		s.prepareInstances(m)
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
func (s *AliyunRDS) Stop() {
	if s.dt != nil {
		s.dt.stop()
	}
}

// updateWindow updates the time window for metrics collection
func (s *AliyunRDS) updateWindow(relativeTo time.Time) {
	windowEnd := relativeTo.Add(-time.Duration(s.Delay))

	if s.windowEnd.IsZero() {
		// this is the first run, no window info, so just get a single period
		s.windowStart = windowEnd.Add(-time.Duration(s.Period))
	} else {
		// subsequent window, start where the last window left off
		s.windowStart = s.windowEnd
	}

	s.windowEnd = windowEnd
}

// gatherMetric collects metrics for a specific metric name
func (s *AliyunRDS) gatherMetric(acc telegraf.Accumulator, metricName string, m *metric) error {
	for _, region := range s.Regions {
		for _, instanceID := range m.instanceList {
			dataPoints, err := s.fetchRDSPerformanceDatapoints(region, metricName, instanceID)
			if err != nil {
				return err
			}

			if len(dataPoints) == 0 {
				s.Log.Debugf("No RDS performance metrics returned for instance %s in region %s", instanceID, region)
				continue
			}

			for _, dp := range dataPoints {
				fields := make(map[string]interface{}, len(dp))
				tags := make(map[string]string)
				var ts int64

				for key, value := range dp {
					switch key {
					case "instanceId":
						tags[key] = value.(string)
						// Add discovery tags if available
						if m.discoveryTags != nil {
							if discTags, ok := m.discoveryTags[value.(string)]; ok {
								for k, v := range discTags {
									tags[k] = v
								}
							}
						}
					case "timestamp":
						ts = value.(int64)
					default:
						fields[formatField(metricName, key)] = value
					}
				}

				// Add region tag
				tags["region"] = region

				acc.AddFields("aliyunrds", fields, tags, time.Unix(ts, 0))
			}
		}
	}
	return nil
}

// fetchRDSPerformanceDatapoints queries RDS performance metrics
func (s *AliyunRDS) fetchRDSPerformanceDatapoints(region, metricName, instanceID string) ([]map[string]interface{}, error) {
	var dataPoints []map[string]interface{}

	req := rds.CreateDescribeDBInstancePerformanceRequest()
	req.DBInstanceId = instanceID
	req.Key = metricName
	req.StartTime = s.windowStart.UTC().Format("2006-01-02T15:04Z")
	req.EndTime = s.windowEnd.UTC().Format("2006-01-02T15:04Z")
	req.RegionId = region

	resp, err := s.rdsClient.DescribeDBInstancePerformance(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get the database instance performance metrics: %w", err)
	}
	httpStatus := resp.GetHttpStatus()
	if httpStatus != 0 && httpStatus != 200 {
		return nil, fmt.Errorf("failed to get the database instance performance metrics: status %d, %v",
			httpStatus, resp.BaseResponse.GetHttpContentString())
	}

	for _, performanceKey := range resp.PerformanceKeys.PerformanceKey {
		keyNames := strings.Split(performanceKey.ValueFormat, "&")
		for _, performanceValue := range performanceKey.Values.PerformanceValue {
			parsedTime, err := time.Parse(time.RFC3339, performanceValue.Date)
			if err != nil {
				return nil, fmt.Errorf("failed to parse response performance time datapoints: %w", err)
			}

			values := strings.Split(performanceValue.Value, "&")
			if len(values) == 1 && len(keyNames) == 1 {
				valueAsFloat, err := strconv.ParseFloat(values[0], 32)
				if err != nil {
					return nil, fmt.Errorf("failed to convert the performance value string to a float: %w", err)
				}
				dataPoints = append(dataPoints, map[string]interface{}{
					"instanceId":               instanceID,
					performanceKey.ValueFormat: valueAsFloat,
					"timestamp":                parsedTime.Unix(),
				})
				continue
			}

			// Multiple values with "&" separated lists
			for i, v := range values {
				if i >= len(keyNames) {
					break
				}
				valueAsFloat, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to convert the performance value string to a float: %w", err)
				}
				dataPoints = append(dataPoints, map[string]interface{}{
					"instanceId": instanceID,
					keyNames[i]:  valueAsFloat,
					"timestamp":  parsedTime.Unix(),
				})
			}
		}
	}

	return dataPoints, nil
}

// prepareInstances prepares the instance list from discovery and configuration
func (s *AliyunRDS) prepareInstances(m *metric) {
	m.dtLock.Lock()
	defer m.dtLock.Unlock()

	// Start with configured instances
	instanceMap := make(map[string]bool)
	for _, inst := range m.Instances {
		instanceMap[inst] = true
	}

	// Add discovered instances if discovery is enabled
	if s.dt != nil {
		var newData bool
		var discoveryData map[string]interface{}

		// First check if there's new data in the channel
	L:
		for {
			select {
			case discoveryData = <-s.dt.DataChan:
				newData = true
				continue
			default:
				break L
			}
		}

		// Process discovery data if we have any (new or need initial processing)
		if newData || (len(m.instanceList) == 0 && discoveryData == nil) {
			// If no new data but need to initialize, try to get initial discovery data
			if discoveryData == nil && len(m.instanceList) == 0 {
				var err error
				discoveryData, err = s.dt.getDiscoveryDataAcrossRegions(nil)
				if err != nil {
					s.Log.Debugf("Failed to get initial discovery data: %v", err)
				}
			}

			if discoveryData != nil {
				if m.discoveryTags == nil {
					m.discoveryTags = make(map[string]map[string]string)
				}

				// Process discovered instances
				for instanceID, data := range discoveryData {
					instanceMap[instanceID] = true
					// Store discovery data as tags
					if dataMap, ok := data.(map[string]interface{}); ok {
						tags := make(map[string]string)
						tags["RegionId"] = getStringValue(dataMap, "RegionId")
						tags["DBInstanceType"] = getStringValue(dataMap, "DBInstanceType")
						tags["Engine"] = getStringValue(dataMap, "Engine")
						tags["EngineVersion"] = getStringValue(dataMap, "EngineVersion")
						tags["DBInstanceDescription"] = getStringValue(dataMap, "DBInstanceDescription")
						m.discoveryTags[instanceID] = tags
					}
				}
			}

			// Rebuild instance list
			m.instanceList = make([]string, 0, len(instanceMap))
			for inst := range instanceMap {
				m.instanceList = append(m.instanceList, inst)
			}
		}
	} else {
		// No discovery, use only configured instances
		m.instanceList = make([]string, 0, len(instanceMap))
		for inst := range instanceMap {
			m.instanceList = append(m.instanceList, inst)
		}
	}
}

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// formatField formats the field name from metric name and statistic
func formatField(metricName, statistic string) string {
	if metricName == statistic {
		statistic = "value"
	}
	return fmt.Sprintf("%s_%s", snakeCase(metricName), snakeCase(statistic))
}

// snakeCase converts a string to snake_case
func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.ReplaceAll(s, "__", "_")
	return s
}

func init() {
	inputs.Add("aliyunrds", func() telegraf.Input {
		return &AliyunRDS{
			RateLimit:         200,
			DiscoveryInterval: config.Duration(time.Minute),
		}
	})
}
