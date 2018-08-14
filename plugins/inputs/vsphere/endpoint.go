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
	collectMux      sync.RWMutex
	initialized     bool
	collectClient   *Client
	discoverClient  *Client
	wg              *ConcurrentWaitGroup
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
	getObjects       func(context.Context, *view.ContainerView) (objectMap, error)
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
		wg:           NewConcurrentWaitGroup(),
	}

	e.resourceKinds = map[string]resourceKind{
		"cluster": {
			name:             "cluster",
			pKey:             "clustername",
			enabled:          anythingEnabled(parent.ClusterMetricExclude),
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
			enabled:          anythingEnabled(parent.HostMetricExclude),
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
			enabled:          anythingEnabled(parent.VMMetricExclude),
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
			enabled:          anythingEnabled(parent.DatastoreMetricExclude),
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

func anythingEnabled(ex []string) bool {
	for _, s := range ex {
		if s == "*" {
			return false
		}
	}
	return true
}

func newFilterOrPanic(include []string, exclude []string) filter.Filter {
	f, err := filter.NewIncludeExcludeFilter(include, exclude)
	if err != nil {
		panic(fmt.Sprintf("Include/exclude filters are invalid: %s", err))
	}
	return f
}

func (e *Endpoint) startDiscovery(ctx context.Context) {
	e.discoveryTicker = time.NewTicker(e.Parent.ObjectDiscoveryInterval.Duration)
	go func() {
		for {
			select {
			case <-e.discoveryTicker.C:
				err := e.discover(ctx)
				if err != nil && err != context.Canceled {
					log.Printf("E! [input.vsphere]: Error in discovery for %s: %v", e.URL.Host, err)
				}
			case <-ctx.Done():
				log.Printf("D! [input.vsphere]: Exiting discovery goroutine for %s", e.URL.Host)
				e.discoveryTicker.Stop()
				return
			}
		}
	}()
}

func (e *Endpoint) init(ctx context.Context) error {

	err := e.setupMetricIds(ctx)
	if err != nil {
		return err
	}

	if e.Parent.ObjectDiscoveryInterval.Duration.Seconds() > 0 {

		// Run an initial discovery. If force_discovery_on_init isn't set, we kick it off as a
		// goroutine without waiting for it. This will probably cause us to report an empty
		// dataset on the first collection, but it solves the issue of the first collection timing out.
		if e.Parent.ForceDiscoverOnInit {
			log.Printf("D! [input.vsphere]: Running initial discovery and waiting for it to finish")
			err := e.discover(ctx)
			if err != nil {
				return err
			}
			e.startDiscovery(ctx)
		} else {
			// Otherwise, just run it in the background. We'll probably have an incomplete first metric
			// collection this way.
			go func() {
				err := e.discover(ctx)
				if err != nil && err != context.Canceled {
					log.Printf("E! [input.vsphere]: Error in discovery for %s: %v", e.URL.Host, err)
				}
				e.startDiscovery(ctx)
			}()
		}
	}
	e.initialized = true
	return nil
}

func (e *Endpoint) setupMetricIds(ctx context.Context) error {
	client, err := NewClient(e.URL, e.Parent)
	if err != nil {
		return err
	}
	defer client.Close()

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

func (e *Endpoint) getMetadata(ctx context.Context, in interface{}) interface{} {
	rq := in.(*metricQRequest)
	//log.Printf("D! [input.vsphere]: Querying metadata for %s", rq.obj.name)
	metrics, err := e.discoverClient.Perf.AvailableMetric(ctx, rq.obj.ref.Reference(), rq.res.sampling)
	if err != nil && err != context.Canceled {
		log.Printf("E! [input.vsphere]: Error while getting metric metadata. Discovery will be incomplete. Error: %s", err)
	}
	return &metricQResponse{metrics: &metrics, obj: rq.obj}
}

func (e *Endpoint) discover(ctx context.Context) error {
	// Add returning false means we've been released from Wait and no
	// more tasks are allowed. This happens when the plugin is stopped
	// or reloaded.
	if !e.wg.Add(1) {
		return context.Canceled
	}
	defer e.wg.Done()

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

	log.Printf("D! [input.vsphere]: Discover new objects for %s", e.URL.Host)

	instInfo := make(map[string]resourceInfo)
	resourceKinds := make(map[string]resourceKind)

	// Populate resource objects, and endpoint instance info.
	for k, res := range e.resourceKinds {
		// Need to do this for all resource types even if they are not enabled (but datastore)
		if res.enabled || (k != "datastore" && k != "vm") {
			objects, err := res.getObjects(ctx, e.discoverClient.Root)
			if err != nil {
				return err
			}

			// Set up a worker pool for processing metadata queries concurrently
			wp := NewWorkerPool(10)
			wp.Run(ctx, e.getMetadata, e.Parent.DiscoverConcurrency)

			// Fill the input channels with resources that need to be queried
			// for metadata.
			wp.Fill(ctx, func(ctx context.Context, f PushFunc) {
				for _, obj := range objects {
					f(ctx, &metricQRequest{obj: obj, res: &res})
				}
			})

			// Drain the resulting metadata and build instance infos.
			wp.Drain(ctx, func(ctx context.Context, in interface{}) {
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

func getClusters(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name", "parent"}, &resources)
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

func getHosts(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.HostSystem
	err := root.Retrieve(ctx, []string{"HostSystem"}, []string{"name", "parent"}, &resources)
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

func getVMs(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.VirtualMachine
	err := root.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "runtime.host", "config.guestId"}, &resources)
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

func getDatastores(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datastore
	err := root.Retrieve(ctx, []string{"Datastore"}, []string{"name", "parent"}, &resources)
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

func (e *Endpoint) collect(ctx context.Context, acc telegraf.Accumulator) error {
	// Add returning false means we've been released from Wait and no
	// more tasks are allowed. This happens when the plugin is stopped
	// or reloaded.
	if !e.wg.Add(1) {
		return context.Canceled
	}
	defer e.wg.Done()

	var err error
	if !e.initialized {
		err := e.init(ctx)
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
		err = e.discover(ctx)
		if err != nil {
			return err
		}
	}
	for k, res := range e.resourceKinds {
		if res.enabled {
			count, duration, err := e.collectResource(ctx, k, acc)
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

func (e *Endpoint) chunker(ctx context.Context, f PushFunc, res *resourceKind, now time.Time, latest time.Time) {
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.ObjectsPerQuery)
	metrics := 0
	total := 0
	nRes := 0
	for _, object := range res.objects {
		info, found := e.instanceInfo[object.ref.Value]
		if !found {
			log.Printf("E! [input.vsphere]: Internal error: Instance info not found for MOID %s", object.ref)
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
			if mr > 0 || (!res.realTime && metrics >= e.Parent.MetricsPerQuery) || nRes >= e.Parent.ObjectsPerQuery {
				log.Printf("D! [input.vsphere]: Querying %d objects, %d metrics (%d remaining) of type %s for %s. Processed objects: %d. Total objects %d",
					len(pqs), metrics, mr, res.name, e.URL.Host, total+1, len(res.objects))

				// To prevent deadlocks, don't send work items if the context has been cancelled.
				if ctx.Err() == context.Canceled {
					return
				}

				// Call push function
				f(ctx, pqs)
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
		// Call push function
		f(ctx, pqs)
	}
}

func (e *Endpoint) collectResource(ctx context.Context, resourceType string, acc telegraf.Accumulator) (int, float64, error) {

	// Do we have new data yet?
	res := e.resourceKinds[resourceType]
	now := time.Now()
	latest, hasLatest := e.lastColls[resourceType]
	if hasLatest {
		elapsed := time.Now().Sub(latest).Seconds() + 5.0 // Allow 5 second jitter.
		log.Printf("D! [input.vsphere]: Latest: %s, elapsed: %f, resource: %s", latest, elapsed, resourceType)
		if !res.realTime && elapsed < float64(res.sampling) {
			// No new data would be available. We're outta herE! [input.vsphere]:
			log.Printf("D! [input.vsphere]: Sampling period for %s of %d has not elapsed for %s",
				resourceType, res.sampling, e.URL.Host)
			return 0, 0, nil
		}
	} else {
		latest = time.Now().Add(time.Duration(-res.sampling) * time.Second)
	}

	internalTags := map[string]string{"resourcetype": resourceType}
	sw := NewStopwatchWithTags("endpoint_gather", e.URL.Host, internalTags)

	log.Printf("D! [input.vsphere]: [input.vsphere] Start of sample period deemed to be %s", latest)
	log.Printf("D! [input.vsphere]: Collecting metrics for %d objects of type %s for %s",
		len(res.objects), resourceType, e.URL.Host)

	count := int64(0)

	// Set up a worker pool for collecting chunk metrics
	wp := NewWorkerPool(10)
	wp.Run(ctx, func(ctx context.Context, in interface{}) interface{} {
		chunk := in.([]types.PerfQuerySpec)
		n, err := e.collectChunk(ctx, chunk, resourceType, res, acc)
		log.Printf("D! [input.vsphere]: Query returned %d metrics", n)
		if err != nil {
			return err
		}
		atomic.AddInt64(&count, int64(n))
		return nil

	}, 10)

	// Fill the input channel of the worker queue by running the chunking
	// logic implemented in chunker()
	wp.Fill(ctx, func(ctx context.Context, f PushFunc) {
		e.chunker(ctx, f, &res, now, latest)
	})

	// Drain the pool. We're getting errors back. They should all be nil
	var err error
	wp.Drain(ctx, func(ctx context.Context, in interface{}) {
		if in != nil {
			err = in.(error)
		}
	})

	e.lastColls[resourceType] = now // Use value captured at the beginning to avoid blind spots.

	sw.Stop()
	SendInternalCounterWithTags("endpoint_gather_count", e.URL.Host, internalTags, count)
	return int(count), time.Now().Sub(now).Seconds(), err
}

func (e *Endpoint) collectChunk(ctx context.Context, pqs []types.PerfQuerySpec, resourceType string,
	res resourceKind, acc telegraf.Accumulator) (int, error) {
	count := 0
	prefix := "vsphere" + e.Parent.Separator + resourceType

	metrics, err := e.collectClient.Perf.Query(ctx, pqs)
	if err != nil {
		return count, err
	}

	ems, err := e.collectClient.Perf.ToMetricSeries(ctx, metrics)
	if err != nil {
		return count, err
	}

	// Iterate through results
	for _, em := range ems {
		moid := em.Entity.Reference().Value
		instInfo, found := e.instanceInfo[moid]
		if !found {
			log.Printf("E! [input.vsphere]: MOID %s not found in cache. Skipping! (This should not happen!)", moid)
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
			objectRef, ok := res.objects[moid]
			if !ok {
				log.Printf("E! [input.vsphere]: MOID %s not found in cache. Skipping", moid)
				continue
			}
			e.populateTags(&objectRef, resourceType, &res, t, &v)

			// Now deal with the values
			for idx, value := range v.Value {
				ts := em.SampleInfo[idx].Timestamp

				// Organize the metrics into a bucket per measurement.
				// Data SHOULD be presented to us with the same timestamp for all samples, but in case
				// they don't we use the measurement name + timestamp as the key for the bucket.
				mn, fn := e.makeMetricIdentifier(prefix, name)
				bKey := mn + " " + v.Instance + " " + strconv.FormatInt(ts.UnixNano(), 10)
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: t}
					buckets[bKey] = bucket
				}
				if value < 0 {
					log.Printf("D! [input.vsphere]: Negative value for %s on %s. Indicates missing samples", name, objectRef.name)
					continue
				}
				bucket.fields[fn] = value
				count++
			}
		}
		// We've iterated through all the metrics and collected buckets for each
		// measurement name. Now emit them!
		for _, bucket := range buckets {
			//log.Printf("Bucket key: %s", key)
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	return count, nil
}

func (e *Endpoint) populateTags(objectRef *objectRef, resourceType string, resource *resourceKind, t map[string]string, v *performance.MetricSeries) {
	// Map name of object.
	if resource.pKey != "" {
		t[resource.pKey] = objectRef.name
	}

	// Map parent reference
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
