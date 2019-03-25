package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmware/govmomi/vim25/soap"

	"github.com/influxdata/telegraf"
	vsanMethods "github.com/influxdata/telegraf/plugins/inputs/vsphere/vsan/methods"
	vsanTypes "github.com/influxdata/telegraf/plugins/inputs/vsphere/vsan/types"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	vsanNamespace   = "vsan"
	vsanPath        = "/vsanHealth"
	timeFormat      = "Mon, 02 Jan 2006 15:04:05 MST"
	vsanMetricsName = "vsphere_cluster_vsan"
)

var (
	vsanPerformanceManagerInstance = types.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}

	vsanPerfEntityRefIds = []string{
		"host-domclient",
		"host-domcompmgr",
		"cache-disk",
		"capacity-disk",
		"vsan-vnic-net",
		"vsan-pnic-net",
		"lsom-world-cpu",
		"dom-world-cpu",
	}
)

func inferTimezoneOffset(ts time.Time) time.Duration {
	// Compare timestamp to UTC from our local clock. Round the difference to the nearest 30 minutes (because India and Newfoundland).
	// This SHOULD be our timezone offset. As far as I can tell, this should handle daylight savings weirdness etc. as well.
	now := time.Now() // TODO: Get server time instead!
	delta := ts.Sub(now)
	ds := delta.Seconds() / 1800.0
	return time.Duration(time.Duration(round(ds)*1800.0) * time.Second)
}

/*
All this cryptic code in formatAndSendVsanMetric is to parse the vsanTypes.VsanPerfEntityMetricCSV type, which has the structure:
{
  "@type": "vim.cluster.VsanPerfEntityMetricCSV",
  "entityRefId": "cluster-domclient:5270dc4d-3594-cc26-b33d-f6be33ddb353",
  "sampleInfo": "2017-06-14 23:10:00,2017-06-14 23:15:00,2017-06-14 23:20:00,2017-06-14 23:25:00,2017-06-14 23:30:00,2017-06-14 23:35:00,2017-06-14 23:40:00,2017-06-14 23:45:00,2017-06-14 23:50:00,2017-06-14 23:55:00,2017-06-15 00:00:00,2017-06-15 00:05:00,2017-06-15 00:10:00",
  "value": [
    {
      "@type": "vim.cluster.VsanPerfMetricSeriesCSV",
      "metricId": {
        "@type": "vim.cluster.VsanPerfMetricId",
        "description": null,
        "group": null,
        "label": "iopsRead",
        "metricsCollectInterval": 300,
        "name": null,
        "rollupType": null,
        "statsType": null
      },
      "threshold": null,
      "values": "1,1,1,1,1,1,1,1,1,1,1,1,1"
    },
		...
		...
	]
}
*/
func formatAndSendVsanMetric(entity vsanTypes.VsanPerfEntityMetricCSV, defaultTags map[string]string, cmmds map[string]vsanTypes.CmmdsEntity, acc telegraf.Accumulator) {
	vals := strings.Split(entity.EntityRefId, ":")
	entityName := vals[0]
	uuid := vals[1]
	tags := make(map[string]string)

	for k, v := range defaultTags {
		tags[k] = v
	}

	// Add some additional tags based on CMMDS data
	if strings.Contains(entityName, "-disk") {
		if e, ok := cmmds[uuid]; ok {
			if host, ok := cmmds[e.Owner]; ok {
				if c, ok := host.Content.(map[string]interface{}); ok {
					tags["hostname"] = c["hostname"].(string)
				}
			}
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["deviceName"] = c["devName"].(string)
				if int(c["isSsd"].(float64)) == 0 {
					tags["ssdUuid"] = c["ssdUuid"].(string)
				}
			}
		}
	} else if strings.Contains(entityName, "host-") {
		if e, ok := cmmds[uuid]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "vnic-net") {
		nicInfo := strings.Split(uuid, "|")
		tags["stackName"] = nicInfo[1]
		tags["vnic"] = nicInfo[2]
		if e, ok := cmmds[nicInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "pnic-net") {
		nicInfo := strings.Split(uuid, "|")
		tags["pnic"] = nicInfo[1]
		if e, ok := cmmds[nicInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["hostname"] = c["hostname"].(string)
			}
		}
	} else if strings.Contains(entityName, "world-cpu") {
		cpuInfo := strings.Split(uuid, "|")
		tags["worldName"] = cpuInfo[1]
		tags["worldId"] = cpuInfo[2]
		if e, ok := cmmds[cpuInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["hostname"] = c["hostname"].(string)
			}
		}
	} else {
		tags["uuid"] = uuid
	}

	var timeStamps []string
	log.Printf("D! [inputs.vsphere] SampleInfo: %s", entity.SampleInfo)
	for _, t := range strings.Split(entity.SampleInfo, ",") {
		tsParts := strings.Split(t, " ")
		timeStamps = append(timeStamps, fmt.Sprintf("%sT%sZ", tsParts[0], tsParts[1]))
	}

	// Workaround for vSAN sending timestamps in local time, rather than UTC (yuck!)
	n := len(timeStamps) - 1
	ts, ok := time.Parse(time.RFC3339, timeStamps[n])
	if ok != nil {
		// can't do much if we couldn't parse time
		log.Printf("E! [inputs.vsphere][vSAN]Failed to parse final timestamp: %s. Bailing out", timeStamps[n])
		return
	}
	tzOffset := -inferTimezoneOffset(ts)

	for _, counter := range entity.Value {
		metricLabel := counter.MetricId.Label
		for i, values := range strings.Split(counter.Values, ",") {
			ts, ok := time.Parse(time.RFC3339, timeStamps[i])
			if ok != nil {
				// can't do much if we couldn't parse time
				log.Printf("E! [inputs.vsphere][vSAN]Failed to parse a timestamp: %s", timeStamps[i])
				continue
			}
			ts = ts.Add(tzOffset)
			fields := make(map[string]interface{})
			field := fmt.Sprintf("%s_%s", entityName, metricLabel)
			if v, err := strconv.ParseFloat(values, 32); err == nil {
				fields[field] = v
			}
			acc.AddFields(vsanMetricsName, fields, tags, ts)
		}
	}
}

func getAllVsanMetrics(ctx context.Context, vsanClient *soap.Client, cluster *object.ClusterComputeResource, tags map[string]string, cmmds map[string]vsanTypes.CmmdsEntity, acc telegraf.Accumulator) {
	endTime := time.Now()
	startTime := endTime.Add(time.Duration(-5) * time.Minute)
	log.Printf("D! [inputs.vsphere][vSAN]Querying data between: %s -> %s", startTime.Format(timeFormat), endTime.Format(timeFormat))
	for _, entityRefID := range vsanPerfEntityRefIds {
		var querySpecs []vsanTypes.VsanPerfQuerySpec

		spec := vsanTypes.VsanPerfQuerySpec{
			EntityRefId: fmt.Sprintf("%s:*", entityRefID),
			StartTime:   &startTime,
			EndTime:     &endTime,
		}
		querySpecs = append(querySpecs, spec)

		vsanPerfQueryPerf := vsanTypes.VsanPerfQueryPerf{
			This:       vsanPerformanceManagerInstance,
			QuerySpecs: querySpecs,
			Cluster:    cluster.Reference(),
		}
		res, err := vsanMethods.VsanPerfQueryPerf(ctx, vsanClient, &vsanPerfQueryPerf)
		if err != nil {
			log.Fatal(err)
		}

		for _, ret := range res.Returnval {
			log.Printf("D! [inputs.vsphere][vSAN]\tSuccessfully Fetched data for Entity ==> %s:%d\n", ret.EntityRefId, len(ret.Value))
			formatAndSendVsanMetric(ret, tags, cmmds, acc)
		}
	}
}

func getVsanTags(cluster objectRef, vcenter string) map[string]string {
	tags := make(map[string]string)
	tags["vcenter"] = vcenter
	tags["dcname"] = cluster.dcname
	tags["clustername"] = cluster.name
	tags["moid"] = cluster.ref.Value
	tags["source"] = cluster.name
	return tags
}

func getClusterCmmdsData(ctx context.Context, client *vim25.Client, cluster *object.ClusterComputeResource) (map[string]vsanTypes.CmmdsEntity, error) {
	cmmds := make(map[string]vsanTypes.CmmdsEntity)
	hosts, err := cluster.Hosts(ctx)

	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no hosts found")
	}

	// TODO: Should we check if the host returned is connected or not!?
	host := hosts[0]
	vis, err := host.ConfigManager().VsanInternalSystem(ctx)
	if err != nil {
		return nil, err
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

	cmmdsQuery := types.QueryCmmds{
		This:    vis.Reference(),
		Queries: queries,
	}

	rawCmmds, err := methods.QueryCmmds(ctx, client.RoundTripper, &cmmdsQuery)
	if err != nil {
		return nil, err
	}
	var clusterCmmds vsanTypes.Cmmds

	json.Unmarshal([]byte(rawCmmds.Returnval), &clusterCmmds)
	for _, entity := range clusterCmmds.Res {
		uuid := entity.UUID
		cmmds[uuid] = entity
	}
	return cmmds, nil
}

// CollectVsan invokes the vSAN Performance Manager on the ClusterComputeResource from the input.
func CollectVsan(ctx context.Context, client *vim25.Client, clusterObj objectRef, wg *sync.WaitGroup, vcenter string, acc telegraf.Accumulator) {
	defer wg.Done()
	cluster := object.NewClusterComputeResource(client, clusterObj.ref)
	if clusterName, err := cluster.ObjectName(ctx); err != nil {
		log.Printf("D! [inputs.vsphere][vSAN] Starting vSAN Collection for %s", clusterName)
	}

	tags := getVsanTags(clusterObj, vcenter)
	log.Printf("D! [inputs.vsphere][vSAN] Tags for vSAN: %s", tags)

	cmmds, err := getClusterCmmdsData(ctx, client, cluster)
	if err != nil {
		log.Printf("I! [inputs.vsphere][vSAN] Failed to get CMMDS Data. Cannot resolve UUIDs.")
	}

	// vSAN Client
	vsanClient := client.NewServiceClient(vsanPath, vsanNamespace)
	getAllVsanMetrics(ctx, vsanClient, cluster, tags, cmmds, acc)
}

// VersionSupportsVsan returns true if the supplied API version supports vSAN (i.e. version <= 5.5)
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
