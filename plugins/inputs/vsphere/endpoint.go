package vsphere

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/influxdata/telegraf/filter"
)

var isolateLUN = regexp.MustCompile(`.*/([^/]+)/?$`)

var isIPv4 = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)

var isIPv6 = regexp.MustCompile(`^(?:[A-Fa-f0-9]{0,4}:){1,7}[A-Fa-f0-9]{1,4}$`)

const maxSampleConst = 10 // Absolute maximum number of samples regardless of period

const maxMetadataSamples = 100 // Number of resources to sample for metric metadata

const maxRealtimeMetrics = 50000 // Absolute maximum metrics per realtime query

const hwMarkTTL = 4 * time.Hour

type queryChunk []types.PerfQuerySpec

type queryJob func(queryChunk)

// Endpoint is a high-level representation of a connected vCenter endpoint. It is backed by the lower
// level Client type.
type Endpoint struct {
	Parent            *VSphere
	URL               *url.URL
	resourceKinds     map[string]*resourceKind
	hwMarks           *TSCache
	lun2ds            map[string]string
	discoveryTicker   *time.Ticker
	collectMux        sync.RWMutex
	initialized       bool
	clientFactory     *ClientFactory
	busy              sync.Mutex
	customFields      map[int32]string
	customAttrFilter  filter.Filter
	customAttrEnabled bool
	metricNameLookup  map[int32]string
	metricNameMux     sync.RWMutex
	log               telegraf.Logger
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
	excludePaths     []string
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

type objectMap map[string]*objectRef

type objectRef struct {
	name         string
	altID        string
	ref          types.ManagedObjectReference
	parentRef    *types.ManagedObjectReference //Pointer because it must be nillable
	guest        string
	dcname       string
	customValues map[string]string
	lookup       map[string]string
}

func (e *Endpoint) getParent(obj *objectRef, res *resourceKind) (*objectRef, bool) {
	if pKind, ok := e.resourceKinds[res.parent]; ok {
		if p, ok := pKind.objects[obj.parentRef.Value]; ok {
			return p, true
		}
	}
	return nil, false
}

// NewEndpoint returns a new connection to a vCenter based on the URL and configuration passed
// as parameters.
func NewEndpoint(ctx context.Context, parent *VSphere, address *url.URL, log telegraf.Logger) (*Endpoint, error) {
	e := Endpoint{
		URL:               address,
		Parent:            parent,
		hwMarks:           NewTSCache(hwMarkTTL, log),
		lun2ds:            make(map[string]string),
		initialized:       false,
		clientFactory:     NewClientFactory(address, parent),
		customAttrFilter:  newFilterOrPanic(parent.CustomAttributeInclude, parent.CustomAttributeExclude),
		customAttrEnabled: anythingEnabled(parent.CustomAttributeExclude),
		log:               log,
	}

	e.resourceKinds = map[string]*resourceKind{
		"datacenter": {
			name:             "datacenter",
			vcName:           "Datacenter",
			pKey:             "dcname",
			parentTag:        "",
			enabled:          anythingEnabled(parent.DatacenterMetricExclude),
			realTime:         false,
			sampling:         int32(time.Duration(parent.HistoricalInterval).Seconds()),
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatacenterMetricInclude, parent.DatacenterMetricExclude),
			paths:            parent.DatacenterInclude,
			excludePaths:     parent.DatacenterExclude,
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
			sampling:         int32(time.Duration(parent.HistoricalInterval).Seconds()),
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.ClusterMetricInclude, parent.ClusterMetricExclude),
			paths:            parent.ClusterInclude,
			excludePaths:     parent.ClusterExclude,
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
			excludePaths:     parent.HostExclude,
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
			excludePaths:     parent.VMExclude,
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
			sampling:         int32(time.Duration(parent.HistoricalInterval).Seconds()),
			objects:          make(objectMap),
			filters:          newFilterOrPanic(parent.DatastoreMetricInclude, parent.DatastoreMetricExclude),
			paths:            parent.DatastoreInclude,
			excludePaths:     parent.DatastoreExclude,
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
	e.discoveryTicker = time.NewTicker(time.Duration(e.Parent.ObjectDiscoveryInterval))
	go func() {
		for {
			select {
			case <-e.discoveryTicker.C:
				err := e.discover(ctx)
				if err != nil && err != context.Canceled {
					e.log.Errorf("Discovery for %s: %s", e.URL.Host, err.Error())
				}
			case <-ctx.Done():
				e.log.Debugf("Exiting discovery goroutine for %s", e.URL.Host)
				e.discoveryTicker.Stop()
				return
			}
		}
	}()
}

func (e *Endpoint) initalDiscovery(ctx context.Context) {
	err := e.discover(ctx)
	if err != nil && err != context.Canceled {
		e.log.Errorf("Discovery for %s: %s", e.URL.Host, err.Error())
	}
	e.startDiscovery(ctx)
}

func (e *Endpoint) init(ctx context.Context) error {
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	// Initial load of custom field metadata
	if e.customAttrEnabled {
		fields, err := client.GetCustomFields(ctx)
		if err != nil {
			e.log.Warn("Could not load custom field metadata")
		} else {
			e.customFields = fields
		}
	}

	if time.Duration(e.Parent.ObjectDiscoveryInterval) > 0 {
		e.Parent.Log.Debug("Running initial discovery")
		e.initalDiscovery(ctx)
	}
	e.initialized = true
	return nil
}

func (e *Endpoint) getMetricNameForID(id int32) string {
	e.metricNameMux.RLock()
	defer e.metricNameMux.RUnlock()
	return e.metricNameLookup[id]
}

func (e *Endpoint) reloadMetricNameMap(ctx context.Context) error {
	e.metricNameMux.Lock()
	defer e.metricNameMux.Unlock()
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	mn, err := client.CounterInfoByKey(ctx)
	if err != nil {
		return err
	}
	e.metricNameLookup = make(map[int32]string)
	for key, m := range mn {
		e.metricNameLookup[key] = m.Name()
	}
	return nil
}

func (e *Endpoint) getMetadata(ctx context.Context, obj *objectRef, sampling int32) (performance.MetricList, error) {
	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
	defer cancel1()
	metrics, err := client.Perf.AvailableMetric(ctx1, obj.ref.Reference(), sampling)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (e *Endpoint) getDatacenterName(ctx context.Context, client *Client, cache map[string]string, r types.ManagedObjectReference) (string, bool) {
	return e.getAncestorName(ctx, client, "Datacenter", cache, r)
}

func (e *Endpoint) getAncestorName(ctx context.Context, client *Client, resourceType string, cache map[string]string, r types.ManagedObjectReference) (string, bool) {
	path := make([]string, 0)
	returnVal := ""
	here := r
	done := false
	for !done {
		done = func() bool {
			if name, ok := cache[here.Reference().String()]; ok {
				// Populate cache for the entire chain of objects leading here.
				returnVal = name
				return true
			}
			path = append(path, here.Reference().String())
			o := object.NewCommon(client.Client.Client, r)
			var result mo.ManagedEntity
			ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
			defer cancel1()
			err := o.Properties(ctx1, here, []string{"parent", "name"}, &result)
			if err != nil {
				e.Parent.Log.Warnf("Error while resolving parent. Assuming no parent exists. Error: %s", err.Error())
				return true
			}
			if result.Reference().Type == resourceType {
				// Populate cache for the entire chain of objects leading here.
				returnVal = result.Name
				return true
			}
			if result.Parent == nil {
				e.Parent.Log.Debugf("No parent found for %s (ascending from %s)", here.Reference(), r.Reference())
				return true
			}
			here = result.Parent.Reference()
			return false
		}()
	}
	for _, s := range path {
		cache[s] = returnVal
	}
	return returnVal, returnVal != ""
}

func (e *Endpoint) discover(ctx context.Context) error {
	e.busy.Lock()
	defer e.busy.Unlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := e.reloadMetricNameMap(ctx)
	if err != nil {
		return err
	}

	sw := NewStopwatch("discover", e.URL.Host)

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}

	e.log.Debugf("Discover new objects for %s", e.URL.Host)
	dcNameCache := make(map[string]string)

	numRes := int64(0)

	// Populate resource objects, and endpoint instance info.
	newObjects := make(map[string]objectMap)
	for k, res := range e.resourceKinds {
		e.log.Debugf("Discovering resources for %s", res.name)
		// Need to do this for all resource types even if they are not enabled
		if res.enabled || k != "vm" {
			rf := ResourceFilter{
				finder:       &Finder{client},
				resType:      res.vcName,
				paths:        res.paths,
				excludePaths: res.excludePaths}

			ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
			objects, err := res.getObjects(ctx1, e, &rf)
			cancel1()
			if err != nil {
				return err
			}

			// Fill in datacenter names where available (no need to do it for Datacenters)
			if res.name != "datacenter" {
				for k, obj := range objects {
					if obj.parentRef != nil {
						obj.dcname, _ = e.getDatacenterName(ctx, client, dcNameCache, *obj.parentRef)
						objects[k] = obj
					}
				}
			}

			// No need to collect metric metadata if resource type is not enabled
			if res.enabled {
				if res.simple {
					e.simpleMetadataSelect(ctx, client, res)
				} else {
					e.complexMetadataSelect(ctx, res, objects)
				}
			}
			newObjects[k] = objects

			SendInternalCounterWithTags("discovered_objects", e.URL.Host, map[string]string{"type": res.name}, int64(len(objects)))
			numRes += int64(len(objects))
		}
	}

	// Build lun2ds map
	dss := newObjects["datastore"]
	l2d := make(map[string]string)
	for _, ds := range dss {
		lunID := ds.altID
		m := isolateLUN.FindStringSubmatch(lunID)
		if m != nil {
			l2d[m[1]] = ds.name
		}
	}

	// Load custom field metadata
	var fields map[int32]string
	if e.customAttrEnabled {
		fields, err = client.GetCustomFields(ctx)
		if err != nil {
			e.log.Warn("Could not load custom field metadata")
			fields = nil
		}
	}

	// Atomically swap maps
	e.collectMux.Lock()
	defer e.collectMux.Unlock()

	for k, v := range newObjects {
		e.resourceKinds[k].objects = v
	}
	e.lun2ds = l2d

	if fields != nil {
		e.customFields = fields
	}

	sw.Stop()
	SendInternalCounterWithTags("discovered_objects", e.URL.Host, map[string]string{"type": "instance-total"}, numRes)
	return nil
}

func (e *Endpoint) simpleMetadataSelect(ctx context.Context, client *Client, res *resourceKind) {
	e.log.Debugf("Using fast metric metadata selection for %s", res.name)
	m, err := client.CounterInfoByName(ctx)
	if err != nil {
		e.log.Errorf("Getting metric metadata. Discovery will be incomplete. Error: %s", err.Error())
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
			e.log.Warnf("Metric name %s is unknown. Will not be collected", s)
		}
	}
}

func (e *Endpoint) complexMetadataSelect(ctx context.Context, res *resourceKind, objects objectMap) {
	// We're only going to get metadata from maxMetadataSamples resources. If we have
	// more resources than that, we pick maxMetadataSamples samples at random.
	sampledObjects := make([]*objectRef, len(objects))
	i := 0
	for _, obj := range objects {
		sampledObjects[i] = obj
		i++
	}
	n := len(sampledObjects)
	if n > maxMetadataSamples {
		// Shuffle samples into the maxMetadataSamples positions
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
		func(obj *objectRef) {
			te.Run(ctx, func() {
				metrics, err := e.getMetadata(ctx, obj, res.sampling)
				if err != nil {
					e.log.Errorf("Getting metric metadata. Discovery will be incomplete. Error: %s", err.Error())
				}
				mMap := make(map[string]types.PerfMetricId)
				for _, m := range metrics {
					if m.Instance != "" && res.collectInstances {
						m.Instance = "*"
					} else {
						m.Instance = ""
					}
					if res.filters.Match(e.getMetricNameForID(m.CounterId)) {
						mMap[strconv.Itoa(int(m.CounterId))+"|"+m.Instance] = m
					}
				}
				e.log.Debugf("Found %d metrics for %s", len(mMap), obj.name)
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

func getDatacenters(ctx context.Context, e *Endpoint, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datacenter
	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap, len(resources))
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Parent,
			dcname:       r.Name,
			customValues: e.loadCustomAttributes(&r.ManagedEntity),
		}
	}
	return m, nil
}

func getClusters(ctx context.Context, e *Endpoint, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.ClusterComputeResource
	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]*types.ManagedObjectReference)
	m := make(objectMap, len(resources))
	for _, r := range resources {
		// Wrap in a function to make defer work correctly.
		err := func() error {
			// We're not interested in the immediate parent (a folder), but the data center.
			p, ok := cache[r.Parent.Value]
			if !ok {
				ctx2, cancel2 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
				defer cancel2()
				client, err := e.clientFactory.GetClient(ctx2)
				if err != nil {
					return err
				}
				o := object.NewFolder(client.Client.Client, *r.Parent)
				var folder mo.Folder
				ctx3, cancel3 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
				defer cancel3()
				err = o.Properties(ctx3, *r.Parent, []string{"parent"}, &folder)
				if err != nil {
					e.Parent.Log.Warnf("Error while getting folder parent: %s", err.Error())
					p = nil
				} else {
					pp := folder.Parent.Reference()
					p = &pp
					cache[r.Parent.Value] = p
				}
			}
			m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
				name:         r.Name,
				ref:          r.ExtensibleManagedObject.Reference(),
				parentRef:    p,
				customValues: e.loadCustomAttributes(&r.ManagedEntity),
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

//noinspection GoUnusedParameter
func getHosts(ctx context.Context, e *Endpoint, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.HostSystem
	err := resourceFilter.FindAll(ctx, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Parent,
			customValues: e.loadCustomAttributes(&r.ManagedEntity),
		}
	}
	return m, nil
}

func getVMs(ctx context.Context, e *Endpoint, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.VirtualMachine
	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
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
		lookup := make(map[string]string)

		// Extract host name
		if r.Guest != nil && r.Guest.HostName != "" {
			lookup["guesthostname"] = r.Guest.HostName
		}

		// Collect network information
		for _, net := range r.Guest.Net {
			if net.DeviceConfigId == -1 {
				continue
			}
			if net.IpConfig == nil || net.IpConfig.IpAddress == nil {
				continue
			}
			ips := make(map[string][]string)
			for _, ip := range net.IpConfig.IpAddress {
				addr := ip.IpAddress
				for _, ipType := range e.Parent.IPAddresses {
					if !(ipType == "ipv4" && isIPv4.MatchString(addr) ||
						ipType == "ipv6" && isIPv6.MatchString(addr)) {
						continue
					}

					// By convention, we want the preferred addresses to appear first in the array.
					if _, ok := ips[ipType]; !ok {
						ips[ipType] = make([]string, 0)
					}
					if ip.State == "preferred" {
						ips[ipType] = append([]string{addr}, ips[ipType]...)
					} else {
						ips[ipType] = append(ips[ipType], addr)
					}
				}
			}
			for ipType, ipList := range ips {
				lookup["nic/"+strconv.Itoa(int(net.DeviceConfigId))+"/"+ipType] = strings.Join(ipList, ",")
			}
		}

		// Sometimes Config is unknown and returns a nil pointer
		if r.Config != nil {
			guest = cleanGuestID(r.Config.GuestId)
			uuid = r.Config.Uuid
		}
		cvs := make(map[string]string)
		if e.customAttrEnabled {
			for _, cv := range r.Summary.CustomValue {
				val := cv.(*types.CustomFieldStringValue)
				if val.Value == "" {
					continue
				}
				key, ok := e.customFields[val.Key]
				if !ok {
					e.log.Warnf("Metadata for custom field %d not found. Skipping", val.Key)
					continue
				}
				if e.customAttrFilter.Match(key) {
					cvs[key] = val.Value
				}
			}
		}
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Runtime.Host,
			guest:        guest,
			altID:        uuid,
			customValues: e.loadCustomAttributes(&r.ManagedEntity),
			lookup:       lookup,
		}
	}
	return m, nil
}

func getDatastores(ctx context.Context, e *Endpoint, resourceFilter *ResourceFilter) (objectMap, error) {
	var resources []mo.Datastore
	ctx1, cancel1 := context.WithTimeout(ctx, time.Duration(e.Parent.Timeout))
	defer cancel1()
	err := resourceFilter.FindAll(ctx1, &resources)
	if err != nil {
		return nil, err
	}
	m := make(objectMap)
	for _, r := range resources {
		lunID := ""
		if r.Info != nil {
			info := r.Info.GetDatastoreInfo()
			if info != nil {
				lunID = info.Url
			}
		}
		m[r.ExtensibleManagedObject.Reference().Value] = &objectRef{
			name:         r.Name,
			ref:          r.ExtensibleManagedObject.Reference(),
			parentRef:    r.Parent,
			altID:        lunID,
			customValues: e.loadCustomAttributes(&r.ManagedEntity),
		}
	}
	return m, nil
}

func (e *Endpoint) loadCustomAttributes(entity *mo.ManagedEntity) map[string]string {
	if !e.customAttrEnabled {
		return map[string]string{}
	}
	cvs := make(map[string]string)
	for _, v := range entity.CustomValue {
		cv, ok := v.(*types.CustomFieldStringValue)
		if !ok {
			e.Parent.Log.Warnf("Metadata for custom field %d not of string type. Skipping", cv.Key)
			continue
		}
		key, ok := e.customFields[cv.Key]
		if !ok {
			e.Parent.Log.Warnf("Metadata for custom field %d not found. Skipping", cv.Key)
			continue
		}
		if e.customAttrFilter.Match(key) {
			cvs[key] = cv.Value
		}
	}
	return cvs
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
	if time.Duration(e.Parent.ObjectDiscoveryInterval) == 0 {
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
func submitChunkJob(ctx context.Context, te *ThrottledExecutor, job queryJob, pqs queryChunk) {
	te.Run(ctx, func() {
		job(pqs)
	})
}

func (e *Endpoint) chunkify(ctx context.Context, res *resourceKind, now time.Time, latest time.Time, job queryJob) {
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

	pqs := make(queryChunk, 0, e.Parent.MaxQueryObjects)
	numQs := 0

	for _, obj := range res.objects {
		timeBuckets := make(map[int64]*types.PerfQuerySpec)
		for metricIdx, metric := range res.metrics {
			// Determine time of last successful collection
			metricName := e.getMetricNameForID(metric.CounterId)
			if metricName == "" {
				e.log.Debugf("Unable to find metric name for id %d. Skipping!", metric.CounterId)
				continue
			}
			start, ok := e.hwMarks.Get(obj.ref.Value, metricName)
			if !ok {
				start = latest.Add(time.Duration(-res.sampling) * time.Second * (time.Duration(e.Parent.MetricLookback) - 1))
			}
			start = start.Truncate(20 * time.Second) // Truncate to maximum resolution

			// Create bucket if we don't already have it
			bucket, ok := timeBuckets[start.Unix()]
			if !ok {
				bucket = &types.PerfQuerySpec{
					Entity:     obj.ref,
					MaxSample:  maxSampleConst,
					MetricId:   make([]types.PerfMetricId, 0),
					IntervalId: res.sampling,
					Format:     "normal",
				}
				bucket.StartTime = &start
				bucket.EndTime = &now
				timeBuckets[start.Unix()] = bucket
			}

			// Add this metric to the bucket
			bucket.MetricId = append(bucket.MetricId, metric)

			// Bucket filled to capacity?
			// OR if we're past the absolute maximum limit
			if (!res.realTime && len(bucket.MetricId) >= maxMetrics) || len(bucket.MetricId) > maxRealtimeMetrics {
				e.log.Debugf("Submitting partial query: %d metrics (%d remaining) of type %s for %s. Total objects %d",
					len(bucket.MetricId), len(res.metrics)-metricIdx, res.name, e.URL.Host, len(res.objects))

				// Don't send work items if the context has been cancelled.
				if ctx.Err() == context.Canceled {
					return
				}

				// Run collection job
				delete(timeBuckets, start.Unix())
				submitChunkJob(ctx, te, job, queryChunk{*bucket})
			}
		}
		// Handle data in time bucket and submit job if we've reached the maximum number of object.
		for _, bucket := range timeBuckets {
			pqs = append(pqs, *bucket)
			numQs += len(bucket.MetricId)
			if (!res.realTime && numQs > e.Parent.MaxQueryObjects) || numQs > maxRealtimeMetrics {
				e.log.Debugf("Submitting final bucket job for %s: %d metrics", res.name, numQs)
				submitChunkJob(ctx, te, job, pqs)
				pqs = make(queryChunk, 0, e.Parent.MaxQueryObjects)
				numQs = 0
			}
		}
	}
	// Submit any jobs left in the queue
	if len(pqs) > 0 {
		e.log.Debugf("Submitting job for %s: %d objects, %d metrics", res.name, len(pqs), numQs)
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
	estInterval := time.Minute
	if !res.lastColl.IsZero() {
		s := time.Duration(res.sampling) * time.Second
		rawInterval := localNow.Sub(res.lastColl)
		paddedInterval := rawInterval + time.Duration(res.sampling/2)*time.Second
		estInterval = paddedInterval.Truncate(s)
		if estInterval < s {
			estInterval = s
		}
		e.log.Debugf("Raw interval %s, padded: %s, estimated: %s", rawInterval, paddedInterval, estInterval)
	}
	e.log.Debugf("Interval estimated to %s", estInterval)
	res.lastColl = localNow

	latest := res.latestSample
	if !latest.IsZero() {
		elapsed := now.Sub(latest).Seconds() + 5.0 // Allow 5 second jitter.
		e.log.Debugf("Latest: %s, elapsed: %f, resource: %s", latest, elapsed, resourceType)
		if !res.realTime && elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			e.log.Debugf("Sampling period for %s of %d has not elapsed on %s",
				resourceType, res.sampling, e.URL.Host)
			return nil
		}
	} else {
		latest = now.Add(time.Duration(-res.sampling) * time.Second)
	}

	internalTags := map[string]string{"resourcetype": resourceType}
	sw := NewStopwatchWithTags("gather_duration", e.URL.Host, internalTags)

	e.log.Debugf("Collecting metrics for %d objects of type %s for %s",
		len(res.objects), resourceType, e.URL.Host)

	count := int64(0)

	var tsMux sync.Mutex
	latestSample := time.Time{}

	// Divide workload into chunks and process them concurrently
	e.chunkify(ctx, res, now, latest,
		func(chunk queryChunk) {
			n, localLatest, err := e.collectChunk(ctx, chunk, res, acc, estInterval)
			e.log.Debugf("CollectChunk for %s returned %d metrics", resourceType, n)
			if err != nil {
				acc.AddError(errors.New("while collecting " + res.name + ": " + err.Error()))
				return
			}
			e.Parent.Log.Debugf("CollectChunk for %s returned %d metrics", resourceType, n)
			atomic.AddInt64(&count, int64(n))
			tsMux.Lock()
			defer tsMux.Unlock()
			if localLatest.After(latestSample) && !localLatest.IsZero() {
				latestSample = localLatest
			}
		})

	e.log.Debugf("Latest sample for %s set to %s", resourceType, latestSample)
	if !latestSample.IsZero() {
		res.latestSample = latestSample
	}
	sw.Stop()
	SendInternalCounterWithTags("gather_count", e.URL.Host, internalTags, count)
	return nil
}

func (e *Endpoint) alignSamples(info []types.PerfSampleInfo, values []int64, interval time.Duration) ([]types.PerfSampleInfo, []float64) {
	rInfo := make([]types.PerfSampleInfo, 0, len(info))
	rValues := make([]float64, 0, len(values))
	bi := 1.0
	var lastBucket time.Time
	for idx := range info {
		// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
		// data coming back with missing values. Take care of that gracefully!
		if idx >= len(values) {
			e.log.Debugf("len(SampleInfo)>len(Value) %d > %d during alignment", len(info), len(values))
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
			rValues[p] = ((bi-1)/bi)*rValues[p] + v/bi
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
	return rInfo, rValues
}

func (e *Endpoint) collectChunk(ctx context.Context, pqs queryChunk, res *resourceKind, acc telegraf.Accumulator, interval time.Duration) (int, time.Time, error) {
	e.log.Debugf("Query for %s has %d QuerySpecs", res.name, len(pqs))
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

	e.log.Debugf("Query for %s returned metrics for %d objects", resourceType, len(ems))

	// Iterate through results
	for _, em := range ems {
		moid := em.Entity.Reference().Value
		instInfo, found := res.objects[moid]
		if !found {
			e.log.Errorf("MOID %s not found in cache. Skipping! (This should not happen!)", moid)
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
				e.log.Errorf("MOID %s not found in cache. Skipping", moid)
				continue
			}
			e.populateTags(objectRef, resourceType, res, t, &v)

			nValues := 0
			alignedInfo, alignedValues := e.alignSamples(em.SampleInfo, v.Value, interval)

			for idx, sample := range alignedInfo {
				// According to the docs, SampleInfo and Value should have the same length, but we've seen corrupted
				// data coming back with missing values. Take care of that gracefully!
				if idx >= len(alignedValues) {
					e.log.Debugf("Len(SampleInfo)>len(Value) %d > %d", len(alignedInfo), len(alignedValues))
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
					e.log.Errorf("Could not determine unit for %s. Skipping", name)
				}
				v := alignedValues[idx]
				if info.UnitInfo.GetElementDescription().Key == "percent" {
					bucket.fields[fn] = v / 100.0
				} else {
					if e.Parent.UseIntSamples {
						bucket.fields[fn] = int64(round(v))
					} else {
						bucket.fields[fn] = v
					}
				}
				count++

				// Update hiwater marks
				e.hwMarks.Put(moid, name, ts)
			}
			if nValues == 0 {
				e.log.Debugf("Missing value for: %s, %s", name, objectRef.name)
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
			if gh := objectRef.lookup["guesthostname"]; gh != "" {
				t["guesthostname"] = gh
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

		// Add IP addresses to NIC data.
		if resourceType == "vm" && objectRef.lookup != nil {
			key := "nic/" + t["interface"] + "/"
			if ip, ok := objectRef.lookup[key+"ipv6"]; ok {
				t["ipv6"] = ip
			}
			if ip, ok := objectRef.lookup[key+"ipv4"]; ok {
				t["ipv4"] = ip
			}
		}
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

	// Fill in custom values if they exist
	if objectRef.customValues != nil {
		for k, v := range objectRef.customValues {
			if v != "" {
				t[k] = v
			}
		}
	}
}

func (e *Endpoint) makeMetricIdentifier(prefix, metric string) (metricName string, fieldName string) {
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
