package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
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
	Namespace     = "vsan"
	Path          = "/vsanHealth"
	firstDuration = 300
)

func (e *Endpoint) collectVSan(ctx context.Context, resourceType string, acc telegraf.Accumulator) error {
	if !VersionSupportsVsan(e.apiVersion) {
		log.Printf("I! [inputs.vsan]: Minimum API Version 5.5 required for vSAN. Found: %.1f. Skipping VCenter: %s", e.apiVersion, e.URL.Host)
		return nil
	}
	res := e.resourceKinds[resourceType]
	var wg sync.WaitGroup

	client, err := e.clientFactory.GetClient(ctx)
	if err != nil {
		return err
	}
	metrics := e.getVSanMetadata(ctx, client.Client.Client, res)
	for _, obj := range res.objects {
		client, err = e.clientFactory.GetClient(ctx)
		if err != nil {
			log.Printf("D! [inputs.vsan]: Failed to get client: %s", err)
			return err
		}
		wg.Add(1)

		go func(ctx context.Context, obj objectRef, vimClient *vim25.Client, acc telegraf.Accumulator) {
			defer wg.Done()
			e.CollectVSAN(ctx, obj, vimClient, metrics, acc)
		}(ctx, obj, client.Client.Client, acc)
	}
	return nil
}

func (e *Endpoint) CollectVSAN(ctx context.Context, clusterRef objectRef, client *vim25.Client, metrics []string, acc telegraf.Accumulator) {

	cluster := object.NewClusterComputeResource(client, clusterRef.ref)
	soapClient := client.NewServiceClient(Path, Namespace)
	cmmds, err := QueryCmmds(ctx, client, cluster)
	if err != nil {
		log.Printf("E! [inputs.vsan]: Error while query cmmds data. Error: %s", err)
		cmmds = make(map[string]CmmdsEntity)
	}

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

		perfManager := vsantypes.ManagedObjectReference{
			Type:  "VsanPerformanceManager",
			Value: "vsan-performance-manager",
		}

		perfRequest := vsantypes.VsanPerfQueryPerf{
			This:       perfManager,
			QuerySpecs: perfSpecs,
			Cluster:    &vsantypes.ManagedObjectReference{cluster.Reference().Type, cluster.Reference().Value},
		}
		perfRes, err := vsanmethods.VsanPerfQueryPerf(ctx, soapClient, &perfRequest)

		if err != nil {
			log.Printf("E! [inputs.vsan]: Error while query performance data. Is vsan performace enabled? Error: %s", err)
			continue
		}

		for _, em := range perfRes.Returnval {
			tags := PopulateClusterTags(clusterRef, e.URL.Host)
			buckets := make(map[string]metricEntry)
			log.Printf("D! [inputs.vsan]\tSuccessfully Fetched data for Entity ==> %s:%d\n", em.EntityRefId, len(em.Value))
			timestamps := strings.Split(em.SampleInfo, ",")
			log.Printf("D! [inputs.vsan]: Time Stamp : %s", em.SampleInfo)

			vals := strings.Split(em.EntityRefId, ":") //host-domclient:5ca25228-f047-558e-2b73-02001491d8eb
			entityName, uuid := vals[0], vals[1]

			for _, value := range em.Value {
				metricName := value.MetricId.Label
				tags = PopulateCMMDSTags(tags, entityName, uuid, cmmds)

				// Now deal with the values. Iterate backwards so we start with the latest value
				valuesSlice := strings.Split(value.Values, ",")
				for idx := len(valuesSlice) - 1; idx >= 0; idx-- {
					ts, _ := time.Parse("2006-01-02 15:04:05", timestamps[idx])

					// Since non-realtime metrics are queries with a lookback, we need to check the high-water mark
					// to determine if this should be included. Only samples not seen before should be included.

					value, _ := strconv.ParseFloat(valuesSlice[idx], 64)

					// Organize the metrics into a bucket per measurement.
					// For now each measurement has one field, so measurement is equal to field label
					measurement := fmt.Sprintf("vsan-%s", metricName)
					field := fmt.Sprintf("vsan-%s", metricName)
					bKey := measurement + " " + strconv.FormatInt(ts.UnixNano(), 10) //bucket key
					bucket, found := buckets[bKey]
					if !found {
						bucket = metricEntry{name: measurement, ts: ts, fields: make(map[string]interface{}), tags: tags}
						buckets[bKey] = bucket
					}
					bucket.fields[field] = value
				}
			}

			// Update highwater marks
			if lens := len(timestamps); lens > 0 {
				latest, _ := time.Parse("2006-01-02 15:04:05", timestamps[lens-1])
				e.hwMarks.Put(entityRefId, latest)
			}

			for _, bucket := range buckets {
				acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
			}
		}
	}
}

func (e *Endpoint) getVSanMetadata(ctx context.Context, client *vim25.Client, res *resourceKind) []string {
	perfManager := vsantypes.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}
	soapClient := client.NewServiceClient(Path, Namespace)
	entityRes, err := vsanmethods.VsanPerfGetSupportedEntityTypes(ctx, soapClient,
		&vsantypes.VsanPerfGetSupportedEntityTypes{
			This: perfManager,
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
	log.Println("D! vSan Metric:", metrics)
	return metrics
}

func QueryCmmds(ctx context.Context, client *vim25.Client, clusterObj *object.ClusterComputeResource) (map[string]CmmdsEntity, error) {

	hosts, err := clusterObj.Hosts(ctx)
	if err != nil {
		log.Println("E! Error happen when get hosts: ", err)
		return nil, err
	}

	if len(hosts) == 0 {
		log.Println("I! No host in cluster: ", clusterObj.Name())
		return make(map[string]CmmdsEntity), nil
	}

	vis, err2 := hosts[0].ConfigManager().VsanInternalSystem(ctx)
	if err2 != nil {
		log.Println("E! Error happen when get VsanInternalSystem: ", err)
		return nil, err2
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
		log.Println("E! Query cmmds error: ", err)
		return nil, err
	}
	var clusterCmmds Cmmds

	err = json.Unmarshal([]byte(res.Returnval), &clusterCmmds)
	if err != nil {
		log.Println("E! Error when turning to json : ", err)
		return nil, err
	}

	cmmdsMap := make(map[string]CmmdsEntity)
	for _, entity := range clusterCmmds.Res {
		uuid := entity.UUID
		cmmdsMap[uuid] = entity
	}
	return cmmdsMap, nil
}

func PopulateClusterTags(clusterRef objectRef, vcenter string) map[string]string {
	tags := make(map[string]string)
	tags["vcenter"] = vcenter
	tags["dcname"] = clusterRef.dcname
	tags["clustername"] = clusterRef.name
	tags["moid"] = clusterRef.ref.Value
	tags["source"] = clusterRef.name
	return tags
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
