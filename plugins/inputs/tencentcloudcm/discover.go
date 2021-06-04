package tencentcloudcm

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

// DiscoverTool discovers objects for given regions
type DiscoverTool struct {
	DiscoveredInstances map[string]map[string]map[string][]*Instance
	DiscoveredMetrics   map[string][]string
	rw                  *sync.RWMutex

	Log telegraf.Logger `toml:"-"`
}

// NewDiscoverTool Factory
func NewDiscoverTool(log telegraf.Logger) *DiscoverTool {
	return &DiscoverTool{
		DiscoveredInstances: map[string]map[string]map[string][]*Instance{},
		rw:                  &sync.RWMutex{},
		Log:                 log,
	}
}

// DiscoverMetrics discovers metrics supported by registered products
func (d *DiscoverTool) DiscoverMetrics() {
	// discover metrics once
	for namespace, p := range Registry {
		if d.DiscoveredMetrics == nil {
			d.DiscoveredMetrics = map[string][]string{}
		}
		d.DiscoveredMetrics[namespace] = p.Metrics()
	}
}

func (d *DiscoverTool) discover(accounts []*Account, interval time.Duration, endpoint string) {
	for _, account := range accounts {
		for _, namespace := range account.Namespaces {
			for _, region := range namespace.Regions {
				// skip discover is specified
				if len(region.Instances) != 0 {
					continue
				}
				p, ok := Registry[namespace.Name]
				if !ok {
					continue
				}
				instances, err := p.Discover(account.Crs, region.RegionName, endpoint)
				if err != nil {
					d.Log.Errorf("discover account:%s region:%s endpoint:%s failed, error: %s",
						account.Name, region.RegionName, endpoint, err)
				}
				if d.DiscoveredInstances[account.Name] == nil {
					d.DiscoveredInstances[account.Name] = map[string]map[string][]*Instance{}
				}
				if d.DiscoveredInstances[account.Name][namespace.Name] == nil {
					d.DiscoveredInstances[account.Name][namespace.Name] = map[string][]*Instance{}
				}
				if d.DiscoveredInstances[account.Name][namespace.Name][region.RegionName] == nil {
					d.DiscoveredInstances[account.Name][namespace.Name][region.RegionName] = []*Instance{}
				}

				d.rw.Lock()
				d.DiscoveredInstances[account.Name][namespace.Name][region.RegionName] = instances
				d.rw.Unlock()
			}
		}
	}
}

// Discover discovers instances of registered products
func (d *DiscoverTool) Discover(accounts []*Account, interval config.Duration, endpoint string) {

	ticker := time.NewTicker(time.Duration(interval))

	d.discover(accounts, time.Duration(interval), endpoint)

	// discover instances periodically
	for range ticker.C {
		d.discover(accounts, time.Duration(interval), endpoint)
	}

}

// GetInstances Get discovered instances
func (d *DiscoverTool) GetInstances(account, namespace, region string) []*Instance {
	instances := []*Instance{}
	d.rw.RLock()
	v1, ok := d.DiscoveredInstances[account]
	if ok {
		v2, ok := v1[namespace]
		if ok {
			ins, ok := v2[region]
			if ok {
				instances = ins
			}
		}
	}
	d.rw.RUnlock()
	return instances
}

var Registry = map[string]Product{}

func Add(namespace string, p Product) {
	Registry[namespace] = p
}
