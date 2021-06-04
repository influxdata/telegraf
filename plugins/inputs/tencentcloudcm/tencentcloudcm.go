package tencentcloudcm

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/influxdata/telegraf/plugins/inputs"
)

// TencentCloudCM contains the configuration for the Tencent Cloud Cloud Monitor plugin.
type TencentCloudCM struct {
	Endpoint string `toml:"endpoint"`

	Period config.Duration `toml:"period"`
	Delay  config.Duration `toml:"delay"`

	RateLimit int             `toml:"ratelimit"`
	Timeout   config.Duration `toml:"timeout"`

	DiscoveryInterval config.Duration `toml:"discovery_interval"`

	Accounts []*Account `toml:"accounts"`

	client cmClient

	Log telegraf.Logger `toml:"-"`

	windowStart  time.Time
	windowEnd    time.Time
	discoverTool *DiscoverTool
}

// Account defines a Tencent Cloud account
type Account struct {
	Name       string       `toml:"name"`
	SecretID   string       `toml:"secret_id"`
	SecretKey  string       `toml:"secret_key"`
	Namespaces []*Namespace `toml:"namespaces"`

	Crs *common.Credential
}

// Namespace defines a Tencent Cloud CM namespace
type Namespace struct {
	Name    string    `toml:"namespace"`
	Metrics []string  `toml:"metrics"`
	Regions []*Region `toml:"regions"`
}

// Region defines a Tencent Cloud region
type Region struct {
	RegionName string      `toml:"region"`
	Instances  []*Instance `toml:"instances"`
}

// Instance defines a generic Tencent Cloud instance
type Instance struct {
	Dimensions []*Dimension `toml:"dimensions"`
}

// Dimension defines a simplified Tencent Cloud Cloud Monitor dimension.
type Dimension struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

// MetricObject defines a metric with additional information
type MetricObject struct {
	Metric    string
	Region    string
	Namespace string
	Account   *Account
	Instances []*Instance
}

type cmClient interface {
	GetMetricObjects(t TencentCloudCM) []MetricObject
	NewClient(region string, crs *common.Credential, t TencentCloudCM) Client
	NewGetMonitorDataRequest(namespace, metric string, instances []*Instance, t TencentCloudCM) *GetMonitorDataRequest
	GatherMetrics(client Client, request *GetMonitorDataRequest, t TencentCloudCM) (*GetMonitorDataResponse, error)
}

// SampleConfig implements telegraf.Input interface
func (t *TencentCloudCM) SampleConfig() string {
	return `
	## Endpoint to make request against, the correct endpoint is automatically
	## determined and this option should only be set if you wish to override the
	## default.
	##   ex: endpoint = "tencentcloudapi.com"
	# endpoint = ""
  
	# The default period for Tencent Cloud Cloud Monitor metrics is 1 minute (60s). However not all
	# metrics are made available to the 1 minute period. Some are collected at
	# 5 minute, 60 minute, or larger intervals.
	# See: https://intl.cloud.tencent.com/document/product/248/33882
	# Note that if a period is configured that is smaller than the default for a
	# particular metric, that metric will not be returned by the Tencent Cloud API
	# and will not be collected by Telegraf.
	#
	# Requested Tencent Cloud Cloud Monitor aggregation Period (required - must be a multiple of 60s)
	period = "1m"
  
	# Collection Delay (must account for metrics availability via Tencent Cloud API)
	delay = "5m"
  
	## Maximum requests per second. Note that the global default Tencent Cloud API rate limit is
	## 20 calls/second (1,200 calls/minute), so if you define multiple namespaces, these should add up to a
	## maximum of 20.
	## See https://intl.cloud.tencent.com/document/product/248/33881
	ratelimit = 1000
  
	# Timeout for http requests made by the Tencent Cloud client.
	timeout = "5s"
  
	## By default, Tencent Cloud CM Input plugin will automatically discover instances in specified regions
	## This sets the interval for discover and update the instances discovered.
	##
	## how often the discovery API call executed (default 1m)
	# discovery_interval = "1m"
  
	# Tencent Cloud Account (required - you can provide multiple entries and distinguish them using
	# optional name field, if name is empty, index number will be used as default)
	[[inputs.tencentcloudcm.accounts]]
	  name = ""
	  secret_id = ""
	  secret_key = ""
  
	  # Namespaces to Pull
	  [[inputs.tencentcloudcm.accounts.namespaces]]
		# Tencent Cloud CM Namespace (required - see https://intl.cloud.tencent.com/document/product/248/34716#namespace)
		namespace = "QCE/CVM"
  
		# Metrics filter, all metrics will be pulled if not left empty. Different namespaces may have different
		# metric names, e.g. CVM Monitoring Metrics: https://intl.cloud.tencent.com/document/product/248/6843
		# metrics = ["CPUUsage", "MemUsage"]
  
		[[inputs.tencentcloudcm.accounts.namespaces.regions]]
		  # Tencent Cloud regions (required - Allowed values: https://intl.cloud.tencent.com/document/api/248/33876)
		  region = "ap-guangzhou"
  
		  ## Dimension filters for Metric. Different namespaces may have different
		  ## dimension requirements, e.g. CVM Monitoring Metrics: https://intl.cloud.tencent.com/document/product/248/6843It must be specified if the namespace does not support instance auto discovery
		  ## Currently, discovery supported for the following namespaces:
		  ## - QCE/CVM
		  ## - QCE/CDB
		  ## - QCE/CES
		  ## - QCE/REDIS
		  ## - QCE/LB_PUBLIC
		  ## - QCE/LB_PRIVATE
		  # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances]]
		  # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances.dimensions]]
		  #   name = ""
		  #   value = ""
`
}

// Description implements telegraf.Input interface
func (t *TencentCloudCM) Description() string {
	return "Pull Metric Statistics from Tencent Cloud Cloud Monitor"
}

// Init is for setup, and validating config.
func (t *TencentCloudCM) Init() error {

	if t.Period <= 0 {
		return fmt.Errorf("period is empty")
	}

	if len(t.Accounts) == 0 {
		return errors.New("account is empty")
	}

	// create account credential
	for i := range t.Accounts {

		if t.Accounts[i].SecretID == "" {
			return fmt.Errorf("secret_id is empty")
		}
		if t.Accounts[i].SecretKey == "" {
			return fmt.Errorf("secret_key is empty")
		}

		t.Accounts[i].Crs = common.NewCredential(
			t.Accounts[i].SecretID,
			t.Accounts[i].SecretKey,
		)
		if t.Accounts[i].Name == "" {
			t.Accounts[i].Name = fmt.Sprintf("%v", i)
		}
		// check if namespace supports auto discovery
		for _, namespace := range t.Accounts[i].Namespaces {
			if namespace.Name == "" {
				return fmt.Errorf("namespace is empty")
			}
			for _, region := range namespace.Regions {
				if region.RegionName == "" {
					return fmt.Errorf("region is empty")
				}
				_, ok := Registry[namespace.Name]
				if len(region.Instances) == 0 && !ok {
					return fmt.Errorf(
						"unsupported namespace %s for discovering instances, please specify instances and dimensions",
						namespace.Name,
					)
				}
			}
		}
	}

	// start discovery
	t.discoverTool = NewDiscoverTool(t.Log)
	go t.discoverTool.Discover(t.Accounts, t.DiscoveryInterval, t.Endpoint)
	t.discoverTool.DiscoverMetrics()
	for i := range t.Accounts {
		for j := range t.Accounts[i].Namespaces {
			// use discovered metrics if not specified
			if len(t.Accounts[i].Namespaces[j].Metrics) == 0 {
				metrics, ok := t.discoverTool.DiscoveredMetrics[t.Accounts[i].Namespaces[j].Name]
				if !ok {
					errorInfo := fmt.Sprintf("unsupported namespace %s for discovering metrics, please specify metrics",
						t.Accounts[i].Namespaces[j].Name)
					t.Log.Error(errorInfo)
					return errors.New(errorInfo)
				}
				t.Accounts[i].Namespaces[j].Metrics = metrics
			}
		}
	}

	t.client = &cloudmonitorClient{Accounts: t.Accounts}

	return nil

}

func (t *TencentCloudCM) updateWindow(relativeTo time.Time) {
	windowEnd := relativeTo.Add(-time.Duration(t.Delay))

	t.windowStart = windowEnd.Add(-time.Duration(t.Period) * 2)

	t.windowEnd = windowEnd
}

// GetMetricObjects gets all metric objects
func (t *TencentCloudCM) GetMetricObjects() []MetricObject {

	// metricObejcts holds all metrics with it's corresponding region,
	// namespace, credential and instances(dimensions) information.
	metricObjects := []MetricObject{}

	// construct metric object
	for i := range t.Accounts {
		for j := range t.Accounts[i].Namespaces {
			for k := range t.Accounts[i].Namespaces[j].Regions {
				for l := range t.Accounts[i].Namespaces[j].Metrics {
					instances := t.Accounts[i].Namespaces[j].Regions[k].Instances
					if len(instances) == 0 {
						instances = t.discoverTool.GetInstances(t.Accounts[i].Name, t.Accounts[i].Namespaces[j].Name, t.Accounts[i].Namespaces[j].Regions[k].RegionName)
					}
					if len(instances) == 0 {
						continue
					}
					metricObjects = append(metricObjects, MetricObject{
						t.Accounts[i].Namespaces[j].Metrics[l],
						t.Accounts[i].Namespaces[j].Regions[k].RegionName,
						t.Accounts[i].Namespaces[j].Name,
						t.Accounts[i],
						instances,
					})
				}
			}
		}
	}
	return metricObjects
}

func (t *TencentCloudCM) Gather(acc telegraf.Accumulator) error {

	t.updateWindow(time.Now())

	metricObjects := t.client.GetMetricObjects(*t)

	lmtr := limiter.NewRateLimiter(t.RateLimit, time.Second)
	defer lmtr.Stop()

	wg := sync.WaitGroup{}
	rLock := sync.Mutex{}
	results := []GetMonitorDataResponse{}

	// requestIDMap contains request ID and metric objects for later aggregation
	requestIDMap := map[string]MetricObject{}

	for i := range metricObjects {

		wg.Add(1)
		<-lmtr.C
		go func(m MetricObject) {
			defer wg.Done()

			client := t.client.NewClient(m.Region, m.Account.Crs, *t)
			request := t.client.NewGetMonitorDataRequest(m.Namespace, m.Metric, m.Instances, *t)

			result, err := t.client.GatherMetrics(client, request, *t)
			if err != nil {
				acc.AddError(err)
				return
			}

			rLock.Lock()
			requestIDMap[*result.Response.RequestId] = m
			rLock.Unlock()

			rLock.Lock()
			results = append(results, *result)
			rLock.Unlock()

		}(metricObjects[i])

	}
	wg.Wait()

	for i := range results {
		for j := range results[i].Response.DataPoints {
			tags := map[string]string{}
			for k := range results[i].Response.DataPoints[j].Dimensions {
				tags[*results[i].Response.DataPoints[j].Dimensions[k].Name] = *results[i].Response.DataPoints[j].Dimensions[k].Value
			}
			metricObject := requestIDMap[*results[i].Response.RequestId]
			tags["account"] = metricObject.Account.Name
			tags["namespace"] = metricObject.Namespace
			tags["region"] = metricObject.Region
			tags["period"] = fmt.Sprintf("%v", *results[i].Response.Period)
			tags["metric"] = *results[i].Response.MetricName
			tags["request_id"] = *results[i].Response.RequestId

			for index := range results[i].Response.DataPoints[j].Values {
				acc.AddFields(metricObject.Namespace, map[string]interface{}{"value": results[i].Response.DataPoints[j].Values[index]}, tags, time.Unix(int64(*results[i].Response.DataPoints[j].Timestamps[index]), 0).Local())
			}

		}
	}

	return nil
}

func init() {
	inputs.Add("tencentcloudcm", func() telegraf.Input {
		return New()
	})
}

// New instance of the Tencent Cloud Cloud Monitor plugin
func New() *TencentCloudCM {
	return &TencentCloudCM{
		Endpoint:          "tencentcloudapi.com",
		RateLimit:         20,
		Timeout:           config.Duration(5 * time.Second),
		DiscoveryInterval: config.Duration(1 * time.Minute),
	}
}

type cloudmonitorClient struct {
	Accounts []*Account `toml:"accounts"`
}

func (c *cloudmonitorClient) GetMetricObjects(t TencentCloudCM) []MetricObject {
	// metricObejcts holds all metrics with it's corresponding region,
	// namespace, credential and instances(dimensions) information.
	metricObjects := []MetricObject{}

	// construct metric object
	for i := range t.Accounts {
		for j := range t.Accounts[i].Namespaces {
			for k := range t.Accounts[i].Namespaces[j].Regions {
				for l := range t.Accounts[i].Namespaces[j].Metrics {
					instances := t.Accounts[i].Namespaces[j].Regions[k].Instances
					if len(instances) == 0 {
						instances = t.discoverTool.GetInstances(t.Accounts[i].Name, t.Accounts[i].Namespaces[j].Name, t.Accounts[i].Namespaces[j].Regions[k].RegionName)
					}
					if len(instances) == 0 {
						continue
					}
					metricObjects = append(metricObjects, MetricObject{
						t.Accounts[i].Namespaces[j].Metrics[l],
						t.Accounts[i].Namespaces[j].Regions[k].RegionName,
						t.Accounts[i].Namespaces[j].Name,
						t.Accounts[i],
						instances,
					})
				}
			}
		}
	}
	return metricObjects
}

func (c *cloudmonitorClient) NewClient(region string, crs *common.Credential, t TencentCloudCM) Client {
	client := Client{}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = fmt.Sprintf("monitor.%s", t.Endpoint)
	cpf.HttpProfile.ReqTimeout = int(time.Duration(t.Timeout).Milliseconds()) / 1000
	client.Init(region).WithCredential(crs).WithProfile(cpf)
	return client
}

func (c *cloudmonitorClient) NewGetMonitorDataRequest(namespace, metric string, instances []*Instance, t TencentCloudCM) *GetMonitorDataRequest {
	request := NewGetMonitorDataRequest()
	request.Namespace = common.StringPtr(namespace)
	request.MetricName = common.StringPtr(metric)
	period := uint64(time.Duration(t.Period).Seconds())
	request.Period = &period
	request.StartTime = common.StringPtr(t.windowStart.Format(time.RFC3339))
	request.EndTime = common.StringPtr(t.windowEnd.Format(time.RFC3339))
	request.Instances = []*MonitorInstance{}
	// Transform instances and dimensions from config to monitor struct
	for i := range instances {
		request.Instances = append(request.Instances, &MonitorInstance{})
		for j := range instances[i].Dimensions {
			request.Instances[i].Dimensions = append(request.Instances[i].Dimensions, &MonitorDimension{
				Name:  &instances[i].Dimensions[j].Name,
				Value: &instances[i].Dimensions[j].Value,
			})
		}
	}
	return request
}

func (c *cloudmonitorClient) GatherMetrics(client Client, request *GetMonitorDataRequest, t TencentCloudCM) (*GetMonitorDataResponse, error) {
	response, err := client.GetMonitorData(request)
	if val, ok := err.(*tcerrors.TencentCloudSDKError); ok {
		t.Log.Errorf("An API error has returned for %s: %s", *request.Namespace, err)
		return nil, errors.New(val.Error())
	}
	if err != nil {
		t.Log.Errorf("GetMonitorData failed, error: %s", err)
		return nil, err
	}
	return response, nil
}
