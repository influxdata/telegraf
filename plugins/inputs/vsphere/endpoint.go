package vsphere

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf/filter"

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
	lastColls       map[string]time.Time
	instanceInfo    map[string]resourceInfo
	resourceKinds   map[string]resourceKind
	metricNames     map[int32]string
	discoveryTicker *time.Ticker
	clientMux       sync.Mutex
	collectMux      sync.RWMutex
	initialized     bool
	collectClient   *Client
	discoverClient  *Client
}

type resourceKind struct {
	name             string
	pKey             string
	enabled          bool
	realTime         bool
	sampling         int32
	objects          objectMap
	filters          filter.Filter
	collectInstances bool
	getObjects       func(*view.ContainerView) (objectMap, error)
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

type metricQRequest struct {
	res *resourceKind
	obj objectRef
}

type metricQResponse struct {
	obj     objectRef
	metrics *performance.MetricList
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

	e.resourceKinds = map[string]resourceKind{
		"cluster": {
			name:             "cluster",
			pKey:             "clustername",
			enabled:          parent.GatherClusters,
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.ClusterMetricInclude, parent.ClusterMetricExclude),
			collectInstances: parent.ClusterInstances,
			getObjects:       getClusters,
		},
		"host": {
			name:             "host",
			pKey:             "esxhostname",
			enabled:          parent.GatherHosts,
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.HostMetricInclude, parent.HostMetricExclude),
			collectInstances: parent.HostInstances,
			getObjects:       getHosts,
		},
		"vm": {
			name:             "vm",
			pKey:             "vmname",
			enabled:          parent.GatherVms,
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.VMMetricInclude, parent.VMMetricExclude),
			collectInstances: parent.VMInstances,
			getObjects:       getVMs,
		},
		"datastore": {
			name:             "datastore",
			pKey:             "dsname",
			enabled:          parent.GatherDatastores,
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatastoreMetricInclude, parent.DatastoreMetricExclude),
			collectInstances: parent.DatastoreInstances,
			getObjects:       getDatastores,
		},
	}

	return &e
}

func newFilterOrPanic(include []string, exclude []string) filter.Filter {
	f, err := filter.NewIncludeExcludeFilter(include, exclude)
	if err != nil {
		panic(fmt.Sprintf("Include/exclude filters are invalid: %s", err))
	}
	return f
}

func (e *Endpoint) init() error {

	err := e.setupMetricIds()
	if err != nil {
		log.Printf("E! Error in metric setup for %s: %v", e.URL.Host, err)
		return err
	}

	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() > 0 {
		discoverFunc := func() {
			err = e.discover()
			if err != nil {
				log.Printf("E! Error in initial discovery for %s: %v", e.URL.Host, err)
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

		// Run an initial discovery. If force_discovery_on_init isn't set, we kick it off as a
		// goroutine without waiting for it. This will probably cause us to report an empty
		// dataset on the first collection, but it solves the issue of the first collection timing out.
		//
		if e.Parent.ForceDiscoverOnInit {
			log.Printf("D! Running initial discovery and waiting for it to finish")
			discoverFunc()
		}
	}
	e.initialized = true
	return nil
}

func (e *Endpoint) setupMetricIds() error {
	client, err := NewClient(e.URL, e.Parent)
	if err != nil {
		return err
	}
	defer client.Close()

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

func (e *Endpoint) getMetadata(in interface{}) interface{} {
	rq := in.(*metricQRequest)
	ctx := context.Background()
	//log.Printf("D! Querying metadata for %s", rq.obj.name)
	metrics, err := e.discoverClient.Perf.AvailableMetric(ctx, rq.obj.ref.Reference(), rq.res.sampling)
	if err != nil {
		log.Printf("E! Error while getting metric metadata. Discovery will be incomplete. Error: %s", err)
	}
	return &metricQResponse{metrics: &metrics, obj: rq.obj}
}

func (e *Endpoint) discover() error {
	sw := NewStopwatch("discover", e.URL.Host)
	var err error
	e.discoverClient, err = NewClient(e.URL, e.Parent)
	if err != nil {
		return err
	}
	defer func() {
		e.discoverClient.Close()
		e.discoverClient = nil
	}()

	log.Printf("D! Discover new objects for %s", e.URL.Host)

	instInfo := make(map[string]resourceInfo)
	resourceKinds := make(map[string]resourceKind)

	// Populate resource objects, and endpoint instance info.
	//
	for k, res := range e.resourceKinds {
		// Need to do this for all resource types even if they are not enabled (but datastore)
		if res.enabled || (k != "datastore" && k != "vm") {
			objects, err := res.getObjects(e.discoverClient.Root)
			if err != nil {
				return err
			}

			// Set up a worker pool for processing metadata queries concurrently
			wp := NewWorkerPool(10)
			wp.Run(e.getMetadata, e.Parent.DiscoverConcurrency)

			// Fill the input channels with resources that need to be queried
			// for metadata.
			wp.Fill(func(in chan interface{}) {
				for _, obj := range objects {
					in <- &metricQRequest{obj: obj, res: &res}
				}
			})

			// Drain the resulting metadata and build instance infos.
			wp.Drain(func(in interface{}) {
				resp := in.(*metricQResponse)
				mList := make(performance.MetricList, 0)
				if res.enabled {
					for _, m := range *resp.metrics {
						if m.Instance != "" && !res.collectInstances {
							continue
						}
						if res.filters.Match(e.metricNames[m.CounterId]) {
							mList = append(mList, m)
						}
					}
				}
				instInfo[resp.obj.ref.Value] = resourceInfo{name: resp.obj.name, metrics: mList}
			})
			res.objects = objects
			resourceKinds[k] = res
		}
	}

	// Atomically swap maps
	//
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	e.instanceInfo = instInfo
	e.resourceKinds = resourceKinds

	sw.Stop()
	SendInternalCounter("discovered_objects", e.URL.Host, int64(len(instInfo)))
	return nil
}

func getClusters(root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(context.Background(), []string{"ClusterComputeResource"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap, len(resources))
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

	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	e.collectClient, err = NewClient(e.URL, e.Parent)
	if err != nil {
		return err
	}
	defer func() {
		e.collectClient.Close()
		e.collectClient = nil
	}()

	// If discovery interval is disabled (0), discover on each collection cycle
	//
	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() == 0 {
		err = e.discover()
		if err != nil {
			log.Printf("E! Error in discovery prior to collect for %s: %v", e.URL.Host, err)
			return err
		}
	}
	for k, res := range e.resourceKinds {
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

func (e *Endpoint) chunker(in chan interface{}, res *resourceKind, now time.Time, latest time.Time) {
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
	metrics := 0
	total := 0
	nRes := 0
	for _, object := range res.objects {
		info, found := e.instanceInfo[object.ref.Value]
		if !found {
			log.Printf("E! Internal error: Instance info not found for MOID %s", object.ref)
		}
		mr := len(info.metrics)
		for mr > 0 {
			mc := mr
			headroom := e.Parent.MetricsPerQuery - metrics
			if !res.realTime && mc > headroom { // Metric query limit only applies to non-realtime metrics
				mc = headroom
			}
			fm := len(info.metrics) - mr
			pq := types.PerfQuerySpec{
				Entity:     object.ref,
				MaxSample:  1,
				MetricId:   info.metrics[fm : fm+mc],
				IntervalId: res.sampling,
			}

			if !res.realTime {
				pq.StartTime = &latest
				pq.EndTime = &now
			}
			pqs = append(pqs, pq)
			mr -= mc
			metrics += mc

			// We need to dump the current chunk of metrics for one of two reasons:
			// 1) We filled up the metric quota while processing the current resource
			// 2) We are at the last resource and have no more data to process.
			//
			if mr > 0 || (!res.realTime && metrics >= e.Parent.MetricsPerQuery) || nRes >= e.Parent.ObjectsPerQuery {
				log.Printf("D! Querying %d objects, %d metrics (%d remaining) of type %s for %s. Processed objects: %d. Total objects %d",
					len(pqs), metrics, mr, res.name, e.URL.Host, total+1, len(res.objects))
				in <- pqs
				pqs = make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
				metrics = 0
				nRes = 0
			}
		}
		total++
		nRes++
	}
	// There may be dangling stuff in the queue. Handle them
	//
	if len(pqs) > 0 {
		in <- pqs
	}
}

func (e *Endpoint) collectResource(resourceType string, acc telegraf.Accumulator) (int, float64, error) {
	// Do we have new data yet?
	//
	res := e.resourceKinds[resourceType]
	now := time.Now()
	latest, hasLatest := e.lastColls[resourceType]
	if hasLatest {
		elapsed := time.Now().Sub(latest).Seconds() + 5.0 // Allow 5 second jitter.
		log.Printf("D! Latest: %s, elapsed: %f, resource: %s", latest, elapsed, resourceType)
		if elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			//
			log.Printf("D! Sampling period for %s of %d has not elapsed for %s",
				resourceType, res.sampling, e.URL.Host)
			return 0, 0, nil
		}
	} else {
		latest = time.Now().Add(time.Duration(-res.sampling) * time.Second)
	}

	internalTags := map[string]string{"resourcetype": resourceType}
	sw := NewStopwatchWithTags("endpoint_gather", e.URL.Host, internalTags)

	log.Printf("D! Start of sample period deemed to be %s", latest)
	log.Printf("D! Collecting metrics for %d objects of type %s for %s",
		len(res.objects), resourceType, e.URL.Host)

	count := int64(0)
	//	chunkCh := make(chan []types.PerfQuerySpec)
	//	errorCh := make(chan error, e.Parent.CollectConcurrency) // Try not to block on errors.
	//	doneCh := make(chan bool)

	// Set up a worker pool for collecting chunk metrics
	wp := NewWorkerPool(10)
	wp.Run(func(in interface{}) interface{} {
		chunk := in.([]types.PerfQuerySpec)
		n, err := e.collectChunk(chunk, resourceType, res, acc)
		log.Printf("D! Query returned %d metrics", n)
		if err != nil {
			return err
		}
		atomic.AddInt64(&count, int64(n))
		return nil

	}, 10)

	// Fill the input channel of the worker queue by running the chunking
	// logic implemented in chunker()
	wp.Fill(func(in chan interface{}) {
		e.chunker(in, &res, now, latest)
	})

	// Drain the pool. We're getting errors back. They should all be nil
	var err error
	wp.Drain(func(in interface{}) {
		if in != nil {
			err = in.(error)
		}
	})

	e.lastColls[resourceType] = now // Use value captured at the beginning to avoid blind spots.

	sw.Stop()
	SendInternalCounterWithTags("endpoint_gather_count", e.URL.Host, internalTags, count)
	return int(count), time.Now().Sub(now).Seconds(), err
}

func (e *Endpoint) collectChunk(pqs []types.PerfQuerySpec, resourceType string,
	res resourceKind, acc telegraf.Accumulator) (int, error) {
	count := 0
	prefix := "vsphere" + e.Parent.Separator + resourceType

	ctx := context.Background()
	metrics, err := e.collectClient.Perf.Query(ctx, pqs)
	if err != nil {
		//TODO: Check the error and attempt to handle gracefully. (ie: object no longer exists)
		log.Printf("E! Error querying metrics of %s for %s %s", resourceType, e.URL.Host, err)
		return count, err
	}

	ems, err := e.collectClient.Perf.ToMetricSeries(ctx, metrics)
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
			e.populateTags(&objectRef, resourceType, &res, t, &v)

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
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: t}
					buckets[bKey] = bucket
				}
				if value < 0 {
					log.Printf("D! Negative value for %s on %s. Indicates missing samples", name, objectRef.name)
					continue
				}
				bucket.fields[fn] = value
				count++
			}
		}
		// We've iterated through all the metrics and collected buckets for each
		// measurement name. Now emit them!
		//
		for _, bucket := range buckets {
			//log.Printf("Bucket key: %s", key)
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	return count, nil
}

func (e *Endpoint) populateTags(objectRef *objectRef, resourceType string, resource *resourceKind, t map[string]string, v *performance.MetricSeries) {
	// Map name of object.
	//
	if resource.pKey != "" {
		t[resource.pKey] = objectRef.name
	}

	// Map parent reference
	//
	parent, found := e.instanceInfo[objectRef.parentRef.Value]
	if found {
		switch resourceType {
		case "host":
			t["clustername"] = parent.name
			break

		case "vm":
			t["guest"] = objectRef.guest
			t["esxhostname"] = parent.name
			hostRes := e.resourceKinds["host"]
			hostRef, ok := hostRes.objects[objectRef.parentRef.Value]
			if ok {
				cluster, ok := e.instanceInfo[hostRef.parentRef.Value]
				if ok {
					t["clustername"] = cluster.name
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
