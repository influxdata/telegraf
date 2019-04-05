package vsan

import (
	"context"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"strconv"
	"strings"
	"time"

	"log"
)

const FirstDuration = 5 * 300

type VSan struct {
	VCenter  string `toml:"vcenter"`
	Username string
	Password string
	//VSanPerfInclude []string `toml:"vsan_perf_exclude"`
	//VSanPerfExclude []string `toml:"vsan_perf_exclude"`
	client  *Client // a client for VSan
	cancel  context.CancelFunc
	hwMarks *TSCache
}

type metricEntry struct {
	tags   map[string]string
	name   string
	ts     time.Time
	fields map[string]interface{}
}

// Description returns a short textual description of the plugin
func (v *VSan) Description() string {
	return "a vsan plugin"
}

// SampleConfig returns a set of default configuration to be used as a boilerplate when setting up
// Telegraf.
func (v *VSan) SampleConfig() string {
	return ""
}

// Start is called from telegraf core when a plugin is started and allows it to
// perform initialization tasks.
func (v *VSan) Start(acc telegraf.Accumulator) error {
	v.hwMarks = NewTSCache(1 * time.Hour)
	return nil
}

func (v *VSan) Stop() {

}

func (v *VSan) Gather(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	defer cancel()

	var err error
	v.client, err = NewClient(ctx, v.VCenter, v.Username, v.Password)
	if err != nil {
		log.Printf("E! [inputs.vsan]: Error while create a new client. Error: %s", err)
		return err
	}

	entityRefId := "host-domclient"
	startTime, ok := v.hwMarks.Get(entityRefId)
	if !ok {
		startTime = time.Now().Add(time.Duration(-FirstDuration) * time.Second)
	}
	log.Printf("D! [inputs.vsan]: Query Start Time : %s", startTime)

	res, err := v.client.QueryPerf(ctx, startTime, entityRefId)

	if err != nil {
		log.Printf("E! [inputs.vsan]: Error while query performance data. Please check vsan performace is enabled. Error: %s", err)
		return err
	}

	cmmds, err := v.client.QueryCmmds(ctx)
	if err != nil {
		log.Printf("E! [inputs.vsan]: Error while query cmmds data. Error: %s", err)
		return err //todo: we don't want shut down because cmmds are not collected
	}

	for _, em := range res.Returnval {
		buckets := make(map[string]metricEntry)
		log.Printf("D! [inputs.vsan]\tSuccessfully Fetched data for Entity ==> %s:%d\n", em.EntityRefId, len(em.Value))
		timestamps := strings.Split(em.SampleInfo, ",")

		vals := strings.Split(em.EntityRefId, ":") //host-domclient:5ca25228-f047-558e-2b73-02001491d8eb
		entityName, uuid := vals[0], vals[1]

		for _, value := range em.Value {
			metricName := value.MetricId.Label
			tags := v.PopulateTags(entityName, uuid, metricName, cmmds)

			// Now deal with the values. Iterate backwards so we start with the latest value
			valuesSlice := strings.Split(value.Values, ",")
			for idx := len(valuesSlice) - 1; idx >= 0; idx-- {
				ts, _ := time.Parse("2006-01-02 15:04:05", timestamps[idx])

				// Since non-realtime metrics are queries with a lookback, we need to check the high-water mark
				// to determine if this should be included. Only samples not seen before should be included.

				value, _ := strconv.ParseFloat(valuesSlice[idx], 64)
				log.Printf("D! [inputs.vsan]: Time Stamp : %s", ts)

				// Organize the metrics into a bucket per measurement.
				// For now each measurement has one field, so measurement is equal to field label
				measurement := "vsan-" + metricName
				field := "vsan-" + metricName
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
			v.hwMarks.Put(entityRefId, latest)
		}

		for _, bucket := range buckets {
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}

	return nil

}

func (v *VSan) PopulateTags(entityName string, uuid string, metricName string, cmmds map[string]CmmdsEntity) map[string]string {
	tags := make(map[string]string)
	tags["vcenter"] = v.VCenter

	//Add additional tags based on CMMDS data
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
		//tags["worldId"] = cpuInfo[2]
		if e, ok := cmmds[cpuInfo[0]]; ok {
			if c, ok := e.Content.(map[string]interface{}); ok {
				tags["hostname"] = c["hostname"].(string)
			}
		}
	} else {
		tags["uuid"] = uuid
	}
	return tags
}

func init() {
	inputs.Add("vsan", func() telegraf.Input {
		return &VSan{
			//VSanPerfInclude: []string{"*"},
			//VSanPerfExclude: nil,
		}
	})
}
