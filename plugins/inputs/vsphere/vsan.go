package vsphere

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"

	vsanmethods "github.com/influxdata/telegraf/plugins/inputs/vsphere/vsan-sdk/methods"
	vsantypes "github.com/influxdata/telegraf/plugins/inputs/vsphere/vsan-sdk/types"

	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	vsanNamespace           = "vsan"
	vsanPath                = "/vsanHealth"
	hwMarksKey              = "vsan-perf"
	vsanPerfMetricsName     = "vsphere_cluster_vsan_performance"
	vsanHealthMetricsName   = "vsphere_cluster_vsan_health"
	vsanCapacityMetricsName = "vsphere_cluster_vsan_capacity"
)

var (
	perfManagerRef = vsantypes.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}
	healthSystemRef = vsantypes.ManagedObjectReference{
		Type:  "VsanVcClusterHealthSystem",
		Value: "vsan-cluster-health-system",
	}
	spaceManagerRef = vsantypes.ManagedObjectReference{
		Type:  "VsanSpaceReportSystem",
		Value: "vsan-cluster-space-report-system",
	}
)

// collectVsan is the entry point for vsan metrics collection
func (e *Endpoint) collectVsan(ctx context.Context, resourceType string, acc telegraf.Accumulator) error {
	if !versionSupportsVsan(e.apiVersion) {
		log.Printf("I! [inputs.vsphere][vSAN] Minimum API Version 5.5 required for vSAN. Found: %s. Skipping VCenter: %s", e.apiVersion, e.URL.Host)
		return nil
	}
	res := e.resourceKinds[resourceType]
	var waitGroup sync.WaitGroup

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("fail to get client when collect vsan: %v", err)
	}
	vimClient := client.Client.Client
	metrics := e.getVsanPerfMetadata(ctx, vimClient, res)
	// Iterate over all clusters, run a goroutine for each cluster
	for _, obj := range res.objects {
		waitGroup.Add(1)
		go func(clusterObj objectRef) {
			defer waitGroup.Done()
			e.collectVsanPerCluster(ctx, clusterObj, vimClient, metrics, acc)
		}(obj)
	}
	return nil
}

// collectVsanPerCluster is called by goroutines in collectVsan function.
func (e *Endpoint) collectVsanPerCluster(ctx context.Context, clusterRef objectRef, client *vim25.Client, metrics []string, acc telegraf.Accumulator) {
	// 1. Construct a map for cmmds
	cluster := object.NewClusterComputeResource(client, clusterRef.ref)
	cmmds, err := getCmmdsMap(ctx, client, cluster)
	if err != nil {
		log.Printf("E! [inputs.vsphere][vSAN] Error while query cmmds data. Error: %s. Skipping", err)
		cmmds = make(map[string]CmmdsEntity)
	}
	// 2. Create a vsan client
	vsanClient := client.NewServiceClient(vsanPath, vsanNamespace)
	// 3. Do collection
	if err = e.queryDiskUsage(ctx, vsanClient, clusterRef, acc); err != nil {
		acc.AddError(errors.New("While querying vsan disk usage:" + err.Error()))
	}
	if err = e.queryHealthSummary(ctx, vsanClient, clusterRef, acc); err != nil {
		acc.AddError(errors.New("While querying vsan health summary:" + err.Error()))
	}
	if len(metrics) > 0 {
		if err = e.queryPerformance(ctx, vsanClient, clusterRef, metrics, cmmds, acc); err != nil {
			acc.AddError(errors.New("While query vsan perf data:" + err.Error()))
		}
	}
}

// getVsanPerfMetadata returns a string list of the performance entity types that will be queried.
func (e *Endpoint) getVsanPerfMetadata(ctx context.Context, client *vim25.Client, res *resourceKind) []string {
	vsanClient := client.NewServiceClient(vsanPath, vsanNamespace)
	entityRes, err := vsanmethods.VsanPerfGetSupportedEntityTypes(ctx, vsanClient,
		&vsantypes.VsanPerfGetSupportedEntityTypes{
			This: perfManagerRef,
		})
	var metrics []string

	if err != nil {
		log.Printf("E! [inputs.vsphere][vSAN] Fail to get supported entities: %v. Skipping vsan performance data.", err)
		return metrics
	}
	// Use the include & exclude configuration to filter all supported metrics
	for _, entity := range entityRes.Returnval {
		if res.filters.Match(entity.Name) {
			metrics = append(metrics, entity.Name)
		}
	}
	metrics = append(metrics)
	log.Printf("D! [inputs.vsphere][vSAN]\tvSan performance Metric: %v", metrics)
	return metrics
}

// getCmmdsMap returns a map which maps a uuid to a CmmdsEntity
func getCmmdsMap(ctx context.Context, client *vim25.Client, clusterObj *object.ClusterComputeResource) (map[string]CmmdsEntity, error) {
	hosts, err := clusterObj.Hosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to get host: %v", err)
	}

	if len(hosts) == 0 {
		log.Printf("I! [inputs.vsphere][vSAN]\tNo host in cluster: %s", clusterObj.Name())
		return make(map[string]CmmdsEntity), nil
	}

	queries := make([]types.HostVsanInternalSystemCmmdsQuery, 2)
	hostnameCmmdsQuery := types.HostVsanInternalSystemCmmdsQuery{
		Type: "HOSTNAME",
	}
	diskCmmdsQuery := types.HostVsanInternalSystemCmmdsQuery{
		Type: "DISK",
	}
	queries = append(queries, hostnameCmmdsQuery)
	queries = append(queries, diskCmmdsQuery)

	// We will iterate cmmds querying on each host until a success.
	// It happens that some hosts return successfully while others fail.
	var resp *types.QueryCmmdsResponse
	for _, host := range hosts {
		vis, err := host.ConfigManager().VsanInternalSystem(ctx)
		if err != nil {
			log.Printf("I! [inputs.vsphere][vSAN] Fail to get VsanInternalSystem from %s: %s", host.Name(), err)
			continue
		}
		request := types.QueryCmmds{
			This:    vis.Reference(),
			Queries: queries,
		}
		resp, err = methods.QueryCmmds(ctx, client.RoundTripper, &request)
		if err != nil {
			log.Printf("I! [inputs.vsphere][vSAN] Fail to query cmmds from %s: %s", host.Name(), err)
		} else {
			log.Printf("I! [inputs.vsphere][vSAN] Successfully get cmmds from %s", host.Name())
			break
		}
	}
	if resp == nil {
		return nil, fmt.Errorf("all hosts fail to query cmmds")
	}
	var clusterCmmds Cmmds

	err = json.Unmarshal([]byte(resp.Returnval), &clusterCmmds)
	if err != nil {
		return nil, fmt.Errorf("fail to convert cmmds to json: %v", err)
	}

	cmmdsMap := make(map[string]CmmdsEntity)
	for _, entity := range clusterCmmds.Res {
		uuid := entity.UUID
		cmmdsMap[uuid] = entity
	}
	return cmmdsMap, nil
}

// queryPerformance adds performance metrics to telegraf accumulator
func (e *Endpoint) queryPerformance(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, metrics []string, cmmds map[string]CmmdsEntity, acc telegraf.Accumulator) error {
	end := time.Now().UTC()
	start, ok := e.hwMarks.Get(hwMarksKey)
	if !ok {
		// Look back 3 sampling periods by default
		start = end.Add(metricLookback * time.Duration(-e.resourceKinds["vsan"].sampling) * time.Second)
	}
	log.Printf("D! [inputs.vsphere][vSAN]\tQuery vsan performance for time interval: %s ~ %s", start, end)
	latest := start

	for _, entityRefId := range metrics {
		var perfSpecs []vsantypes.VsanPerfQuerySpec

		perfSpec := vsantypes.VsanPerfQuerySpec{
			EntityRefId: fmt.Sprintf("%s:*", entityRefId),
			StartTime:   &start,
			EndTime:     &end,
		}
		perfSpecs = append(perfSpecs, perfSpec)

		perfRequest := vsantypes.VsanPerfQueryPerf{
			This:       perfManagerRef,
			QuerySpecs: perfSpecs,
			Cluster:    &vsantypes.ManagedObjectReference{Type: clusterRef.ref.Type, Value: clusterRef.ref.Value},
		}
		resp, err := vsanmethods.VsanPerfQueryPerf(ctx, vsanClient, &perfRequest)

		if err != nil {
			log.Printf("E! [inputs.vsphere][vSAN] Error querying performance data for %s: %s: %s. Is vsan performace enabled?", clusterRef.name, entityRefId, err)
			continue

		}
		tags := populateClusterTags(make(map[string]string), clusterRef, e.URL.Host)

		count := 0
		for _, em := range resp.Returnval {
			count += len(em.Value)
			vals := strings.Split(em.EntityRefId, ":")
			entityName, uuid := vals[0], vals[1]
			tags := populateCMMDSTags(tags, entityName, uuid, cmmds)
			var timeStamps []string
			// 1. Construct a timestamp list from sample info
			for _, t := range strings.Split(em.SampleInfo, ",") {
				tsParts := strings.Split(t, " ")
				if len(tsParts) >= 2 {
					// The return time string is in UTC time
					timeStamps = append(timeStamps, fmt.Sprintf("%sT%sZ", tsParts[0], tsParts[1]))
				}
			}
			// 2. Iterate on each measurement
			for _, counter := range em.Value {
				metricLabel := counter.MetricId.Label
				// 3. Iterate on each data point.
				// For each data point, we attach the corresponding timestamp and add it to accumulator
				for i, values := range strings.Split(counter.Values, ",") {
					ts, ok := time.Parse(time.RFC3339, timeStamps[i])
					if ok != nil {
						log.Printf("E! [inputs.vsphere][vSAN]\tFailed to parse a timestamp: %s. Skipping", timeStamps[i])
						continue
					}
					fields := make(map[string]interface{})
					field := fmt.Sprintf("%s_%s", entityName, metricLabel)
					if v, err := strconv.ParseFloat(values, 32); err == nil {
						fields[field] = v
					}
					acc.AddFields(vsanPerfMetricsName, fields, tags, ts)
				}
			}
			if len(timeStamps) > 0 {
				lastSample, err := time.Parse(time.RFC3339, timeStamps[len(timeStamps)-1])
				if err == nil && lastSample.After(latest) {
					latest = lastSample
				}
			}
		}
		log.Printf("I! [inputs.vsphere][vSAN] Successfully Fetched data for Entity ==> %s:%s:%d\n", clusterRef.name, entityRefId, count)
	}
	e.hwMarks.Put(hwMarksKey, latest)
	return nil
}

// queryDiskUsage adds 'FreeCapacityB' and 'TotalCapacityB' metrics to telegraf accumulator
func (e *Endpoint) queryDiskUsage(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, acc telegraf.Accumulator) error {
	resp, err := vsanmethods.VsanQuerySpaceUsage(ctx, vsanClient,
		&vsantypes.VsanQuerySpaceUsage{
			This:    spaceManagerRef,
			Cluster: vsantypes.ManagedObjectReference{Type: clusterRef.ref.Type, Value: clusterRef.ref.Value},
		})
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields["FreeCapacityB"] = resp.Returnval.FreeCapacityB
	fields["TotalCapacityB"] = resp.Returnval.TotalCapacityB
	tags := populateClusterTags(make(map[string]string), clusterRef, e.URL.Host)
	acc.AddFields(vsanCapacityMetricsName, fields, tags)
	return nil
}

// queryDiskUsage adds 'OverallHealth' metric to telegraf accumulator
func (e *Endpoint) queryHealthSummary(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, acc telegraf.Accumulator) error {
	fetchFromCache := true
	resp, err := vsanmethods.VsanQueryVcClusterHealthSummary(ctx, vsanClient,
		&vsantypes.VsanQueryVcClusterHealthSummary{
			This:           healthSystemRef,
			Cluster:        vsantypes.ManagedObjectReference{Type: clusterRef.ref.Type, Value: clusterRef.ref.Value},
			Fields:         []string{"overallHealth", "overallHealthDescription"},
			FetchFromCache: &fetchFromCache,
		})
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	overallHealth := resp.Returnval.OverallHealth
	switch overallHealth {
	case "red":
		fields["OverallHealth"] = 2
	case "yellow":
		fields["OverallHealth"] = 1
	case "green":
		fields["OverallHealth"] = 0
	default:
		fields["OverallHealth"] = -1
	}
	tags := populateClusterTags(make(map[string]string), clusterRef, e.URL.Host)
	acc.AddFields(vsanHealthMetricsName, fields, tags)
	return nil
}

// populateClusterTags takes in a tag map, makes a copy, populates cluster related tags and returns the copy.
func populateClusterTags(tags map[string]string, clusterRef objectRef, vcenter string) map[string]string {
	newTags := make(map[string]string)
	// deep copy
	for k, v := range tags {
		newTags[k] = v
	}
	newTags["vcenter"] = vcenter
	newTags["dcname"] = clusterRef.dcname
	newTags["clustername"] = clusterRef.name
	newTags["moid"] = clusterRef.ref.Value
	newTags["source"] = clusterRef.name
	return newTags
}

// populateCMMDSTags takes in a tag map, makes a copy, adds more tags using a cmmds map and returns the copy.
func populateCMMDSTags(tags map[string]string, entityName string, uuid string, cmmds map[string]CmmdsEntity) map[string]string {
	newTags := make(map[string]string)
	// deep copy
	for k, v := range tags {
		newTags[k] = v
	}
	// Add additional tags based on CMMDS data
	if strings.Contains(entityName, "-disk") || strings.Contains(entityName, "disk-") {
		if e, ok := cmmds[uuid]; ok {
			if host, ok := cmmds[e.Owner]; ok {
				if c, ok := host.Content.(map[string]interface{}); ok {
					newTags["hostname"] = c["hostname"].(string)
				}
			}
			if c, ok := e.Content.(map[string]interface{}); ok {
				newTags["deviceName"] = c["devName"].(string)
				if int(c["isSsd"].(float64)) == 0 {
					newTags["ssdUuid"] = c["ssdUuid"].(string)
				}
			}
		}
	} else if strings.Contains(entityName, "host-") {
		if e, ok := cmmds[uuid]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				newTags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "vnic-net") {
		nicInfo := strings.Split(uuid, "|")
		newTags["stackName"] = nicInfo[1]
		newTags["vnic"] = nicInfo[2]
		if e, ok := cmmds[nicInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				newTags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "pnic-net") {
		nicInfo := strings.Split(uuid, "|")
		newTags["pnic"] = nicInfo[1]
		if e, ok := cmmds[nicInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				newTags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "world-cpu") {
		cpuInfo := strings.Split(uuid, "|")
		newTags["worldName"] = cpuInfo[1]
		//newTags["worldId"] = cpuInfo[2]
		if e, ok := cmmds[cpuInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				newTags["hostname"] = c["hostname"].(string)
			}
		}
	} else {
		newTags["uuid"] = uuid
	}
	return newTags
}

// versionSupportsVsan returns if vsan is supported for a given version, that is version >= 5.5.
func versionSupportsVsan(version string) bool {
	v := strings.Split(version, ".")
	major, err := strconv.Atoi(v[0])
	if err != nil {
		log.Printf("E! [inputs.vsphere][vSAN] Failed to parse version: %s", version)
	}
	if major < 5 {
		return false
	}
	minor, err := strconv.Atoi(v[1])
	if err != nil {
		log.Printf("E! [inputs.vsphere][vSAN] Failed to parse version: %s.", version)
	}
	if major == 5 && minor < 5 {
		return false
	}
	return true
}

type CmmdsEntity struct {
	UUID    string      `json:"uuid"`
	Owner   string      `json:"owner"` // ESXi UUID
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type Cmmds struct {
	Res []CmmdsEntity `json:"result"`
}
