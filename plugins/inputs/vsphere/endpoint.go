package vsphere

import (
	"context"
	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Endpoint struct {
	Parent          *VSphere
	Url             *url.URL
	client          *Client
	lastColls       map[string]time.Time
	nameCache       map[string]string
	resources       map[string]resource
	discoveryTicker *time.Ticker
	clientMux       *sync.Mutex
	collectMux      *sync.RWMutex
	initialized     bool
}

type resource struct {
	enabled    bool
	realTime   bool
	sampling   int32
	objects    objectMap
	metricIds  []types.PerfMetricId
	wildcards  []string
	getObjects func(*view.ContainerView) (objectMap, error)
}

type objectMap map[string]objectRef

type objectRef struct {
	name      string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
	guest     string
}

func NewEndpoint(parent *VSphere, url *url.URL) *Endpoint {
	e := Endpoint{
		Url:         url,
		Parent:      parent,
		lastColls:   make(map[string]time.Time),
		nameCache:   make(map[string]string),
		clientMux:   &sync.Mutex{},
		collectMux:  &sync.RWMutex{},
		initialized: false,
	}

	e.resources = map[string]resource{
		"cluster": {
			enabled:    parent.GatherClusters,
			realTime:   false,
			sampling:   300,
			objects:    make(objectMap),
			wildcards:  parent.ClusterMetrics,
			getObjects: getClusters,
		},
		"host": {
			enabled:    parent.GatherHosts,
			realTime:   true,
			sampling:   20,
			objects:    make(objectMap),
			wildcards:  parent.HostMetrics,
			getObjects: getHosts,
		},
		"vm": {
			enabled:    parent.GatherVms,
			realTime:   true,
			sampling:   20,
			objects:    make(objectMap),
			wildcards:  parent.VmMetrics,
			getObjects: getVMs,
		},
		"datastore": {
			enabled:    parent.GatherDatastores,
			realTime:   false,
			sampling:   300,
			objects:    make(objectMap),
			wildcards:  parent.DatastoreMetrics,
			getObjects: getDatastores,
		},
	}

	return &e
}

func (e *Endpoint) init() error {

	err := e.setupMetricIds()
	if err != nil {
		log.Printf("E! Error in metric setup for %s: %v", e.Url.Host, err)
		return err
	}

	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() > 0 {
		// Run an initial discovery.
		//
		err = e.discover()
		if err != nil {
			log.Printf("E! Error in initial discovery for %s: %v", e.Url.Host, err)
			return err
		}

		// Create discovery ticker
		//
		e.discoveryTicker = time.NewTicker(e.Parent.ObjectDiscoveryInterval.Duration)
		go func() {
			for range e.discoveryTicker.C {
				err := e.discover()
				if err != nil {
					log.Printf("E! Error in discovery for %s: %v", e.Url.Host, err)
				}
			}
		}()
	}

	e.initialized = true
	return nil
}

func (e *Endpoint) setupMetricIds() error {
	client, err := e.getClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	metricMap, err := client.Perf.CounterInfoByName(ctx)
	if err != nil {
		return err
	}

	for k, res := range e.resources {
		if res.enabled {
			res.metricIds, err = resolveMetricWildcards(metricMap, res.wildcards)
			if err != nil {
				return err
			}
			e.resources[k] = res
		}
	}

	return nil
}

func resolveMetricWildcards(metricMap map[string]*types.PerfCounterInfo, wildcards []string) ([]types.PerfMetricId, error) {
	// Nothing specified assumes we're looking at everything
	//
	if wildcards == nil || len(wildcards) == 0 {
		return nil, nil
	}
	tmpMap := make(map[string]types.PerfMetricId)
	for _, pattern := range wildcards {
		exclude := false
		if pattern[0] == '!' {
			pattern = pattern[1:]
			exclude = true
		}
		p, err := glob.Compile(pattern)
		if err != nil {
			return nil, err
		}
		for name, info := range metricMap {
			if p.Match(name) {
				if exclude {
					delete(tmpMap, name)
					log.Printf("D! excluded %s", name)
				} else {
					tmpMap[name] = types.PerfMetricId{CounterId: info.Key}
					log.Printf("D! included %s", name)
				}
			}
		}
	}
	result := make([]types.PerfMetricId, len(tmpMap))
	idx := 0
	for _, id := range tmpMap {
		result[idx] = id
		idx++
	}
	return result, nil
}

func (e *Endpoint) discover() error {
	log.Printf("D! Discover new objects for %s", e.Url.Host)

	client, err := e.getClient()
	if err != nil {
		return err
	}

	nameCache := make(map[string]string)
	resources := e.resources

	// Populate resource objects, and endpoint name cache
	//
	for k, res := range resources {
		// Need to do this for all resource types even if they are not enabled (but datastore)
		//
		if res.enabled || k != "datastore" {
			objects, err := res.getObjects(client.Root)
			if err != nil {
				e.checkClient()
				return err
			}

			for _, obj := range objects {
				nameCache[obj.ref.Value] = obj.name
			}
			res.objects = objects
			resources[k] = res
		}
	}

	// Atomically swap maps
	//
	if e.collectMux == nil {
		e.collectMux = &sync.RWMutex{}
	}
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	e.nameCache = nameCache
	e.resources = resources

	log.Printf("D! Discovered %d objects for %s", len(nameCache), e.Url.Host)

	return nil
}

func getClusters(root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(context.Background(), []string{"ClusterComputeResource"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func getHosts(root *view.ContainerView) (objectMap, error) {
	var resources []mo.HostSystem
	err := root.Retrieve(context.Background(), []string{"HostSystem"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func getVMs(root *view.ContainerView) (objectMap, error) {
	var resources []mo.VirtualMachine
	err := root.Retrieve(context.Background(), []string{"VirtualMachine"}, []string{"name", "runtime.host", "config.guestId"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		var guest string
		// Sometimes Config is unknown and returns a nil pointer
		//
		if r.Config != nil {
			guest = cleanGuestId(r.Config.GuestId)
		} else {
			guest = "unknown"
		}
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Runtime.Host, guest: guest}
	}
	return m, nil
}

func getDatastores(root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datastore
	err := root.Retrieve(context.Background(), []string{"Datastore"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.Summary.Name] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func (e *Endpoint) collect(acc telegraf.Accumulator) error {

	var err error
	if !e.initialized {
		err := e.init()
		if err != nil {
			return err
		}
	}

	// If discovery interval is disabled (0), discover on each collection cycle
	//
	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() == 0 {
		err = e.discover()
		if err != nil {
			log.Printf("E! Error in discovery prior to collect for %s: %v", e.Url.Host, err)
			return err
		}
	}

	for k, res := range e.resources {
		if res.enabled {
			count, duration, err := e.collectResource(k, acc)
			if err != nil {
				return err
			}

			acc.AddGauge("vsphere",
				map[string]interface{}{"gather.count": count, "gather.duration": duration},
				map[string]string{"vcenter": e.Url.Host, "type": k},
				time.Now())
		}
	}

	return nil
}

func (e *Endpoint) collectResource(resourceType string, acc telegraf.Accumulator) (int, float64, error) {

	// Do we have new data yet?
	//
	res := e.resources[resourceType]
	now := time.Now()
	latest, hasLatest := e.lastColls[resourceType]
	if hasLatest {
		elapsed := time.Now().Sub(latest).Seconds()
		if elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			//
			log.Printf("D! Sampling period for %s of %d has not elapsed for %s", resourceType, res.sampling, e.Url.Host)
			return 0, 0, nil
		}
	}

	if !hasLatest {
		latest = now.Add(-time.Duration(res.sampling) * time.Second)
		e.lastColls[resourceType] = latest
	}

	if len(res.metricIds) == 0 {
		log.Printf("D! Collecting all metrics for %d objects of type %s for %s", len(res.objects), resourceType, e.Url.Host)
	} else {
		log.Printf("D! Collecting %d metrics for %d objects of type %s for %s", len(res.metricIds), len(res.objects), resourceType, e.Url.Host)
	}

	client, err := e.getClient()
	if err != nil {
		return 0, 0, err
	}
	ctx := context.Background()

	// Object maps may change, so we need to hold the collect lock
	//
	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	measurementName := "vsphere." + resourceType
	count := 0
	start := time.Now()
	total := 0
	lastTS := latest
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
	for _, object := range res.objects {
		pq := types.PerfQuerySpec{
			Entity:     object.ref,
			MaxSample:  1,
			MetricId:   res.metricIds,
			IntervalId: res.sampling,
		}

		if !res.realTime {
			pq.StartTime = &latest
			pq.EndTime = &now
		}

		pqs = append(pqs, pq)
		total++

		// Filled up a chunk or at end of data? Run a query with the collected objects
		//
		if len(pqs) >= int(e.Parent.ObjectsPerQuery) || total == len(res.objects) {
			log.Printf("D! Querying %d objects of type %s for %s. Object count: %d. Total objects %d", len(pqs), resourceType, e.Url.Host, total, len(res.objects))
			metrics, err := client.Perf.Query(ctx, pqs)
			if err != nil {
				//TODO: Check the error and attempt to handle gracefully. (ie: object no longer exists)
				log.Printf("E! Error querying metrics of %s for %s", resourceType, e.Url.Host)
				e.checkClient()
				return count, time.Now().Sub(start).Seconds(), err
			}

			ems, err := client.Perf.ToMetricSeries(ctx, metrics)
			if err != nil {
				e.checkClient()
				return count, time.Now().Sub(start).Seconds(), err
			}

			// Iterate through result and fields list
			//
			for _, em := range ems {
				moid := em.Entity.Reference().Value
				for _, v := range em.Value {
					name := v.Name
					for idx, value := range v.Value {

						objectName := e.nameCache[moid]
						t := map[string]string{
							"vcenter":  e.Url.Host,
							"hostname": objectName,
							"moid":     moid,
						}

						objectRef, ok := res.objects[moid]
						if ok {
							parent := e.nameCache[objectRef.parentRef.Value]
							switch resourceType {
							case "host":
								t["cluster"] = parent
								break

							case "vm":
								t["guest"] = objectRef.guest
								t["esxhost"] = parent
								hostRes := e.resources["host"]
								hostRef, ok := hostRes.objects[objectRef.parentRef.Value]
								if ok {
									cluster, ok := e.nameCache[hostRef.parentRef.Value]
									if ok {
										t["cluster"] = cluster
									}
								}
								break
							}
						}

						if v.Instance != "" {
							if strings.HasPrefix(name, "cpu.") {
								t["cpu"] = v.Instance
							} else if strings.HasPrefix(name, "datastore.") {
								t["datastore"] = v.Instance
							} else if strings.HasPrefix(name, "disk.") {
								t["disk"] = cleanDiskTag(v.Instance)
							} else if strings.HasPrefix(name, "net.") {
								t["interface"] = v.Instance
							} else if strings.HasPrefix(name, "storageAdapter.") {
								t["adapter"] = v.Instance
							} else if strings.HasPrefix(name, "storagePath.") {
								t["path"] = v.Instance
							} else if strings.HasPrefix(name, "sys.resource") {
								t["resource"] = v.Instance
							} else if strings.HasPrefix(name, "vflashModule.") {
								t["module"] = v.Instance
							} else if strings.HasPrefix(name, "virtualDisk.") {
								t["disk"] = v.Instance
							} else {
								// default to instance
								t["instance"] = v.Instance
							}
						}

						ts := em.SampleInfo[idx].Timestamp
						if ts.After(lastTS) {
							lastTS = ts
						}

						f := map[string]interface{}{name: value}
						acc.AddFields(measurementName, f, t, ts)
						count++
					}
				}
			}
			pqs = make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
		}
	}

	if count > 0 {
		e.lastColls[resourceType] = lastTS
	}

	log.Printf("D! Collection of %s for %s, took %v returning %d metrics", resourceType, e.Url.Host, time.Now().Sub(start), count)
	return count, time.Now().Sub(start).Seconds(), nil
}

func cleanGuestId(id string) string {
	if strings.HasSuffix(id, "Guest") {
		return id[:len(id)-5]
	}

	return id
}

func cleanDiskTag(disk string) string {
	if strings.HasPrefix(disk, "<") {
		i := strings.Index(disk, ">")
		if i > -1 {
			s1 := disk[1:i]
			s2 := disk[i+1:]
			if s1 == s2 {
				return s1
			}
		}
	}

	return disk
}

func (e *Endpoint) getClient() (*Client, error) {
	if e.client == nil {
		e.clientMux.Lock()
		defer e.clientMux.Unlock()
		if e.client == nil {
			log.Printf("D! Creating new vCenter client for: %s", e.Url.Host)
			client, err := NewClient(e.Url, e.Parent)
			if err != nil {
				return nil, err
			}
			e.client = client
		}
	}
	return e.client, nil
}

func (e *Endpoint) checkClient() {
	if e.client != nil {
		active, err := e.client.Client.SessionManager.SessionIsActive(context.Background())
		if err != nil {
			log.Printf("E! SessionIsActive returned an error on %s: %v", e.Url.Host, err)
			e.client = nil
		}
		if !active {
			log.Printf("I! Session no longer active, reseting client: %s", e.Url.Host)
			e.client = nil
		}
	}
}
