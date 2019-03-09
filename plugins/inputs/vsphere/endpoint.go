package vsphere

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf/filter"

	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var isolateLUN = regexp.MustCompile(".*/([^/]+)/?$")

const metricLookback = 3 // Number of time periods to look back at for non-realtime metrics

const rtMetricLookback = 3 // Number of time periods to look back at for realtime metrics

const maxSampleConst = 10 // Absolute maximim number of samples regardless of period

const maxMetadataSamples = 100 // Number of resources to sample for metric metadata

// Endpoint is a high-level representation of a connected vCenter endpoint. It is backed by the lower
// level Client type.
type Endpoint struct {
	Parent          *VSphere
	URL             *url.URL
	resourceKinds   map[string]*resourceKind
	hwMarks         *TSCache
	lun2ds          map[string]string
	discoveryTicker *time.Ticker
	collectMux      sync.RWMutex
	initialized     bool
	clientFactory   *ClientFactory
	busy            sync.Mutex
}

type resourceKind struct {
	name             string
	vcName           string
	pKey             string
	parentTag        string
	enabled          bool
	realTime         bool
	sampling         int32
	objects          objectMap
	filters          filter.Filter
	paths            []string
	collectInstances bool
	getObjects       func(context.Context, *Endpoint, *ResourceFilter) (objectMap, error)
	include          []string
	simple           bool
	metrics          performance.MetricList
	parent           string
	latestSample     time.Time
	lastColl         time.Time
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

func (e *Endpoint) getParent(obj *objectRef, res *resourceKind) (*objectRef, bool) {
	if pKind, ok := e.resourceKinds[res.parent]; ok {
		if p, ok := pKind.objects[obj.parentRef.Value]; ok {
			return &p, true
		}
	}
	return nil, false
}

// NewEndpoint returns a new connection to a vCenter based on the URL and configuration passed
// as parameters.
func NewEndpoint(ctx context.Context, parent *VSphere, url *url.URL) (*Endpoint, error) {
	e := Endpoint{
		URL:           url,
		Parent:        parent,
		hwMarks:       NewTSCache(1 * time.Hour),
		lun2ds:        make(map[string]string),
		initialized:   false,
		clientFactory: NewClientFactory(ctx, url, parent),
	}

	e.resourceKinds = map[string]*resourceKind{
		"datacenter": {
			name:             "datacenter",
			vcName:           "Datacenter",
			pKey:             "dcname",
			parentTag:        "",
			enabled:          anythingEnabled(parent.DatacenterMetricExclude),
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatacenterMetricInclude, parent.DatacenterMetricExclude),
			paths:            parent.DatacenterInclude,
			simple:           isSimple(parent.DatacenterMetricInclude, parent.DatacenterMetricExclude),
			include:          parent.DatacenterMetricInclude,
			collectInstances: parent.DatacenterInstances,
			getObjects:       getDatacenters,
			parent:           "",
		},
		"cluster": {
			name:             "cluster",
			vcName:           "ClusterComputeResource",
			pKey:             "clustername",
			parentTag:        "dcname",
			enabled:          anythingEnabled(parent.ClusterMetricExclude),
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.ClusterMetricInclude, parent.ClusterMetricExclude),
			paths:            parent.ClusterInclude,
			simple:           isSimple(parent.ClusterMetricInclude, parent.ClusterMetricExclude),
			include:          parent.ClusterMetricInclude,
			collectInstances: parent.ClusterInstances,
			getObjects:       getClusters,
			parent:           "datacenter",
		},
		"host": {
			name:             "host",
			vcName:           "HostSystem",
			pKey:             "esxhostname",
			parentTag:        "clustername",
			enabled:          anythingEnabled(parent.HostMetricExclude),
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.HostMetricInclude, parent.HostMetricExclude),
			paths:            parent.HostInclude,
			simple:           isSimple(parent.HostMetricInclude, parent.HostMetricExclude),
			include:          parent.HostMetricInclude,
			collectInstances: parent.HostInstances,
			getObjects:       getHosts,
			parent:           "cluster",
		},
		"vm": {
			name:             "vm",
			vcName:           "VirtualMachine",
			pKey:             "vmname",
			parentTag:        "esxhostname",
			enabled:          anythingEnabled(parent.VMMetricExclude),
			realTime:         true,
			sampling:         20,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.VMMetricInclude, parent.VMMetricExclude),
			paths:            parent.VMInclude,
			simple:           isSimple(parent.VMMetricInclude, parent.VMMetricExclude),
			include:          parent.VMMetricInclude,
			collectInstances: parent.VMInstances,
			getObjects:       getVMs,
			parent:           "host",
		},
		"datastore": {
			name:             "datastore",
			vcName:           "Datastore",
			pKey:             "dsname",
			enabled:          anythingEnabled(parent.DatastoreMetricExclude),
			realTime:         false,
			sampling:         300,
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatastoreMetricInclude, parent.DatastoreMetricExclude),
			paths:            parent.DatastoreInclude,
			simple:           isSimple(parent.DatastoreMetricInclude, parent.DatastoreMetricExclude),
			include:          parent.DatastoreMetricInclude,
			collectInstances: parent.DatastoreInstances,
			getObjects:       getDatastores,
			parent:           "",
		},
	}

	// Start discover and other goodness
	err := e.init(ctx)

	return &e, err
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

func isSimple(include []string, exclude []string) bool {
	if len(exclude) > 0 || len(include) == 0 {
		return false
	}
	for _, s := range include {
		if strings.Contains(s, "*") {
			return false
		}
	}
	return true
}

func (e *Endpoint) startDiscovery(ctx context.Context) {
	e.discoveryTicker = time.NewTicker(e.Parent.ObjectDiscoveryInterval.Duration)
	go func() {
		for {
			select {
			case <-e.discoveryTicker.C:
				err := e.discover(ctx)
				if err != nil && err != context.Canceled {
					log.Printf("E! [inputs.vsphere]: Error in discovery for %s: %v", e.URL.Host, err)
				}
			case <-ctx.Done():
				log.Printf("D! [inputs.vsphere]: Exiting discovery goroutine for %s", e.URL.Host)
				e.discoveryTicker.Stop()
				return
			}
		}
	}()
}

func (e *Endpoint) initalDiscovery(ctx context.Context) {
	err := e.discover(ctx)
	if err != nil && err != context.Canceled {
		log.Printf("E! [inputs.vsphere]: Error in discovery for %s: %v", e.URL.Host, err)
	}
	e.startDiscovery(ctx)
}

func (e *Endpoint) init(ctx context.Context) error {

	if e.Parent.ObjectDiscoveryInterval.Duration > 0 {

		// Run an initial discovery. If force_discovery_on_init isn't set, we kick it off as a
		// goroutine without waiting for it. This will probably cause us to report an empty
		// dataset on the first collection, but it solves the issue of the first collection timing out.
		if e.Parent.ForceDiscoverOnInit {
			log.Printf("D! [inputs.vsphere]: Running initial discovery and waiting for it to finish")
			e.initalDiscovery(ctx)
		} else {
			// Otherwise, just run it in the background. We'll probably have an incomplete first metric
			// collection this way.
			go func() {
				e.initalDiscovery(ctx)
			}()
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

	mn, err := client.CounterInfoByName(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[int32]string)
	for name, m := range mn {
		names[m.Key] = name
	}
	return names, nil
}

func (e *Endpoint) getMetadata(ctx context.Context, obj objectRef, sampling int32) (performance.MetricList, error) {
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
	defer cancel1()
	metrics, err := client.Perf.AvailableMetric(ctx1, obj.ref.Reference(), sampling)
	if err != nil {
		return nil, err
	}
	return metrics, nil
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
		ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
		defer cancel1()
		err := o.Properties(ctx1, here, []string{"parent", "name"}, &result)
		if err != nil {
			log.Printf("W! [inputs.vsphere]: Error while resolving parent. Assuming no parent exists. Error: %s", err)
			break
		}
		if result.Reference().Type == "Datacenter" {
			// Populate cache for the entire chain of objects leading here.
			returnVal = result.Name
			break
		}
		if result.Parent == nil {
			log.Printf("D! [inputs.vsphere]: No parent found for %s (ascending from %s)", here.Reference(), r.Reference())
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

	log.Printf("D! [inputs.vsphere]: Discover new objects for %s", e.URL.Host)
	resourceKinds := make(map[string]resourceKind)
	dcNameCache := make(map[string]string)

	numRes := int64(0)

	// Populate resource objects, and endpoint instance info.
	newObjects := make(map[string]objectMap)
	for k, res := range e.resourceKinds {
		log.Printf("D! [inputs.vsphere] Discovering resources for %s", res.name)
		// Need to do this for all resource types even if they are not enabled
		if res.enabled || k != "vm" {
			rf := ResourceFilter{
				finder:  &Finder{client},
				resType: res.vcName,
				paths:   res.paths}

			ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
			defer cancel1()
			objects, err := res.getObjects(ctx1, e, &rf)
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

			// No need to collect metric metadata if resource type is not enabled
			if res.enabled {
				if res.simple {
					e.simpleMetadataSelect(ctx, client, res)
				} else {
					e.complexMetadataSelect(ctx, res, objects, metricNames)
				}
			}
			newObjects[k] = objects

			SendInternalCounterWithTags("discovered_objects", e.URL.Host, map[string]string{"type": res.name}, int64(len(objects)))
			numRes += int64(len(objects))
		}
	}

	// Build lun2ds map
	dss := resourceKinds["datastore"]
	l2d := make(map[string]string)
	for _, ds := range dss.objects {
		url := ds.altID
		m := isolateLUN.FindStringSubmatch(url)
		if m != nil {
			l2d[m[1]] = ds.name
		}
	}

	// Atomically swap maps
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	for k, v := range newObjects {
		e.resourceKinds[k].objects = v
	}
	e.lun2ds = l2d

	sw.Stop()
	SendInternalCounterWithTags("discovered_objects", e.URL.Host, map[string]string{"type": "instance-total"}, numRes)
	return nil
}

func (e *Endpoint) simpleMetadataSelect(ctx context.Context, client *Client, res *resourceKind) {
	log.Printf("D! [inputs.vsphere] Using fast metric metadata selection for %s", res.name)
	m, err := client.CounterInfoByName(ctx)
	if err != nil {
		log.Printf("E! [inputs.vsphere]: Error while getting metric metadata. Discovery will be incomplete. Error: %s", err)
		return
	}
	res.metrics = make(performance.MetricList, 0, len(res.include))
	for _, s := range res.include {
		if pci, ok := m[s]; ok {
			cnt := types.PerfMetricId{
				CounterId: pci.Key,
			}
			if res.collectInstances {
				cnt.Instance = "*"
			} else {
				cnt.Instance = ""
			}
			res.metrics = append(res.metrics, cnt)
		} else {
			log.Printf("W! [inputs.vsphere] Metric name %s is unknown. Will not be collected", s)
		}
	}
}

func (e *Endpoint) complexMetadataSelect(ctx context.Context, res *resourceKind, objects objectMap, metricNames map[int32]string) {
	// We're only going to get metadata from maxMetadataSamples resources. If we have
	// more resources than that, we pick maxMetadataSamples samples at random.
	sampledObjects := make([]objectRef, len(objects))
	i := 0
	for _, obj := range objects {
		sampledObjects[i] = obj
		i++
	}
	n := len(sampledObjects)
	if n > maxMetadataSamples {
		// Shuffle samples into the maxMetadatSamples positions
		for i := 0; i < maxMetadataSamples; i++ {
			j := int(rand.Int31n(int32(i + 1)))
			t := sampledObjects[i]
			sampledObjects[i] = sampledObjects[j]
			sampledObjects[j] = t
		}
		sampledObjects = sampledObjects[0:maxMetadataSamples]
	}

	instInfoMux := sync.Mutex{}
	te := NewThrottledExecutor(e.Parent.DiscoverConcurrency)
	for _, obj := range sampledObjects {
		func(obj objectRef) {
			te.Run(ctx, func() {
				metrics, err := e.getMetadata(ctx, obj, res.sampling)
				if err != nil {
					log.Printf("E! [inputs.vsphere]: Error while getting metric metadata. Discovery will be incomplete. Error: %s", err)
				}
				mMap := make(map[string]types.PerfMetricId)
				for _, m := range metrics {
					if m.Instance != "" && res.collectInstances {
						m.Instance = "*"
					} else {
						m.Instance = ""
					}
					if res.filters.Match(metricNames[m.CounterId]) {
						mMap[strconv.Itoa(int(m.CounterId))+"|"+m.Instance] = m
					}
				}
				log.Printf("D! [inputs.vsphere] Found %d metrics for %s", len(mMap), obj.name)
				instInfoMux.Lock()
				defer instInfoMux.Unlock()
				if len(mMap) > len(res.metrics) {
					res.metrics = make(performance.MetricList, len(mMap))
					i := 0
					for _, m := range mMap {
						res.metrics[i] = m
						i++
					}
				}
			})
		}(obj)
	}
	te.Wait()
}

func getDatacenters(ctx context.Context, e *Endpoint, filter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datacenter
	ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
	defer cancel1()
	err := filter.FindAll(ctx1, &resources)
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

func getClusters(ctx context.Context, e *Endpoint, filter *ResourceFilter) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
	defer cancel1()
	err := filter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]*types.ManagedObjectReference)
	m := make(objectMap, len(resources))
	for _, r := range resources {
		// We're not interested in the immediate parent (a folder), but the data center.
		p, ok := cache[r.Parent.Value]
		if !ok {
			ctx2, cancel2 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
			defer cancel2()
			client, err := e.clientFactory.GetClient(ctx2)
			if err != nil {
				return nil, err
			}
			o := object.NewFolder(client.Client.Client, *r.Parent)
			var folder mo.Folder
			ctx3, cancel3 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
			defer cancel3()
			err = o.Properties(ctx3, *r.Parent, []string{"parent"}, &folder)
			if err != nil {
				log.Printf("W! [inputs.vsphere] Error while getting folder parent: %e", err)
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

func getHosts(ctx context.Context, e *Endpoint, filter *ResourceFilter) (objectMap, error) {
	var resources []mo.HostSystem
	err := filter.FindAll(ctx, &resources)
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

func getVMs(ctx context.Context, e *Endpoint, filter *ResourceFilter) (objectMap, error) {
	var resources []mo.VirtualMachine
	ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
	defer cancel1()
	err := filter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		if r.Runtime.PowerState != "poweredOn" {
			continue
		}
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

func getDatastores(ctx context.Context, e *Endpoint, filter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datastore
	ctx1, cancel1 := context.WithTimeout(ctx, e.Parent.Timeout.Duration)
	defer cancel1()
	err := filter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		url := ""
		if r.Info != nil {
			info := r.Info.GetDatastoreInfo()
			if info != nil {
				url = info.Url
			}
		}
		m[r.ExtensibleManagedObject.Reference().Value] = objectRef{
			name: r.Name, ref: r.ExtensibleManagedObject.Reference(), parentRef: r.Parent, altID: url}
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
	if _, err := e.clientFactory.GetClient(ctx); err != nil {
		return err
	}

	e.collectMux.RLock()
	defer e.collectMux.RUnlock()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// If discovery interval is disabled (0), discover on each collection cycle
	if e.Parent.ObjectDiscoveryInterval.Duration == 0 {
		err := e.discover(ctx)
		if err != nil {
			return err
		}
	}
	var wg sync.WaitGroup
	for k, res := range e.resourceKinds {
		if res.enabled {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				err := e.collectResource(ctx, k, acc)
				if err != nil {
					acc.AddError(err)
				}
			}(k)
		}
	}
	wg.Wait()

	// Purge old timestamps from the cache
	e.hwMarks.Purge()
	return nil
}

// Workaround to make sure pqs is a copy of the loop variable and won't change.
func submitChunkJob(ctx context.Context, te *ThrottledExecutor, job func([]types.PerfQuerySpec), pqs []types.PerfQuerySpec) {
	te.Run(ctx, func() {
		job(pqs)
	})
}

func (e *Endpoint) chunkify(ctx context.Context, res *resourceKind, now time.Time, latest time.Time, acc telegraf.Accumulator, job func([]types.PerfQuerySpec)) {
	te := NewThrottledExecutor(e.Parent.CollectConcurrency)
	maxMetrics := e.Parent.MaxQueryMetrics
	if maxMetrics < 1 {
		maxMetrics = 1
	}

	// Workaround for vCenter weirdness. Cluster metrics seem to count multiple times
	// when checking query size, so keep it at a low value.
	// Revisit this when we better understand the reason why vCenter counts it this way!
	if res.name == "cluster" && maxMetrics > 10 {
		maxMetrics = 10
	}
	pqs := make([]types.PerfQuerySpec, 0, e.Parent.MaxQueryObjects)
	metrics := 0
	total := 0
	nRes := 0
	for _, object := range res.objects {
		mr := len(res.metrics)
		for mr > 0 {
			mc := mr
			headroom := maxMetrics - metrics
			if !res.realTime && mc > headroom { // Metric query limit only applies to non-realtime metrics
				mc = headroom
			}
			fm := len(res.metrics) - mr
			pq := types.PerfQuerySpec{
				Entity:     object.ref,
				MaxSample:  maxSampleConst,
				MetricId:   res.metrics[fm : fm+mc],
				IntervalId: res.sampling,
				Format:     "normal",
			}

			start, ok := e.hwMarks.Get(object.ref.Value)
			if !ok {
				// Look back 3 sampling periods by default
				start = latest.Add(time.Duration(-res.sampling) * time.Second * (metricLookback - 1))
			}
			pq.StartTime = &start
			pq.EndTime = &now

			// Make sure endtime is always after start time. We may occasionally see samples from the future
			// returned from vCenter. This is presumably due to time drift between vCenter and EXSi nodes.
			if pq.StartTime.After(*pq.EndTime) {
				log.Printf("D! [inputs.vsphere] Future sample. Res: %s, StartTime: %s, EndTime: %s, Now: %s", pq.Entity, *pq.StartTime, *pq.EndTime, now)
				end := start.Add(time.Second)
				pq.EndTime = &end
			}

			pqs = append(pqs, pq)
			mr -= mc
			metrics += mc

			// We need to dump the current chunk of metrics for one of two reasons:
			// 1) We filled up the metric quota while processing the current resource
			// 2) We are at the last resource and have no more data to process.
			// 3) The query contains more than 100,000 individual metrics
			if mr > 0 || nRes >= e.Parent.MaxQueryObjects || len(pqs) > 100000 {
				log.Printf("D! [inputs.vsphere]: Queueing query: %d objects, %d metrics (%d remaining) of type %s for %s. Processed objects: %d. Total objects %d",
					len(pqs), metrics, mr, res.name, e.URL.Host, total+1, len(res.objects))

				// Don't send work items if the context has been cancelled.
				if ctx.Err() == context.Canceled {
					return
				}

				// Run collection job
				submitChunkJob(ctx, te, job, pqs)
				pqs = make([]types.PerfQuerySpec, 0, e.Parent.MaxQueryObjects)
				metrics = 0
				nRes = 0
			}
		}
		total++
		nRes++
	}
	// Handle final partially filled chunk
	if len(pqs) > 0 {
		// Run collection job
		log.Printf("D! [inputs.vsphere]: Queuing query: %d objects, %d metrics (0 remaining) of type %s for %s. Total objects %d (final chunk)",
			len(pqs), metrics, res.name, e.URL.Host, len(res.objects))
		submitChunkJob(ctx, te, job, pqs)
	}

	// Wait for background collection to finish
	te.Wait()
}

func (e *Endpoint) collectResource(ctx context.Context, resourceType string, acc telegraf.Accumulator) error {
	res := e.resourceKinds[resourceType]
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}
	now, err := client.GetServerTime(ctx)
	if err != nil {
		return err
	}

	// Estimate the interval at which we're invoked. Use local time (not server time)
	// since this is about how we got invoked locally.
	localNow := time.Now()
	estInterval := time.Duration(time.Minute)
	if !res.lastColl.IsZero() {
		estInterval = localNow.Sub(res.lastColl).Truncate(time.Duration(res.sampling) * time.Second)
	}
	log.Printf("D! [inputs.vsphere] Interval estimated to %s", estInterval)

	latest := res.latestSample
	if !latest.IsZero() {
		elapsed := now.Sub(latest).Seconds() + 5.0 // Allow 5 second jitter.
		log.Printf("D! [inputs.vsphere]: Latest: %s, elapsed: %f, resource: %s", latest, elapsed, resourceType)
		if !res.realTime && elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			log.Printf("D! [inputs.vsphere]: Sampling period for %s of %d has not elapsed on %s",
				resourceType, res.sampling, e.URL.Host)
			return nil
		}
	} else {
		latest = now.Add(time.Duration(-res.sampling) * time.Second)
	}

	internalTags := map[string]string{"resourcetype": resourceType}
	sw := NewStopwatchWithTags("gather_duration", e.URL.Host, internalTags)

	log.Printf("D! [inputs.vsphere]: Collecting metrics for %d objects of type %s for %s",
		len(res.objects), resourceType, e.URL.Host)

	count := int64(0)

	var tsMux sync.Mutex
	latestSample := time.Time{}

	// Divide workload into chunks and process them concurrently
	e.chunkify(ctx, res, now, latest, acc,
		func(chunk []types.PerfQuerySpec) {
			n, localLatest, err := e.collectChunk(ctx, chunk, res, acc, now, estInterval)
			log.Printf("D! [inputs.vsphere] CollectChunk for %s returned %d metrics", resourceType, n)
			if err != nil {
				acc.AddError(errors.New("While collecting " + res.name + ": " + err.Error()))
			}
			atomic.AddInt64(&count, int64(n))
			tsMux.Lock()
			defer tsMux.Unlock()
			if localLatest.After(latestSample) && !localLatest.IsZero() {
				latestSample = localLatest
			}
		})

	log.Printf("D! [inputs.vsphere] Latest sample for %s set to %s", resourceType, latestSample)
	if !latestSample.IsZero() {
		res.latestSample = latestSample
	}
	sw.Stop()
	SendInternalCounterWithTags("gather_count", e.URL.Host, internalTags, count)
	return nil
}

func alignSamples(info []types.PerfSampleInfo, values []int64, interval time.Duration) ([]types.PerfSampleInfo, []float64) {
	rInfo := make([]types.PerfSampleInfo, 0, len(info))
	rValues := make([]float64, 0, len(values))
	bi := 1.0
	var lastBucket time.Time
	for idx := range info {
		// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
		// data coming back with missing values. Take care of that gracefully!
		if idx >= len(values) {
			log.Printf("D! [inputs.vsphere] len(SampleInfo)>len(Value) %d > %d", len(info), len(values))
			break
		}
		v := float64(values[idx])
		if v < 0 {
			continue
		}
		ts := info[idx].Timestamp
		roundedTs := ts.Truncate(interval)

		// Are we still working on the same bucket?
		if roundedTs == lastBucket {
			bi++
			p := len(rValues) - 1
			rValues[p] = ((bi-1)/bi)*float64(rValues[p]) + v/bi
		} else {
			rValues = append(rValues, v)
			roundedInfo := types.PerfSampleInfo{
				Timestamp: roundedTs,
				Interval:  info[idx].Interval,
			}
			rInfo = append(rInfo, roundedInfo)
			bi = 1.0
			lastBucket = roundedTs
		}
	}
	//log.Printf("D! [inputs.vsphere] Aligned samples: %d collapsed into %d", len(info), len(rInfo))
	return rInfo, rValues
}

func (e *Endpoint) collectChunk(ctx context.Context, pqs []types.PerfQuerySpec, res *resourceKind, acc telegraf.Accumulator, now time.Time, interval time.Duration) (int, time.Time, error) {
	log.Printf("D! [inputs.vsphere] Query for %s has %d QuerySpecs", res.name, len(pqs))
	latestSample := time.Time{}
	count := 0
	resourceType := res.name
	prefix := "vsphere" + e.Parent.Separator + resourceType

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return count, latestSample, err
	}

	metricInfo, err := client.CounterInfoByName(ctx)
	if err != nil {
		return count, latestSample, err
	}

	ems, err := client.QueryMetrics(ctx, pqs)
	if err != nil {
		return count, latestSample, err
	}

	log.Printf("D! [inputs.vsphere] Query for %s returned metrics for %d objects", resourceType, len(ems))

	// Iterate through results
	for _, em := range ems {
		moid := em.Entity.Reference().Value
		instInfo, found := res.objects[moid]
		if !found {
			log.Printf("E! [inputs.vsphere]: MOID %s not found in cache. Skipping! (This should not happen!)", moid)
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
				log.Printf("E! [inputs.vsphere]: MOID %s not found in cache. Skipping", moid)
				continue
			}
			e.populateTags(&objectRef, resourceType, res, t, &v)

			nValues := 0
			alignedInfo, alignedValues := alignSamples(em.SampleInfo, v.Value, interval)

			for idx, sample := range alignedInfo {
				// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
				// data coming back with missing values. Take care of that gracefully!
				if idx >= len(alignedValues) {
					log.Printf("D! [inputs.vsphere] len(SampleInfo)>len(Value) %d > %d", len(alignedInfo), len(alignedValues))
					break
				}
				ts := sample.Timestamp
				if ts.After(latestSample) {
					latestSample = ts
				}
				nValues++

				// Organize the metrics into a bucket per measurement.
				mn, fn := e.makeMetricIdentifier(prefix, name)
				bKey := mn + " " + v.Instance + " " + strconv.FormatInt(ts.UnixNano(), 10)
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: t}
					buckets[bKey] = bucket
				}

				// Percentage values must be scaled down by 100.
				info, ok := metricInfo[name]
				if !ok {
					log.Printf("E! [inputs.vsphere]: Could not determine unit for %s. Skipping", name)
				}
				v := alignedValues[idx]
				if info.UnitInfo.GetElementDescription().Key == "percent" {
					bucket.fields[fn] = float64(v) / 100.0
				} else {
					if e.Parent.UseIntSamples {
						bucket.fields[fn] = int64(round(v))
					} else {
						bucket.fields[fn] = v
					}
				}
				count++

				// Update highwater marks
				e.hwMarks.Put(moid, ts)
			}
			if nValues == 0 {
				log.Printf("D! [inputs.vsphere]: Missing value for: %s, %s", name, objectRef.name)
				continue
			}
		}
		// We've iterated through all the metrics and collected buckets for each
		// measurement name. Now emit them!
		for _, bucket := range buckets {
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	return count, latestSample, nil
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
	parent, found := e.getParent(objectRef, resource)
	if found {
		t[resource.parentTag] = parent.name
		if resourceType == "vm" {
			if objectRef.guest != "" {
				t["guest"] = objectRef.guest
			}
			if c, ok := e.resourceKinds["cluster"].objects[parent.parentRef.Value]; ok {
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
		if ds, ok := e.lun2ds[instance]; ok {
			t["dsname"] = ds
		} else {
			t["dsname"] = instance
		}
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

func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}
