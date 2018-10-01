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
	"github.com/vmware/govmomi/object"
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
	discoveryTicker *time.Ticker
	collectMux      sync.RWMutex
	initialized     bool
	clientFactory   *ClientFactory
	busy            sync.Mutex
}

type resourceKind struct {
	name             string
	pKey             string
	parentTag        string
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
	altID     string
	ref       types.ManagedObjectReference
	parentRef *types.ManagedObjectReference //Pointer because it must be nillable
	guest     string
	dcname    string
}

type resourceInfo struct {
	name      string
	metrics   performance.MetricList
	parentRef *types.ManagedObjectReference
}

type metricQRequest struct {
	res *resourceKind
	obj objectRef
}

type metricQResponse struct {
	obj     objectRef
	metrics *performance.MetricList
}

type multiError []error

// NewEndpoint returns a new connection to a vCenter based on the URL and configuration passed
// as parameters.
func NewEndpoint(ctx context.Context, parent *VSphere, url *url.URL) (*Endpoint, error) {
	e := Endpoint{
		URL:           url,
		Parent:        parent,
		lastColls:     make(map[string]time.Time),
		instanceInfo:  make(map[string]resourceInfo),
		initialized:   false,
		clientFactory: NewClientFactory(ctx, url, parent),
	}

	e.resourceKinds = map[string]resourceKind{
		"datacenter": {
			name:             "datacenter",
			pKey:             "dcname",
			parentTag:        "",
			enabled:          anythingEnabled(parent.DatacenterMetricExclude),
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatacenterMetricInclude, parent.DatacenterMetricExclude),
			collectInstances: parent.DatacenterInstances,
			getObjects:       getDatacenters,
		},
		"cluster": {
			name:             "cluster",
			pKey:             "clustername",
			parentTag:        "dcname",
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
			parentTag:        "clustername",
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
			parentTag:        "esxhostname",
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

	// Start discover and other goodness
	err := e.init(ctx)

	return &e, err
}

func (m multiError) Error() string {
	switch len(m) {
	case 0:
		return "No error recorded. Something is wrong!"
	case 1:
		return m[0].Error()
	default:
		s := "Multiple errors detected concurrently: "
		for i, e := range m {
			if i != 0 {
				s += ", "
			}
			s += e.Error()
		}
		return s
	}
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

func (e *Endpoint) initalDiscovery(ctx context.Context) {
	err := e.discover(ctx)
	if err != nil && err != context.Canceled {
		log.Printf("E! [input.vsphere]: Error in discovery for %s: %v", e.URL.Host, err)
	}
	e.startDiscovery(ctx)
}

func (e *Endpoint) init(ctx context.Context) error {

	if e.Parent.ObjectDiscoveryInterval.Duration > 0 {

		// Run an initial discovery. If force_discovery_on_init isn't set, we kick it off as a
		// goroutine without waiting for it. This will probably cause us to report an empty
		// dataset on the first collection, but it solves the issue of the first collection timing out.
		if e.Parent.ForceDiscoverOnInit {
			log.Printf("D! [input.vsphere]: Running initial discovery and waiting for it to finish")
			e.initalDiscovery(ctx)
		} else {
			// Otherwise, just run it in the background. We'll probably have an incomplete first metric
			// collection this way.
			go e.initalDiscovery(ctx)
		}
	}
	e.initialized = true
	return nil
}

func (e *Endpoint) getMetricNameMap(ctx context.Context) (map[int32]string, error) {
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	mn, err := client.Perf.CounterInfoByName(ctx)

	if err != nil {
		return nil, err
	}
	names := make(map[int32]string)
	for name, m := range mn {
		names[m.Key] = name
	}
	return names, nil
}

func (e *Endpoint) getMetadata(ctx context.Context, in interface{}) interface{} {
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	rq := in.(*metricQRequest)
	metrics, err := client.Perf.AvailableMetric(ctx, rq.obj.ref.Reference(), rq.res.sampling)
	if err != nil && err != context.Canceled {
		log.Printf("E! [input.vsphere]: Error while getting metric metadata. Discovery will be incomplete. Error: %s", err)
	}
	return &metricQResponse{metrics: &metrics, obj: rq.obj}
}

func (e *Endpoint) getDatacenterName(ctx context.Context, client *Client, cache map[string]string, r types.ManagedObjectReference) string {
	path := make([]string, 0)
	returnVal := ""
	here := r
	for {
		if name, ok := cache[here.Reference().String()]; ok {
			// Populate cache for the entire chain of objects leading here.
			returnVal = name
			break
		}
		path = append(path, here.Reference().String())
		o := object.NewCommon(client.Client.Client, r)
		var result mo.ManagedEntity
		err := o.Properties(ctx, here, []string{"parent", "name"}, &result)
		if err != nil {
			log.Printf("W! [input.vsphere]: Error while resolving parent. Assuming no parent exists. Error: %s", err)
			break
		}
		if result.Reference().Type == "Datacenter" {
			// Populate cache for the entire chain of objects leading here.
			returnVal = result.Name
			break
		}
		if result.Parent == nil {
			log.Printf("D! [input.vsphere]: No parent found for %s (ascending from %s)", here.Reference(), r.Reference())
			break
		}
		here = result.Parent.Reference()
	}
	for _, s := range path {
		cache[s] = returnVal
	}
	return returnVal
}

func (e *Endpoint) discover(ctx context.Context) error {
	e.busy.Lock()
	defer e.busy.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	metricNames, err := e.getMetricNameMap(ctx)
	if err != nil {
		return err
	}

	sw := NewStopwatch("discover", e.URL.Host)

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	log.Printf("D! [input.vsphere]: Discover new objects for %s", e.URL.Host)

	instInfo := make(map[string]resourceInfo)
	resourceKinds := make(map[string]resourceKind)
	dcNameCache := make(map[string]string)

	// Populate resource objects, and endpoint instance info.
	for k, res := range e.resourceKinds {
		log.Printf("D! [input.vsphere] Discovering resources for %s", res.name)
		// Need to do this for all resource types even if they are not enabled (but datastore)
		if res.enabled || (k != "datastore" && k != "vm") {
			objects, err := res.getObjects(ctx, client.Root)
			if err != nil {
				return err
			}

			// Fill in datacenter names where available (no need to do it for Datacenters)
			if res.name != "Datacenter" {
				for k, obj := range objects {
					if obj.parentRef != nil {
						obj.dcname = e.getDatacenterName(ctx, client, dcNameCache, *obj.parentRef)
						objects[k] = obj
					}
				}
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
			wp.Drain(ctx, func(ctx context.Context, in interface{}) bool {
				switch resp := in.(type) {
				case *metricQResponse:
					mList := make(performance.MetricList, 0)
					if res.enabled {
						for _, m := range *resp.metrics {
							if m.Instance != "" && !res.collectInstances {
								continue
							}
							if res.filters.Match(metricNames[m.CounterId]) {
								mList = append(mList, m)
							}
						}
					}
					instInfo[resp.obj.ref.Value] = resourceInfo{name: resp.obj.name, metrics: mList, parentRef: resp.obj.parentRef}
				case error:
					log.Printf("W! [input.vsphere]: Error while discovering resources: %s", resp)
					return false
				}
				return true
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

func getDatacenters(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.Datacenter
	err := root.Retrieve(ctx, []string{"Datacenter"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap, len(resources))
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent, dcname: r.Name}
	}
	return m, nil
}

func getClusters(ctx context.Context, root *view.ContainerView) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	err := root.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name", "parent"}, &resources)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]*types.ManagedObjectReference)
	m := make(objectMap, len(resources))
	for _, r := range resources {
		// We're not interested in the immediate parent (a folder), but the data center.
		p, ok := cache[r.Parent.Value]
		if !ok {
			o := object.NewFolder(root.Client(), *r.Parent)
			var folder mo.Folder
			err := o.Properties(ctx, *r.Parent, []string{"parent"}, &folder)
			if err != nil {
				log.Printf("W! [input.vsphere] Error while getting folder parent: %e", err)
				p = nil
			} else {
				pp := folder.Parent.Reference()
				p = &pp
				cache[r.Parent.Value] = p
			}
		}
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: p}
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
	err := root.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "runtime.host", "config.guestId", "config.uuid"}, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		guest := "unknown"
		uuid := ""
		// Sometimes Config is unknown and returns a nil pointer
		//
		if r.Config != nil {
			guest = cleanGuestID(r.Config.GuestId)
			uuid = r.Config.Uuid
		}
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Runtime.Host, guest: guest, altID: uuid}
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

// Close shuts down an Endpoint and releases any resources associated with it.
func (e *Endpoint) Close() {
	e.clientFactory.Close()
}

// Collect runs a round of data collections as specified in the configuration.
func (e *Endpoint) Collect(ctx context.Context, acc telegraf.Accumulator) error {
	// If we never managed to do a discovery, collection will be a no-op. Therefore,
	// we need to check that a connection is available, or the collection will
	// silently fail.
	//
	if _, err := e.clientFactory.GetClient(ctx); err != nil {
		return err
	}

	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// If discovery interval is disabled (0), discover on each collection cycle
	//
	if e.Parent.ObjectDiscoveryInterval.Duration == 0 {
		err := e.discover(ctx)
		if err != nil {
			return err
		}
	}
	for k, res := range e.resourceKinds {
		if res.enabled {
			err := e.collectResource(ctx, k, acc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Endpoint) chunker(ctx context.Context, f PushFunc, res *resourceKind, now time.Time, latest time.Time) {
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.MaxQueryObjects)
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
			headroom := e.Parent.MaxQueryMetrics - metrics
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
			if mr > 0 || (!res.realTime && metrics >= e.Parent.MaxQueryMetrics) || nRes >= e.Parent.MaxQueryObjects {
				log.Printf("D! [input.vsphere]: Querying %d objects, %d metrics (%d remaining) of type %s for %s. Processed objects: %d. Total objects %d",
					len(pqs), metrics, mr, res.name, e.URL.Host, total+1, len(res.objects))

				// To prevent deadlocks, don't send work items if the context has been cancelled.
				if ctx.Err() == context.Canceled {
					return
				}

				// Call push function
				f(ctx, pqs)
				pqs = make([]types.PerfQuerySpec, 0, e.Parent.MaxQueryObjects)
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

func (e *Endpoint) collectResource(ctx context.Context, resourceType string, acc telegraf.Accumulator) error {

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
			return nil
		}
	} else {
		latest = time.Now().Add(time.Duration(-res.sampling) * time.Second)
	}

	internalTags := map[string]string{"resourcetype": resourceType}
	sw := NewStopwatchWithTags("gather_duration", e.URL.Host, internalTags)

	log.Printf("D! [input.vsphere]: Start of sample period deemed to be %s", latest)
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

	}, e.Parent.CollectConcurrency)

	// Fill the input channel of the worker queue by running the chunking
	// logic implemented in chunker()
	wp.Fill(ctx, func(ctx context.Context, f PushFunc) {
		e.chunker(ctx, f, &res, now, latest)
	})

	// Drain the pool. We're getting errors back. They should all be nil
	var mux sync.Mutex
	err := make(multiError, 0)
	wp.Drain(ctx, func(ctx context.Context, in interface{}) bool {
		if in != nil {
			mux.Lock()
			defer mux.Unlock()
			err = append(err, in.(error))
			return false
		}
		return true
	})
	e.lastColls[resourceType] = now // Use value captured at the beginning to avoid blind spots.

	sw.Stop()
	SendInternalCounterWithTags("gather_count", e.URL.Host, internalTags, count)
	if len(err) > 0 {
		return err
	}
	return nil
}

func (e *Endpoint) collectChunk(ctx context.Context, pqs []types.PerfQuerySpec, resourceType string,
	res resourceKind, acc telegraf.Accumulator) (int, error) {
	count := 0
	prefix := "vsphere" + e.Parent.Separator + resourceType

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return 0, err
	}

	metricInfo, err := client.Perf.CounterInfoByName(ctx)
	if err != nil {
		return count, err
	}

	metrics, err := client.Perf.Query(ctx, pqs)
	if err != nil {
		return count, err
	}

	ems, err := client.Perf.ToMetricSeries(ctx, metrics)
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
				"vcenter": e.URL.Host,
				"source":  instInfo.name,
				"moid":    moid,
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

				// Percentage values must be scaled down by 100.
				info, ok := metricInfo[name]
				if !ok {
					log.Printf("E! [input.vsphere]: Could not determine unit for %s. Skipping", name)
				}
				if info.UnitInfo.GetElementDescription().Key == "percent" {
					bucket.fields[fn] = float64(value) / 100.0
				} else {
					bucket.fields[fn] = value
				}
				count++
			}
		}
		// We've iterated through all the metrics and collected buckets for each
		// measurement name. Now emit them!
		for _, bucket := range buckets {
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	return count, nil
}

func (e *Endpoint) getParent(obj resourceInfo) (resourceInfo, bool) {
	p := obj.parentRef
	if p == nil {
		log.Printf("D! [input.vsphere] No parent found for %s", obj.name)
		return resourceInfo{}, false
	}
	r, ok := e.instanceInfo[p.Value]
	return r, ok
}

func (e *Endpoint) populateTags(objectRef *objectRef, resourceType string, resource *resourceKind, t map[string]string, v *performance.MetricSeries) {
	// Map name of object.
	if resource.pKey != "" {
		t[resource.pKey] = objectRef.name
	}

	if resourceType == "vm" && objectRef.altID != "" {
		t["uuid"] = objectRef.altID
	}

	// Map parent reference
	parent, found := e.instanceInfo[objectRef.parentRef.Value]
	if found {
		t[resource.parentTag] = parent.name
		if resourceType == "vm" {
			if objectRef.guest != "" {
				t["guest"] = objectRef.guest
			}
			if c, ok := e.getParent(parent); ok {
				t["clustername"] = c.name
			}
		}
	}

	// Fill in Datacenter name
	if objectRef.dcname != "" {
		t["dcname"] = objectRef.dcname
	}

	// Determine which point tag to map to the instance
	name := v.Name
	instance := "instance-total"
	if v.Instance != "" {
		instance = v.Instance
	}
	if strings.HasPrefix(name, "cpu.") {
		t["cpu"] = instance
	} else if strings.HasPrefix(name, "datastore.") {
		t["lun"] = instance
	} else if strings.HasPrefix(name, "disk.") {
		t["disk"] = cleanDiskTag(instance)
	} else if strings.HasPrefix(name, "net.") {
		t["interface"] = instance
	} else if strings.HasPrefix(name, "storageAdapter.") {
		t["adapter"] = instance
	} else if strings.HasPrefix(name, "storagePath.") {
		t["path"] = instance
	} else if strings.HasPrefix(name, "sys.resource") {
		t["resource"] = instance
	} else if strings.HasPrefix(name, "vflashModule.") {
		t["module"] = instance
	} else if strings.HasPrefix(name, "virtualDisk.") {
		t["disk"] = instance
	} else if v.Instance != "" {
		// default
		t["instance"] = v.Instance
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
	return strings.TrimSuffix(id, "Guest")
}

func cleanDiskTag(disk string) string {
	// Remove enclosing "<>"
	return strings.TrimSuffix(strings.TrimPrefix(disk, "<"), ">")
}
