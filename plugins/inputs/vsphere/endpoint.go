package vsphere

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// Endpoint is a high-level representation of a connected vCenter endpoint. It is backed by the lower
// level Client type.
type Endpoint struct {
	Parent          *VSphere
	URL             *url.URL
	client          *Client
	lastColls       map[string]time.Time
	instanceInfo    map[string]resourceInfo
	resources       map[string]resource
	metricNames     map[int32]string
	discoveryTicker *time.Ticker
	clientMux       sync.Mutex
	collectMux      sync.RWMutex
	initialized     bool
}

type resource struct {
	enabled    bool
	realTime   bool
	sampling   int32
	objects    objectMap
	includes   []string
	excludes   []string
	getObjects func(*view.ContainerView) (objectMap, error)
}

type metricEntry struct {
	tags   map[string]string
	name   string
	ts     time.Time
	fields map[string]interface{}
}

type objectMap map[string]objectRef

type objectRef struct {
	name      string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
	guest     string
}

type resourceInfo struct {
	name    string
	metrics performance.MetricList
}

// NewEndpoint returns a new connection to a vCenter based on the URL and configuration passed
// as parameters.
func NewEndpoint(parent *VSphere, url *url.URL) *Endpoint {
	e := Endpoint{
		URL:          url,
		Parent:       parent,
		lastColls:    make(map[string]time.Time),
		instanceInfo: make(map[string]resourceInfo),
		initialized:  false,
	}

	e.resources = map[string]resource{
		"cluster": {
			enabled:    parent.GatherClusters,
			realTime:   false,
			sampling:   300,
			objects:    make(objectMap),
			includes:   parent.ClusterMetricInclude,
			excludes:   parent.ClusterMetricExclude,
			getObjects: getClusters,
		},
		"host": {
			enabled:    parent.GatherHosts,
			realTime:   true,
			sampling:   20,
			objects:    make(objectMap),
			includes:   parent.HostMetricInclude,
			excludes:   parent.HostMetricExclude,
			getObjects: getHosts,
		},
		"vm": {
			enabled:    parent.GatherVms,
			realTime:   true,
			sampling:   20,
			objects:    make(objectMap),
			includes:   parent.VmMetricInclude,
			excludes:   parent.VmMetricExclude,
			getObjects: getVMs,
		},
		"datastore": {
			enabled:    parent.GatherDatastores,
			realTime:   false,
			sampling:   300,
			objects:    make(objectMap),
			includes:   parent.DatastoreMetricInclude,
			excludes:   parent.DatastoreMetricExclude,
			getObjects: getDatastores,
		},
	}

	return &e
}

func (e *Endpoint) init() error {

	err := e.setupMetricIds()
	if err != nil {
		log.Printf("E! Error in metric setup for %s: %v", e.URL.Host, err)
		return err
	}

	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() > 0 {
		// Run an initial discovery.
		//
		err = e.discover()
		if err != nil {
			log.Printf("E! Error in initial discovery for %s: %v", e.URL.Host, err)
			return err
		}

		// Create discovery ticker
		//
		e.discoveryTicker = time.NewTicker(e.Parent.ObjectDiscoveryInterval.Duration)
		go func() {
			for range e.discoveryTicker.C {
				err := e.discover()
				if err != nil {
					log.Printf("E! Error in discovery for %s: %v", e.URL.Host, err)
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
	mn, err := client.Perf.CounterInfoByName(ctx)
	if err != nil {
		return err
	}
	e.metricNames = make(map[int32]string)
	for name, m := range mn {
		e.metricNames[m.Key] = name
	}
	return nil
}

func (e *Endpoint) discover() error {
	start := time.Now()
	log.Printf("D! Discover new objects for %s", e.URL.Host)

	client, err := e.getClient()
	if err != nil {
		return err
	}

	instInfo := make(map[string]resourceInfo)
	resources := e.resources

	// Populate resource objects, and endpoint name cache
	//
	for k, res := range resources {
		// Don't be tempted to skip disabled resources here! We may need them to resolve parent references
		//
		// Precompile includes and excludes
		//
		cInc := make([]glob.Glob, len(res.includes))
		for i, p := range res.includes {
			cInc[i] = glob.MustCompile(p)
		}
		cExc := make([]glob.Glob, len(res.excludes))
		for i, p := range res.excludes {
			cExc[i] = glob.MustCompile(p)
		}
		// Need to do this for all resource types even if they are not enabled (but datastore)
		//
		if res.enabled || k != "datastore" {
			objects, err := res.getObjects(client.Root)
			if err != nil {
				e.checkClient()
				return err
			}
			for _, obj := range objects {
				ctx := context.Background()
				mList := make(performance.MetricList, 0)
				metrics, err := e.client.Perf.AvailableMetric(ctx, obj.ref.Reference(), res.sampling)
				if err != nil {
					e.checkClient()
					return err
				}
				log.Printf("D! Obj: %s, metrics found: %d, enabled: %t", obj.name, len(metrics), res.enabled)

				// Mmetric metadata gathering is only needed for enabled resource types.
				//
				if res.enabled {
					for _, m := range metrics {
						include := len(cInc) == 0 // Empty include list means include all
						mName := e.metricNames[m.CounterId]
						//log.Printf("% %s", mName, cInc)
						if !include { // If not included by default
							for _, p := range cInc {
								if p.Match(mName) {
									include = true
								}
							}
						}
						if include {
							for _, p := range cExc {
								if p.Match(mName) {
									include = false
									log.Printf("D! Excluded: %s", mName)
									break
								}
							}
							if include { // If still included after processing excludes
								mList = append(mList, m)
								log.Printf("D! Included %s Sampling: %d", mName, res.sampling)
							}
						}

					}
				}
				instInfo[obj.ref.Value] = resourceInfo{name: obj.name, metrics: mList}
			}
			res.objects = objects
			resources[k] = res
		}
	}

	// Atomically swap maps
	//
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	e.instanceInfo = instInfo
	e.resources = resources

	log.Printf("D! Discovered %d objects for %s. Took %s", len(instInfo), e.URL.Host, time.Now().Sub(start))

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
			guest = cleanGuestID(r.Config.GuestId)
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
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
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
			log.Printf("E! Error in discovery prior to collect for %s: %v", e.URL.Host, err)
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
				map[string]string{"vcenter": e.URL.Host, "type": k},
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
		log.Printf("D! Latest: %s, elapsed: %f, resource: %s", latest, elapsed, resourceType)
		if elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			//
			log.Printf("D! Sampling period for %s of %d has not elapsed for %s", resourceType, res.sampling, e.URL.Host)
			return 0, 0, nil
		}
	} else {
		latest = time.Now().Add(time.Duration(-res.sampling) * time.Second)
	}
	log.Printf("Start of sample period deemed to be %s", latest)

	log.Printf("D! Collecting metrics for %d objects of type %s for %s", len(res.objects), resourceType, e.URL.Host)

	// Object maps may change, so we need to hold the collect lock in read mode
	//
	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	count := 0
	start := time.Now()
	total := 0
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
	for _, object := range res.objects {
		info, found := e.instanceInfo[object.ref.Value]
		if !found {
			log.Printf("E! Internal error: Instance info not found for MOID %s", object.ref)
		}
		if len(info.metrics) > 0 {

			pq := types.PerfQuerySpec{
				Entity:     object.ref,
				MaxSample:  1,
				MetricId:   info.metrics,
				IntervalId: res.sampling,
			}

			if !res.realTime {
				pq.StartTime = &latest
				pq.EndTime = &now
			}

			pqs = append(pqs, pq)
		} else {
			log.Printf("D! No metrics available for %s. Skipping.", info.name)
			// Maintainers: Don't be tempted to skip a turn in the loop here! We still need to check if
			// the chunk needs processing or we risk skipping the last chunk. (Ask me how I found this out... :) )
		}
		total++

		// Filled up a chunk or at end of data? Run a query with the collected objects
		//
		if len(pqs) > 0 && (len(pqs) >= int(e.Parent.ObjectsPerQuery) || total >= len(res.objects)) {
			log.Printf("D! Querying %d objects of type %s for %s. Object count: %d. Total objects %d", len(pqs), resourceType, e.URL.Host, total, len(res.objects))
			n, err := e.collectChunk(pqs, resourceType, res, acc)
			if err != nil {
				e.checkClient()
				return count, time.Now().Sub(start).Seconds(), err
			}
			count += n
			pqs = make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
		}
	}
	e.lastColls[resourceType] = now // Use value captured at the beginning to avoid blind spots.
	log.Printf("D! Collection of %s for %s, took %v returning %d metrics", resourceType, e.URL.Host, time.Now().Sub(start), count)
	return count, time.Now().Sub(start).Seconds(), nil
}

func (e *Endpoint) collectChunk(pqs []types.PerfQuerySpec, resourceType string, res resource, acc telegraf.Accumulator) (int, error) {
	count := 0
	prefix := "vsphere" + e.Parent.Separator + resourceType
	client, err := e.getClient()
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	metrics, err := client.Perf.Query(ctx, pqs)
	if err != nil {
		//TODO: Check the error and attempt to handle gracefully. (ie: object no longer exists)
		log.Printf("E! Error querying metrics of %s for %s", resourceType, e.URL.Host)
		return count, err
	}

	ems, err := client.Perf.ToMetricSeries(ctx, metrics)
	if err != nil {
		return count, err
	}

	// Iterate through results
	//
	for _, em := range ems {
		moid := em.Entity.Reference().Value
		instInfo, found := e.instanceInfo[moid]
		if !found {
			log.Printf("E! MOID %s not found in cache. Skipping! (This should not happen!)", moid)
			continue
		}
		buckets := make(map[string]metricEntry)
		for _, v := range em.Value {
			name := v.Name
			t := map[string]string{
				"vcenter":  e.URL.Host,
				"hostname": instInfo.name,
				"moid":     moid,
			}
			// Populate tags
			//
			objectRef, ok := res.objects[moid]
			if !ok {
				log.Printf("E! MOID %s not found in cache. Skipping", moid)
				continue
			}
			e.populateTags(objectRef, resourceType, t, &v)

			// Now deal with the values
			//
			for idx, value := range v.Value {
				ts := em.SampleInfo[idx].Timestamp

				// Organize the metrics into a bucket per measurement.
				// Data SHOULD be presented to us with the same timestamp for all samples, but in case
				// they don't we use the measurement name + timestamp as the key for the bucket.
				//
				mn, fn := e.makeMetricIdentifier(prefix, name)
				bKey := mn + " " + v.Instance + " " + strconv.FormatInt(ts.UnixNano(), 10)
				//log.Printf("D! Bucket key: %s, resource: %s, field: %s", bKey, instInfo.name, fn)
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: t}
					buckets[bKey] = bucket
				}
				bucket.fields[fn] = value
				count++
			}
		}
		// We've iterated through all the metrics and collected buckets for each
		// measurement name. Now emit them!
		//
		log.Printf("D! Collected %d buckets for %s", len(buckets), instInfo.name)
		for key, bucket := range buckets {
			log.Printf("D! Key: %s, Tags: %s, Fields: %s", key, bucket.tags, bucket.fields)
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	return count, nil
}

func (e *Endpoint) populateTags(objectRef objectRef, resourceType string, t map[string]string, v *performance.MetricSeries) {
	// Map name of object. For vms and hosts, we use the default "hostname".
	//
	switch resourceType {
	case "datastore":
		t["dsname"] = objectRef.name
	case "disk":
		t["diskname"] = objectRef.name
	case "cluster":
		t["clustername"] = objectRef.name
	}

	// Map parent reference
	//
	parent, found := e.instanceInfo[objectRef.parentRef.Value]
	if found {
		switch resourceType {
		case "host":
			t["cluster"] = parent.name
			break

		case "vm":
			t["guest"] = objectRef.guest
			t["esxhost"] = parent.name
			hostRes := e.resources["host"]
			hostRef, ok := hostRes.objects[objectRef.parentRef.Value]
			if ok {
				cluster, ok := e.instanceInfo[hostRef.parentRef.Value]
				if ok {
					t["cluster"] = cluster.name
				}
			}
			break
		}
	}

	// Determine which point tag to map to the instance
	//
	name := v.Name
	if v.Instance != "" {
		if strings.HasPrefix(name, "cpu.") {
			t["cpu"] = v.Instance
		} else if strings.HasPrefix(name, "datastore.") {
			t["lun"] = v.Instance
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
}

func (e *Endpoint) makeMetricIdentifier(prefix, metric string) (string, string) {
	parts := strings.Split(metric, ".")
	if len(parts) == 1 {
		return prefix, parts[0]
	}
	return prefix + e.Parent.Separator + parts[0], strings.Join(parts[1:], e.Parent.Separator)
}

func cleanGuestID(id string) string {
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
			log.Printf("D! Creating new vCenter client for: %s", e.URL.Host)
			client, err := NewClient(e.URL, e.Parent)
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
			log.Printf("E! SessionIsActive returned an error on %s: %v", e.URL.Host, err)
			e.client = nil
		}
		if !active {
			log.Printf("I! Session no longer active, reseting client: %s", e.URL.Host)
			e.client = nil
		}
	}
}
