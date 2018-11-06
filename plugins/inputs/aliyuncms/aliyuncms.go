package aliyuncms

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/limiter"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
)

const (
	description  = "Pull Metric Statistics from Aliyun CMS"
	sampleConfig = `
## Aliyun Region
## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
region_id = "cn-hangzhou"

## Aliyun Credentials
## Credentials are loaded in the following order
## 1) Ram RoleArn credential
## 2) AccessKey STS token credential
## 3) AccessKey credential
## 4) Ecs Ram Role credential
## 5) RSA keypair cendential
## 6) Environment varialbes credential
## 7) Instance metadata credential
# access_key_id = ""
# access_key_secret = ""
# access_key_sts_token = ""
# role_arn = ""
# role_session_name = ""
# private_key = ""
# public_key_id = ""
# role_name = ""

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

## Metrics to Pull (Required)
## Defaults to all Metrics in Namespace if nothing is provided
## Refreshes Namespace available metrics every 1h
[[inputs.aliyuncms.metrics]]
  names = ["InstanceActiveConnection", "InstanceNewConnection"]

  ## Dimension filters for Metric.  These are optional however all dimensions
  ## defined for the metric names must be specified in order to retrieve
  ## the metric statistics.
  ## See: https://help.aliyun.com/document_detail/28619.html?spm=a2c4g.11186623.2.11.6ac47694AjhHt4
  [[inputs.aliyuncms.metrics.dimensions]]
    value = '{"instanceId": "p-example"}'

  [[inputs.aliyuncms.metrics.dimensions]]
    value = '{"instanceId": "q-example"}'
`
)

type (
	// AliyunCMS is aliyun cms config info.
	AliyunCMS struct {
		RegionID          string `toml:"region_id"`
		AccessKeyID       string `toml:"access_key_id"`
		AccessKeySecret   string `toml:"access_key_secret"`
		AccessKeyStsToken string `toml:"access_key_sts_token"`
		RoleArn           string `toml:"role_arn"`
		RoleSessionName   string `toml:"role_session_name"`
		PrivateKey        string `toml:"private_key"`
		PublicKeyID       string `toml:"public_key_id"`
		RoleName          string `toml:"role_name"`

		Period    internal.Duration `toml:"period"`
		Delay     internal.Duration `toml:"delay"`
		Project   string            `toml:"project"`
		Metrics   []*Metric         `toml:"metrics"`
		RateLimit int               `toml:"ratelimit"`

		client      aliyuncmsClient
		windowStart time.Time
		windowEnd   time.Time
	}

	// Metric describes what metrics to get
	Metric struct {
		MetricNames []string     `toml:"names"`
		Dimensions  []*Dimension `toml:"dimensions"`
	}

	// Dimension describe how to get metrics
	Dimension struct {
		Value string `toml:"value"`
	}

	aliyuncmsClient interface {
		QueryMetricList(request *cms.QueryMetricListRequest) (*cms.QueryMetricListResponse, error)
	}
)

// SampleConfig implements telegraf.Inputs interface
func (s *AliyunCMS) SampleConfig() string {
	return sampleConfig
}

// Description implements telegraf.Inputs interface
func (s *AliyunCMS) Description() string {
	return description
}

// Gather implements telegraf.Inputs interface
func (s *AliyunCMS) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		if err := s.initializeAliyunCMS(); err != nil {
			return err
		}
	}

	s.updateWindow(time.Now())

	// limit concurrency or we can easily exhaust user connection limit
	lmtr := limiter.NewRateLimiter(s.RateLimit, time.Second)
	defer lmtr.Stop()

	var wg sync.WaitGroup
	for _, metric := range s.Metrics {
		for _, metricName := range metric.MetricNames {
			wg.Add(len(metric.Dimensions))
			for _, dimension := range metric.Dimensions {
				<-lmtr.C
				go func(m string, d *Dimension) {
					defer wg.Done()
					acc.AddError(s.gatherMetric(acc, m, d))
				}(metricName, dimension)
			}
			wg.Wait()
		}
	}

	return nil
}

func (s *AliyunCMS) updateWindow(relativeTo time.Time) {
	windowEnd := relativeTo.Add(-s.Delay.Duration)

	if s.windowEnd.IsZero() {
		// this is the first run, no window info, so just get a single period
		s.windowStart = windowEnd.Add(-s.Period.Duration)
	} else {
		// subsequent window, start where last window left off
		s.windowStart = s.windowEnd
	}

	s.windowEnd = windowEnd
}

func init() {
	inputs.Add("aliyuncms", func() telegraf.Input {
		return &AliyunCMS{
			RateLimit: 200,
		}
	})
}

// Initialize AliyunCMS client
func (s *AliyunCMS) initializeAliyunCMS() error {
	if s.RegionID == "" {
		return errors.New("region id is not set")
	}
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
		return errors.New("failed to retrieve credential")
	}
	s.client, err = cms.NewClientWithOptions(s.RegionID, sdk.NewConfig(), credential)
	if err != nil {
		return fmt.Errorf("failed to create cms client: %v", err)
	}

	return nil
}

// Gather given metric and emit error
func (s *AliyunCMS) gatherMetric(acc telegraf.Accumulator, metric string, dimension *Dimension) error {
	tags := make(map[string]string)
	if err := json.Unmarshal([]byte(dimension.Value), &tags); err != nil {
		return fmt.Errorf("failed to decode %s: %v", dimension.Value, err)
	}
	tags["regionId"] = s.RegionID

	req := cms.CreateQueryMetricListRequest()
	req.Period = strconv.FormatInt(int64(s.Period.Duration.Seconds()), 10)
	req.Metric = metric
	req.Length = "10000"
	req.Project = s.Project
	req.EndTime = strconv.FormatInt(s.windowEnd.Unix()*1000, 10)
	req.StartTime = strconv.FormatInt(s.windowStart.Unix()*1000, 10)
	req.Dimensions = dimension.Value

	for more := true; more; {
		resp, err := s.client.QueryMetricList(req)
		if err != nil {
			return fmt.Errorf("failed to query metric list: %v", err)
		} else if resp.Code != "200" {
			return fmt.Errorf("failed to query metric list: %v", resp.Message)
		}

		if len(resp.Datapoints) == 0 {
			break
		}

		var datapoints []map[string]interface{}
		if err = json.Unmarshal([]byte(resp.Datapoints), &datapoints); err != nil {
			return fmt.Errorf("failed to decode response datapoints: %v", err)
		}

		for _, datapoint := range datapoints {
			fields := make(map[string]interface{})

			if average, ok := datapoint["Average"]; ok {
				fields[formatField(metric, "Average")] = average
			}
			if minimum, ok := datapoint["Minimum"]; ok {
				fields[formatField(metric, "Minimum")] = minimum
			}
			if maximum, ok := datapoint["Maximum"]; ok {
				fields[formatField(metric, "Maximum")] = maximum
			}
			if value, ok := datapoint["Value"]; ok {
				fields[formatField(metric, "Value")] = value
			}
			tags["userId"] = datapoint["userId"].(string)

			datapointTime := int64(datapoint["timestamp"].(float64)) / 1000
			acc.AddFields(formatMeasurement(s.Project), fields, tags, time.Unix(datapointTime, 0))
		}

		req.Cursor = resp.Cursor
		more = req.Cursor != ""
	}

	return nil
}

// Formatting helpers
func formatField(metricName string, statistic string) string {
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
