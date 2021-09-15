package tencentcloudcm

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"

	"github.com/influxdata/telegraf/plugins/inputs"
)

// TencentCloudCM contains the configuration for the Tencent Cloud Cloud Monitor plugin.
type TencentCloudCM struct {
	Endpoint string `toml:"endpoint"`

	Period config.Duration `toml:"period"`
	Delay  config.Duration `toml:"delay"`

	RateLimit int             `toml:"ratelimit"`
	Timeout   config.Duration `toml:"timeout"`
	BatchSize int             `toml:"batch_size"`

	DiscoveryInterval config.Duration `toml:"discovery_interval"`

	Accounts []*Account `toml:"accounts"`

	client cmClient

	Log telegraf.Logger `toml:"-"`

	windowStart  time.Time
	windowEnd    time.Time
	discoverTool *discoverTool
}

// Account defines a Tencent Cloud account
type Account struct {
	Name       string       `toml:"name"`
	SecretID   string       `toml:"secret_id"`
	SecretKey  string       `toml:"secret_key"`
	Namespaces []*Namespace `toml:"namespaces"`

	crs *common.Credential
}

// Namespace defines a Tencent Cloud CM namespace
type Namespace struct {
	Name    string    `toml:"namespace"`
	Metrics []string  `toml:"metrics"`
	Regions []*Region `toml:"regions"`
}

// Region defines a Tencent Cloud region
type Region struct {
	RegionName       string      `toml:"region"`
	Instances        []*Instance `toml:"instances"`
	monitorInstances []*monitor.Instance
}

func (r *Region) instancesToMonitor() {
	if r.monitorInstances == nil {
		r.monitorInstances = []*monitor.Instance{}
	}
	for _, instance := range r.Instances {
		monitorInstance := &monitor.Instance{}
		for _, dimension := range instance.Dimensions {
			for k, v := range dimension {
				monitorInstance.Dimensions = append(monitorInstance.Dimensions, &monitor.Dimension{
					Name:  common.StringPtr(k),
					Value: common.StringPtr(v),
				})
			}
		}
		r.monitorInstances = append(r.monitorInstances, monitorInstance)
	}
}

type Instance struct {
	Dimensions []map[string]string `toml:"dimensions"`
}

// SampleConfig implements telegraf.Input interface
func (t *TencentCloudCM) SampleConfig() string {
	return `
	## Endpoint to make request against, the correct endpoint is automatically
	## determined and this option should only be set if you wish to override the
	## default.
	##   ex: endpoint = "tencentcloudapi.com"
	# endpoint = ""
  
	## The default period for Tencent Cloud Cloud Monitor metrics is 1 minute (60s). However not all
	## metrics are made available to the 1 minute period. Some are collected at
	## 5 minute, 60 minute, or larger intervals.
	## See: https://intl.cloud.tencent.com/document/product/248/33882
	## Note that if a period is configured that is smaller than the default for a
	## particular metric, that metric will not be returned by the Tencent Cloud API
	## and will not be collected by Telegraf.
	##
	## Requested Tencent Cloud Cloud Monitor aggregation Period (required - must be a multiple of 60s)
	## period = "5m"
  
	## Collection Delay (must account for metrics availability via Tencent Cloud API)
	# delay = "0m"
  
	## Maximum requests per second. Note that the global default Tencent Cloud API rate limit is
	## 20 calls/second (1,200 calls/minute), so if you define multiple namespaces, these should add up to a
	## maximum of 20.
	## See https://intl.cloud.tencent.com/document/product/248/33881
	# ratelimit = 20
  
	## Timeout for http requests made by the Tencent Cloud client.
	# timeout = "5s"
  
	## Batch instance size for intiating a GetMonitorData API call.
	# batch_size = 100
	## By default, Tencent Cloud CM Input plugin will automatically discover instances in specified regions
	## This sets the interval for discover and update the instances discovered.
	##
	## how often the discovery API call executed (default 1m)
	# discovery_interval = "5m"
  
	## Tencent Cloud Account (required - you can provide multiple entries and distinguish them using
	## optional name field, if name is empty, index number will be used as default)
	[[inputs.tencentcloudcm.accounts]]
	  name = ""
	  secret_id = ""
	  secret_key = ""
  
	  ## Namespaces to Pull
	  [[inputs.tencentcloudcm.accounts.namespaces]]
		## Tencent Cloud CM Namespace (required - see https://intl.cloud.tencent.com/document/product/248/34716#namespace)
		namespace = "QCE/CVM"
  
		## Metrics filter, all metrics will be pulled if not left empty. Different namespaces may have different
		## metric names, e.g. CVM Monitoring Metrics: https://intl.cloud.tencent.com/document/product/248/6843
		# metrics = ["CPUUsage", "MemUsage"]
  
		[[inputs.tencentcloudcm.accounts.namespaces.regions]]
		  ## Tencent Cloud regions (required - Allowed values: https://intl.cloud.tencent.com/document/api/248/33876)
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
		  ## - QCE/DC
		  # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances]]
		  # [[inputs.tencentcloudcm.accounts.namespaces.regions.instances.dimensions]]
		  #   name = "value"
`
}

// Description implements telegraf.Input interface
func (t *TencentCloudCM) Description() string {
	return "Pull Metric Statistics from Tencent Cloud Cloud Monitor"
}

// Init is for setup, and validating config.
func (t *TencentCloudCM) Init() error {
	if t.Period <= 0 {
		t.Period = config.Duration(5 * time.Minute)
	}
	if t.Delay <= 0 {
		t.Delay = config.Duration(0 * time.Minute)
	}
	if t.BatchSize <= 0 {
		t.BatchSize = 100
	}

	if len(t.Accounts) == 0 {
		return fmt.Errorf("account is empty")
	}

	t.discoverTool = NewDiscoverTool(t.Log)

	// create account credential
	for i := range t.Accounts {
		if t.Accounts[i].SecretID == "" {
			return fmt.Errorf("secret_id is empty")
		}
		if t.Accounts[i].SecretKey == "" {
			return fmt.Errorf("secret_key is empty")
		}

		t.Accounts[i].crs = common.NewCredential(
			t.Accounts[i].SecretID,
			t.Accounts[i].SecretKey,
		)
		if t.Accounts[i].Name == "" {
			t.Accounts[i].Name = fmt.Sprintf("account_%v", i)
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
				_, ok := t.discoverTool.registry[namespace.Name]
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
	go t.discoverTool.Discover(t.Accounts, t.DiscoveryInterval, t.Endpoint)
	t.discoverTool.DiscoverMetrics()
	for i, account := range t.Accounts {
		for j, namespace := range account.Namespaces {
			// use discovered metrics if not specified
			if len(namespace.Metrics) == 0 {
				metrics, ok := t.discoverTool.DiscoveredMetrics[namespace.Name]
				if !ok {
					return fmt.Errorf("unsupported namespace %s for discovering metrics, please specify metrics",
						namespace.Name)
				}
				t.Accounts[i].Namespaces[j].Metrics = metrics
			}
		}
	}

	t.client = &cloudmonitorClient{Accounts: t.Accounts, Log: t.Log}

	return nil

}

func (t *TencentCloudCM) updateWindow(relativeTo time.Time) {
	windowEnd := relativeTo.Add(-time.Duration(t.Delay))
	// starting point is two times the aggregation period to make sure all points are covered
	t.windowStart = windowEnd.Add(-time.Duration(t.Period) * 2)
	t.windowEnd = windowEnd
}

func (t *TencentCloudCM) Gather(acc telegraf.Accumulator) error {
	t.updateWindow(time.Now())

	metricObjects := t.client.GetMetricObjects(*t)

	lmtr := limiter.NewRateLimiter(t.RateLimit, time.Second)
	defer lmtr.Stop()

	wg := sync.WaitGroup{}
	rLock := sync.Mutex{}
	results := []monitor.GetMonitorDataResponse{}

	// requestIDMap contains request ID and metric objects for later aggregation
	requestIDMap := map[string]metricObject{}
	for _, obj := range metricObjects {
		wg.Add(1)
		<-lmtr.C
		go func(m metricObject) {
			defer wg.Done()

			for {

				client, err := t.client.NewClient(m.Region, m.Account.crs, *t)
				if err != nil {
					acc.AddError(err)
					return
				}

				if len(m.MonitorInstances) >= t.BatchSize {
					batch := m.MonitorInstances[:t.BatchSize]
					if len(batch) == 0 {
						break
					}

					request := t.client.NewGetMonitorDataRequest(m.Namespace, m.Metric, batch, *t)

					result, err := t.client.GatherMetrics(client, request, *t)
					if err != nil {
						acc.AddError(err)
						break
					}

					rLock.Lock()
					requestIDMap[*result.Response.RequestId] = m
					results = append(results, *result)
					rLock.Unlock()

					m.MonitorInstances = m.MonitorInstances[t.BatchSize:]
				} else {
					request := t.client.NewGetMonitorDataRequest(m.Namespace, m.Metric, m.MonitorInstances, *t)

					result, err := t.client.GatherMetrics(client, request, *t)
					if err != nil {
						acc.AddError(err)
						break
					}

					rLock.Lock()
					requestIDMap[*result.Response.RequestId] = m
					results = append(results, *result)
					rLock.Unlock()
					break
				}

			}

		}(obj)

	}
	wg.Wait()
	for _, result := range results {
		for _, datapoints := range result.Response.DataPoints {
			tags := map[string]string{}

			keys := []string{}
			for _, v := range datapoints.Dimensions {
				tags[*v.Name] = *v.Value
				keys = append(keys, *v.Value)
			}

			metricObject := requestIDMap[*result.Response.RequestId]

			if metricObject.isDiscovered {
				instance := t.discoverTool.GetInstance(metricObject.Account.Name, metricObject.Namespace, metricObject.Region, newKey(strings.Join(keys, "-")))
				for k, v := range instance {
					tags[fmt.Sprintf("_%s_%s", metricObject.Namespace, k)] = fmt.Sprintf("%v", v)
				}
			}

			tags["account"] = metricObject.Account.Name
			tags["namespace"] = metricObject.Namespace
			tags["region"] = metricObject.Region
			tags["period"] = fmt.Sprintf("%v", *result.Response.Period)
			tags["metric"] = *result.Response.MetricName
			tags["request_id"] = *result.Response.RequestId

			measurement := fmt.Sprintf("tencentcloudcm_%s", metricObject.Namespace)

			for index, value := range datapoints.Values {
				acc.AddFields(
					measurement,
					map[string]interface{}{*result.Response.MetricName: value},
					tags,
					time.Unix(int64(*datapoints.Timestamps[index]), 0),
				)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("tencentcloudcm", func() telegraf.Input {
		return &TencentCloudCM{
			Endpoint:          "tencentcloudapi.com",
			RateLimit:         20,
			BatchSize:         100,
			Timeout:           config.Duration(5 * time.Second),
			DiscoveryInterval: config.Duration(5 * time.Minute),
		}
	})
}
