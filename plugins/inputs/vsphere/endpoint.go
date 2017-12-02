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
	Parent             *VSphere
	Url                *url.URL
	client             *Client
	lastColls          map[string]time.Time
	nameCache          map[string]string
	resources          map[string]resource
	discoveryTicker    *time.Ticker
	clientMux          *sync.Mutex
	collectMux         *sync.RWMutex
	initialized        bool
}

type resource struct {
	enabled bool
	realTime bool
	sampling int32
	objects objectMap
	metricIds []types.PerfMetricId
	wildcards []string
}

type objectMap map[string]objectRef

type objectRef struct {
	name      string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
}

func NewEndpoint(parent *VSphere, url *url.URL) *Endpoint {
	e := Endpoint{
		Url:          url,
		Parent:       parent,
		lastColls:    make(map[string]time.Time),
		nameCache:    make(map[string]string),
		clientMux:    &sync.Mutex{},
		collectMux:   &sync.RWMutex{},
		initialized:  false,
	}

	e.resources = map[string]resource{
		"cluster": {enabled: parent.GatherClusters, realTime: false, sampling: 300, objects: make(objectMap), wildcards: parent.ClusterMetrics},
		"host": {enabled: parent.GatherHosts, realTime: true, sampling: 20, objects: make(objectMap), wildcards: parent.HostMetrics},
		"vm": {enabled: parent.GatherVms, realTime: true, sampling: 20, objects: make(objectMap), wildcards: parent.VmMetrics},
		"datastore": {enabled: parent.GatherDatastores, realTime: false, sampling: 300, objects: make(objectMap), wildcards: parent.DatastoreMetrics},
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
	conn, err := e.getConnection()
	if err != nil {
		return err
	}
	ctx := context.Background()

	metricMap, err := conn.Perf.CounterInfoByName(ctx)
	if err != nil {
		return err
	}

	for _, res := range e.resources {
		res.metricIds, err = resolveMetricWildcards(metricMap, res.wildcards)
		if err != nil {
			return err
		}
	}

	return nil
}

func resolveMetricWildcards(metricMap map[string]*types.PerfCounterInfo, wildcards []string) ([]types.PerfMetricId, error) {
	// Nothing specified assumes we're looking at everything
	//
	if wildcards == nil {
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

	conn, err := e.getConnection()
	if err != nil {
		return err
	}

	nameCache := make(map[string]string)
	resources := e.resources

	for k, res := range resources {
		if res.enabled {
			var objects objectMap
			switch k {
			case "cluster":
				objects, err = e.getClusters(conn.Root)
			case "host":
				objects, err = e.getHosts(conn.Root)
			case "vm":
				objects, err = e.getVMs(conn.Root)
			case "datastore":
				objects, err = e.getDatastores(conn.Root)
			}
			if err != nil {
				return err
			}

			for _, obj := range res.objects {
				nameCache[obj.ref.Reference().Value] = obj.name
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

	log.Printf("D! Discovered %d objects\n", len(e.nameCache))

	return nil
}

func (e *Endpoint) getClusters(root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(context.Background(), []string{"ClusterComputeResource"}, []string{"summary", "name", "parent"}, &resources)
	if err != nil {
		e.checkConnection()
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func (e *Endpoint) getHosts(root *view.ContainerView) (objectMap, error) {
	var resources []mo.HostSystem
	err := root.Retrieve(context.Background(), []string{"HostSystem"}, []string{"summary", "parent"}, &resources)
	if err != nil {
		e.checkConnection()
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
	}
	return m, nil
}

func (e *Endpoint) getVMs(root *view.ContainerView) (objectMap, error) {
	var resources []mo.VirtualMachine
	err := root.Retrieve(context.Background(), []string{"VirtualMachine"}, []string{"summary", "runtime.host"}, &resources)
	if err != nil {
		e.checkConnection()
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Summary.Config.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Runtime.Host}
	}
	return m, nil
}

func (e *Endpoint) getDatastores(root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datastore
	err := root.Retrieve(context.Background(), []string{"Datastore"}, []string{"summary"}, &resources)
	if err != nil {
		e.checkConnection()
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.Summary.Name] = objectRef{
			name: r.Summary.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent}
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

	sampling := e.resources[resourceType].sampling
	realTime := e.resources[resourceType].realTime

	conn, err := e.getConnection()
	if err != nil {
		return 0, 0, err
	}
	ctx := context.Background()

	// Object maps may change, so we need to hold the collect lock
	//
	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	// Interval = 0 means collection for this metric was diabled, so don't even bother.
	log.Printf("D! Resource type: %s, sampling period is: %d", resourceType, sampling)

	// Do we have new data yet?
	//
	now := time.Now()
	latest, hasLatest := e.lastColls[resourceType]
	if hasLatest {
		elapsed := time.Now().Sub(latest).Seconds()
		if elapsed < float64(sampling) {
			// No new data would be available. We're outta here!
			//
			return 0, 0, nil
		}
	}
	e.lastColls[resourceType] = now

	objects := e.resources[resourceType].objects
	log.Printf("D! Collecting data metrics for %d objects of type %s for %s", len(objects), resourceType, e.Url.Host)

	measurementName := "vsphere." + resourceType
	count := 0
	start := time.Now()
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
	total := 0
	for _, object := range objects {
		pq := types.PerfQuerySpec{
			Entity:     object.ref,
			MaxSample:  1,
			MetricId:   e.resources[resourceType].metricIds,
			IntervalId: sampling,
		}

		if !realTime {
			startTime := now.Add(-time.Duration(sampling) * time.Second)
			pq.StartTime = &startTime
			pq.EndTime = &now
		}

		pqs = append(pqs, pq)
		total++

		// Filled up a chunk or at end of data? Run a query with the collected objects
		//
		if len(pqs) >= int(e.Parent.ObjectsPerQuery) || total == len(objects) {
			log.Printf("D! Querying %d objects of type %s for %s. Total processed: %d. Total objects %d\n", len(pqs), resourceType, e.Url.Host, total, len(objects))
			metrics, err := conn.Perf.Query(ctx, pqs)
			if err != nil {
				log.Printf("E! Error querying metrics for %s on %s", resourceType, e.Url.Host)
				e.checkConnection()
				return count, time.Now().Sub(start).Seconds(), err
			}

			ems, err := conn.Perf.ToMetricSeries(ctx, metrics)
			if err != nil {
				e.checkConnection()
				return count, time.Now().Sub(start).Seconds(), err
			}

			// Iterate through result and fields list
			//
			for _, em := range ems {
				moid := em.Entity.Reference().Value
				for _, v := range em.Value {
					name := v.Name
					for idx, value := range v.Value {
						f := map[string]interface{}{name: value}
						objectName := e.nameCache[moid]
						parent := ""
						parentRef := objects[moid].parentRef
						if parentRef != nil {
							parent = e.nameCache[parentRef.Value]
						}

						t := map[string]string{
							"vcenter":  e.Url.Host,
							"hostname": objectName,
							"moid":     moid,
							"parent":   parent,
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

						acc.AddFields(measurementName, f, t, em.SampleInfo[idx].Timestamp)
						count++
					}
				}
			}
			pqs = make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
		}
	}

	log.Printf("D! Collection of %s took %v\n", resourceType, time.Now().Sub(start))
	return count, time.Now().Sub(start).Seconds(), nil
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

func (e *Endpoint) getConnection() (*Client, error) {
	if e.client == nil {
		e.clientMux.Lock()
		defer e.clientMux.Unlock()
		if e.client == nil {
			log.Printf("D! Creating new vCenter client for: %s\n", e.Url.Host)
			conn, err := NewClient(e.Url, e.Parent.Timeout)
			if err != nil {
				return nil, err
			}
			e.client = conn
		}
	}
	return e.client, nil
}

func (e *Endpoint) checkConnection() {
	if e.client != nil {
		active, err := e.client.Client.SessionManager.SessionIsActive(context.Background())
		if !active || err != nil {
			log.Printf("I! vCenter session no longer active, reseting client: %s", e.Url.Host)
			e.client = nil
		}
	}
}

