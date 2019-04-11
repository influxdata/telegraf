package vsphere

import (
	"context"
	"encoding/json"
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
	Namespace                = "vsan"
	Path                     = "/vsanHealth"
	firstDuration            = 300
	vsanPerfMetricsName      = "vsphere_cluster_vsan_performance"
	vsanHealthfMetricsName   = "vsphere_cluster_vsan_health"
	vsanCapacityfMetricsName = "vsphere_cluster_vsan_capacity"
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

func (e *Endpoint) collectVSan(ctx context.Context, resourceType string, acc telegraf.Accumulator) error {
	if !VersionSupportsVsan(e.apiVersion) {
		log.Printf("I! [inputs.vsan]: Minimum API Version 5.5 required for vSAN. Found: %s. Skipping VCenter: %s", e.apiVersion, e.URL.Host)
		return nil
	}
	res := e.resourceKinds[resourceType]
	var wg sync.WaitGroup

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}
	metrics := e.getVSanPerfMetadata(ctx, client.Client.Client, res)
	for _, obj := range res.objects {
		client, err = e.clientFactory.GetClient(ctx)
		if err != nil {
			log.Printf("D! [inputs.vsan]: Failed to get client: %s", err)
			return err
		}
		wg.Add(1)

		go func(ctx context.Context, obj objectRef, vimClient *vim25.Client, acc telegraf.Accumulator) {
			defer wg.Done()
			e.collectVSanPerCluster(ctx, obj, vimClient, metrics, acc)
		}(ctx, obj, client.Client.Client, acc)
	}
	return nil
}

func (e *Endpoint) collectVSanPerCluster(ctx context.Context, clusterRef objectRef, client *vim25.Client, metrics []string, acc telegraf.Accumulator) {
	cluster := object.NewClusterComputeResource(client, clusterRef.ref)
	vsanClient := client.NewServiceClient(Path, Namespace)
	cmmds, err := getCMMDSMap(ctx, client, cluster)
	if err != nil {
		log.Printf("E! [inputs.vsan]: Error while query cmmds data. Error: %s", err)
		cmmds = make(map[string]CmmdsEntity)
	}
	if err = e.queryDiskUsage(ctx, vsanClient, clusterRef, acc); err != nil {
		acc.AddError(err)
	}
	if err = e.queryHealthSummary(ctx, vsanClient, clusterRef, acc); err != nil {
		acc.AddError(err)
	}
	if len(metrics) > 0 {
		if err = e.queryPerfData(ctx, vsanClient, clusterRef, metrics, cmmds, acc); err != nil {
			acc.AddError(err)
		}
	}
}

func (e *Endpoint) getVSanPerfMetadata(ctx context.Context, client *vim25.Client, res *resourceKind) []string {
	soapClient := client.NewServiceClient(Path, Namespace)
	entityRes, err := vsanmethods.VsanPerfGetSupportedEntityTypes(ctx, soapClient,
		&vsantypes.VsanPerfGetSupportedEntityTypes{
			This: perfManagerRef,
		})
	if err != nil {
		log.Fatal(err)
	}
	var metrics []string

	for _, entity := range entityRes.Returnval {
		if res.filters.Match(entity.Name) {
			metrics = append(metrics, entity.Name)
		}
	}
	metrics = append(metrics, "lsom-world-cpu", "dom-world-cpu")
	log.Println("D! vSan Metric:", metrics)
	return metrics
}

func getCMMDSMap(ctx context.Context, client *vim25.Client, clusterObj *object.ClusterComputeResource) (map[string]CmmdsEntity, error) {
	hosts, err := clusterObj.Hosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to get host: %v", err)
	}

	if len(hosts) == 0 {
		log.Println("I! No host in cluster: ", clusterObj.Name())
		return make(map[string]CmmdsEntity), nil
	}

	vis, err := hosts[0].ConfigManager().VsanInternalSystem(ctx)
	if err != nil {
		return nil, fmt.Errorf("fail to get VsanInternalSystem: %v", err)
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

	request := types.QueryCmmds{
		This:    vis.Reference(),
		Queries: queries,
	}
	res, err := methods.QueryCmmds(ctx, client.RoundTripper, &request)
	if err != nil {
		return nil, fmt.Errorf("fail to query cmmds: %v", err)
	}
	var clusterCmmds Cmmds

	err = json.Unmarshal([]byte(res.Returnval), &clusterCmmds)
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

func (e *Endpoint) queryPerfData(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, metrics []string, cmmds map[string]CmmdsEntity, acc telegraf.Accumulator) error {
	for _, entityRefId := range metrics {
		start, ok := e.hwMarks.Get(entityRefId)
		if !ok {
			start = time.Now().Add(time.Duration(-firstDuration) * time.Second)
		}
		log.Printf("D! [inputs.vsan]: Query Start Time : %s", start)

		var perfSpecs []vsantypes.VsanPerfQuerySpec

		end := time.Now()
		perfSpec := vsantypes.VsanPerfQuerySpec{
			EntityRefId: fmt.Sprintf("%s:*", entityRefId),
			StartTime:   &start,
			EndTime:     &end,
		}
		perfSpecs = append(perfSpecs, perfSpec)

		perfRequest := vsantypes.VsanPerfQueryPerf{
			This:       perfManagerRef,
			QuerySpecs: perfSpecs,
			Cluster:    &vsantypes.ManagedObjectReference{clusterRef.ref.Type, clusterRef.ref.Value},
		}
		resp, err := vsanmethods.VsanPerfQueryPerf(ctx, vsanClient, &perfRequest)

		if err != nil {
			log.Printf("E! [inputs.vsan]: Error while query performance data. Is vsan performace enabled? Error: %s", err)
			continue

		}
		tags := PopulateClusterTags(make(map[string]string), clusterRef, e.URL.Host)

		for _, em := range resp.Returnval {
			log.Printf("D! [inputs.vsphere][vSAN]\tSuccessfully Fetched data for Entity ==> %s:%d\n", em.EntityRefId, len(em.Value))
			vals := strings.Split(em.EntityRefId, ":")
			entityName, uuid := vals[0], vals[1]
			tags := PopulateCMMDSTags(tags, entityName, uuid, cmmds)
			var timeStamps []string
			for _, t := range strings.Split(em.SampleInfo, ",") {
				tsParts := strings.Split(t, " ")
				if len(tsParts) >= 2 {
					timeStamps = append(timeStamps, fmt.Sprintf("%sT%sZ", tsParts[0], tsParts[1]))
				}
			}
			for _, counter := range em.Value {
				metricLabel := counter.MetricId.Label
				for i, values := range strings.Split(counter.Values, ",") {
					ts, ok := time.Parse(time.RFC3339, timeStamps[i])
					if ok != nil {
						// can't do much if we couldn't parse time
						log.Printf("E! [inputs.vsphere][vSAN]Failed to parse a timestamp: %s", timeStamps[i])
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
		}
	}
	return nil
}

func (e *Endpoint) queryDiskUsage(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, acc telegraf.Accumulator) error {
	resp, err := vsanmethods.VsanQuerySpaceUsage(ctx, vsanClient,
		&vsantypes.VsanQuerySpaceUsage{
			This:    spaceManagerRef,
			Cluster: vsantypes.ManagedObjectReference{clusterRef.ref.Type, clusterRef.ref.Value},
		})
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields["FreeCapacityB"] = resp.Returnval.FreeCapacityB
	fields["TotalCapacityB"] = resp.Returnval.TotalCapacityB
	tags := PopulateClusterTags(make(map[string]string), clusterRef, e.URL.Host)
	acc.AddFields(vsanCapacityfMetricsName, fields, tags)
	return nil
}

func (e *Endpoint) queryHealthSummary(ctx context.Context, vsanClient *soap.Client, clusterRef objectRef, acc telegraf.Accumulator) error {
	fetchFromCache := true
	resp, err := vsanmethods.VsanQueryVcClusterHealthSummary(ctx, vsanClient,
		&vsantypes.VsanQueryVcClusterHealthSummary{
			This:           healthSystemRef,
			Cluster:        vsantypes.ManagedObjectReference{clusterRef.ref.Type, clusterRef.ref.Value},
			Fields:         []string{"overallHealth", "overallHealthDescription"},
			FetchFromCache: &fetchFromCache,
		})
	if err != nil {
		return err
	}
	fields := make(map[string]interface{})
	fields["OverallHealth"] = resp.Returnval.OverallHealth
	tags := PopulateClusterTags(make(map[string]string), clusterRef, e.URL.Host)
	acc.AddFields(vsanHealthfMetricsName, fields, tags)
	return nil
}

func PopulateClusterTags(tags map[string]string, clusterRef objectRef, vcenter string) map[string]string {
	newTags := make(map[string]string)
	//deep copy
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

func PopulateCMMDSTags(tags map[string]string, entityName string, uuid string, cmmds map[string]CmmdsEntity) map[string]string {
	newTags := make(map[string]string)
	//deep copy
	for k, v := range tags {
		newTags[k] = v
	}
	//Add additional tags based on CMMDS data
	if strings.Contains(entityName, "-disk") {
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

func VersionSupportsVsan(version string) bool {
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
