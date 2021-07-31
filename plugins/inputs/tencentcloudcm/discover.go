package tencentcloudcm

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

// DiscoverTool discovers objects for given regions
type DiscoverTool struct {
	DiscoveredInstances map[string]map[string]map[string][]*monitor.Instance
	DiscoveredMetrics   map[string][]string
	rw                  *sync.RWMutex

	registry map[string]Product

	Log telegraf.Logger `toml:"-"`
}

// NewDiscoverTool Factory
func NewDiscoverTool(log telegraf.Logger) *DiscoverTool {
	discoverTool := &DiscoverTool{
		DiscoveredInstances: map[string]map[string]map[string][]*monitor.Instance{},
		rw:                  &sync.RWMutex{},
		registry:            map[string]Product{},
		Log:                 log,
	}
	discoverTool.Add("QCE/CVM", &CVM{})
	discoverTool.Add("QCE/CDB", &CDB{})
	discoverTool.Add("QCE/REDIS", &Redis{})
	discoverTool.Add("QCE/LB_PUBLIC", &LBPublic{})
	discoverTool.Add("QCE/LB_PRIVATE", &LBPrivate{})
	discoverTool.Add("QCE/CES", &CES{})
	return discoverTool
}

// DiscoverMetrics discovers metrics supported by registered products
func (d *DiscoverTool) DiscoverMetrics() {
	// discover metrics once
	for namespace, p := range d.registry {
		if d.DiscoveredMetrics == nil {
			d.DiscoveredMetrics = map[string][]string{}
		}
		d.DiscoveredMetrics[namespace] = p.Metrics()
	}
}

func (d *DiscoverTool) discover(accounts []*Account, endpoint string) error {
	discoveries := map[string]map[string]map[string][]*monitor.Instance{}
	for _, account := range accounts {
		for _, namespace := range account.Namespaces {
			for _, region := range namespace.Regions {
				// skip discover is specified
				if len(region.Instances) != 0 {
					continue
				}
				p, ok := d.registry[namespace.Name]
				if !ok {
					continue
				}
				instances, err := p.Discover(account.crs, region.RegionName, endpoint)
				if err != nil {
					d.Log.Errorf("discover account:%s region:%s endpoint:%s failed, error: %s",
						account.Name, region.RegionName, endpoint, err)
					return err
				}
				if discoveries[account.Name] == nil {
					discoveries[account.Name] = map[string]map[string][]*monitor.Instance{}
				}
				if discoveries[account.Name][namespace.Name] == nil {
					discoveries[account.Name][namespace.Name] = map[string][]*monitor.Instance{}
				}
				if discoveries[account.Name][namespace.Name][region.RegionName] == nil {
					discoveries[account.Name][namespace.Name][region.RegionName] = []*monitor.Instance{}
				}
				discoveries[account.Name][namespace.Name][region.RegionName] = instances
			}
		}
	}
	d.rw.Lock()
	d.DiscoveredInstances = discoveries
	d.rw.Unlock()
	return nil
}

// Discover discovers instances of registered products
func (d *DiscoverTool) Discover(accounts []*Account, interval config.Duration, endpoint string) {
	ticker := time.NewTicker(time.Duration(interval))
	defer ticker.Stop()

	err := d.discover(accounts, endpoint)
	if err != nil {
		d.Log.Errorf(err.Error())
	}

	// discover instances periodically
	for range ticker.C {
		err = d.discover(accounts, endpoint)
		if err != nil {
			d.Log.Errorf(err.Error())
		}
	}
}

// GetInstances Get discovered instances
func (d *DiscoverTool) GetInstances(account, namespace, region string) []*monitor.Instance {
	d.rw.RLock()
	v1, ok := d.DiscoveredInstances[account]
	if !ok {
		return nil
	}
	v2, ok := v1[namespace]
	if !ok {
		return nil
	}
	instances, ok := v2[region]
	d.rw.RUnlock()
	return instances
}

// var Registry = map[string]Product{}

func (d *DiscoverTool) Add(namespace string, p Product) {
	d.registry[namespace] = p
}
