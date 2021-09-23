package tencentcloudcm

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

const limit = 100

// discoverTool discovers objects for given regions
type discoverTool struct {
	DiscoveredObjects map[string]discoverObject
	DiscoveredMetrics map[string][]string
	rw                *sync.RWMutex

	registry map[string]Product

	Log telegraf.Logger
}

type discoverObject struct {
	Instances        map[string]map[string]string // Discovered Instances with detailed instance information
	MonitorInstances []*monitor.Instance          // Monitor Instances with information enough for metrics pulling
}

// NewDiscoverTool Factory
func NewDiscoverTool(log telegraf.Logger) *discoverTool {
	discoverTool := &discoverTool{
		DiscoveredObjects: map[string]discoverObject{},
		DiscoveredMetrics: map[string][]string{},
		rw:                &sync.RWMutex{},
		registry:          map[string]Product{},
		Log:               log,
	}
	discoverTool.registry = map[string]Product{
		"QCE/CVM":        &CVM{},
		"QCE/CDB":        &CDB{},
		"QCE/REDIS":      &Redis{},
		"QCE/LB_PUBLIC":  &LBPublic{},
		"QCE/LB_PRIVATE": &LBPrivate{},
		"QCE/CES":        &CES{},
		"QCE/DC":         &DC{},
	}
	return discoverTool
}

// DiscoverMetrics discovers metrics supported by registered products
func (d *discoverTool) DiscoverMetrics() {
	// discover metrics once
	for namespace, p := range d.registry {
		d.DiscoveredMetrics[namespace] = p.Metrics()
	}
}

func (d *discoverTool) discoverObjects(accounts []*Account, endpoint string) {
	discoveredObjects := map[string]discoverObject{}
	for _, account := range accounts {
		for _, namespace := range account.Namespaces {
			for _, region := range namespace.Regions {
				// skip discover if instances are explicitly specified
				if len(region.monitorInstances) != 0 {
					continue
				}
				p, ok := d.registry[namespace.Name]
				if !ok {
					d.Log.Debugf("discover registration for namespace: %s not found", namespace.Name)
					continue
				}
				discoveredObject, err := discover(account.crs, region.RegionName, endpoint, p, d.Log)
				if err != nil {
					d.Log.Errorf("discover account:%s region:%s endpoint:%s failed, error: %s",
						account.Name, region.RegionName, endpoint, err)
					continue
				}
				discoveredObjects[newKey(account.Name, namespace.Name, region.RegionName)] = discoveredObject
			}
		}
	}
	d.rw.Lock()
	d.DiscoveredObjects = discoveredObjects
	d.rw.Unlock()
}

// Discover discovers instances of registered products
func (d *discoverTool) Discover(accounts []*Account, interval config.Duration, endpoint string) {
	ticker := time.NewTicker(time.Duration(interval))
	defer ticker.Stop()

	d.discoverObjects(accounts, endpoint)

	// discover instances periodically
	for range ticker.C {
		d.discoverObjects(accounts, endpoint)
	}
}

// GetInstance Get discovered instance detail
func (d *discoverTool) GetInstance(account, namespace, region, key string) map[string]string {
	d.rw.RLock()
	defer d.rw.RUnlock()
	discoverObject, ok := d.DiscoveredObjects[newKey(account, namespace, region)]
	if !ok {
		return nil
	}
	return discoverObject.Instances[key]
}

// GetMonitorInstances Get discovered monitor instances
func (d *discoverTool) GetMonitorInstances(account, namespace, region string) []*monitor.Instance {
	d.rw.RLock()
	defer d.rw.RUnlock()
	discoverObject, ok := d.DiscoveredObjects[newKey(account, namespace, region)]
	if !ok {
		return nil
	}
	return discoverObject.MonitorInstances
}

func discover(crs *common.Credential, region, endpoint string, p Product, l telegraf.Logger) (discoverObject, error) {
	discoverObject := discoverObject{
		Instances: map[string]map[string]string{},
	}

	var offset int64
	instances := []map[string]string{}

	total, instancesByPage, err := p.Discover(crs, region, endpoint, offset, limit)
	if err != nil {
		return discoverObject, err
	}
	instances = append(instances, instancesByPage...)
	offset += limit

	// discover all instances if total count is bigger than limit
	for i := int64(1); i < int64(total-1)/limit+1; i++ {
		_, instancesByPage, err := p.Discover(crs, region, endpoint, offset, limit)
		if err != nil {
			return discoverObject, err
		}
		instances = append(instances, instancesByPage...)
		offset += limit
	}

	monitorInstances := []*monitor.Instance{}
	for _, instance := range instances {
		keyIsNil := false
		keyVals := []string{}
		dimensions := []*monitor.Dimension{}

		for instanceKey, dimensionKey := range p.Keys() {
			// check if discovered key field is nil
			value := instance[instanceKey]
			if value == "" {
				keyIsNil = true
				l.Debugf("instance key: %s has empty value", instanceKey)
				break
			}

			keyVals = append(keyVals, value)
			dimensions = append(dimensions, &monitor.Dimension{
				Name:  common.StringPtr(dimensionKey),
				Value: common.StringPtr(value),
			})
		}

		// instance with nil key field will be discarded
		if keyIsNil {
			continue
		}

		discoverObject.Instances[newKey(keyVals...)] = instance
		monitorInstances = append(monitorInstances, &monitor.Instance{Dimensions: dimensions})
	}
	discoverObject.MonitorInstances = monitorInstances

	return discoverObject, nil
}

func newKey(vals ...string) string {
	sort.Strings(vals)
	vs := []string{}
	for _, v := range vals {
		vs = append(vs, fmt.Sprintf("%v", v))
	}
	return strings.Join(vs, "-")
}
