package vsan

import (
	"context"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/vsan/vsan-sdk/methods"
	vsantypes "github.com/influxdata/telegraf/plugins/inputs/vsan/vsan-sdk/types"
	"os"
	"strconv"
	"strings"
	"time"

	"log"
)

type VSan struct {
	VCenter  string
	Username string
	Password string
	client   *Client // a soap client for VSan
	cancel   context.CancelFunc
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
	return `
  ## Sample config here
`
}

// Start is called from telegraf core when a plugin is started and allows it to
// perform initialization tasks.
func (v *VSan) Start(acc telegraf.Accumulator) error {
	//log.Println("D! [inputs.vsan]: Starting plugin")
	//ctx, cancel := context.WithCancel(context.Background())
	//v.cancel = cancel
	//var err error
	//v.client, err = NewVSANClient(ctx, v.VCenter, v.Username, v.Password)
	//if err != nil {
	//	log.Fatal(err)
	//}
	return nil
}

func (v *VSan) Gather(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())

	v.cancel = cancel
	var err error
	v.client, err = NewVSANClient(ctx, v.VCenter, v.Username, v.Password)
	if err != nil {
		log.Fatal(err)
	}

	c := v.client.Client
	defer cancel()
	var perfSpecs []vsantypes.VsanPerfQuerySpec
	startTime := time.Now()
	perfSpec := vsantypes.VsanPerfQuerySpec{
		EntityRefId: "host-domclient:*",
		StartTime:   &startTime,
	}
	perfSpecs = append(perfSpecs, perfSpec)

	cluster := vsantypes.ManagedObjectReference{
		Type:  "ClusterComputeResource",
		Value: "domain-c8",
	}

	perfManager := vsantypes.ManagedObjectReference{
		Type:  "VsanPerformanceManager",
		Value: "vsan-performance-manager",
	}

	perfRequest := vsantypes.VsanPerfQueryPerf{
		This:       perfManager,
		QuerySpecs: perfSpecs,
		Cluster:    &cluster,
	}

	res, err := methods.VsanPerfQueryPerf(ctx, c, &perfRequest)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "res: %+v\n", res)
	count := 0

	for _, em := range res.Returnval {
		buckets := make(map[string]metricEntry)
		for _, value := range em.Value {
			name := value.MetricId.Label
			tag := map[string]string{
				"vcenter": v.VCenter,
			}

			// Now deal with the values. Iterate backwards so we start with the latest value
			// tsKey := em.EntityRefId
			valuesSlice := strings.Split(value.Values, ",")
			for idx := len(valuesSlice) - 1; idx >= 0; idx-- {
				ts, _ := time.Parse("2006-01-02 15:04:05", strings.Split(em.SampleInfo, ",")[idx])

				// Since non-realtime metrics are queries with a lookback, we need to check the high-water mark
				// to determine if this should be included. Only samples not seen before should be included.

				value, _ := strconv.ParseFloat(valuesSlice[idx], 64)

				// Organize the metrics into a bucket per measurement.
				// Data SHOULD be presented to us with the same timestamp for all samples, but in case
				// they don't we use the measurement name + timestamp as the key for the bucket.
				mn, fn := "vsan"+name, "vsan"+name //mn=bucket-name=measurement name, fn=field-name
				bKey := mn + " " + " " + strconv.FormatInt(ts.UnixNano(), 10) //bucket key
				bucket, found := buckets[bKey]
				if !found {
					bucket = metricEntry{name: mn, ts: ts, fields: make(map[string]interface{}), tags: tag}
					buckets[bKey] = bucket
				}
				bucket.fields[fn] = value

				// Percentage values must be scaled down by 100.

				count++

				// Update highwater marks for non-realtime metrics.
				//if !res.realTime {
				//	e.hwMarks.Put(tsKey, ts)
				//}
			}
		}
		for _, bucket := range buckets {
			acc.AddFields(bucket.name, bucket.fields, bucket.tags, bucket.ts)
		}
	}
	count++

	return nil

}

func init() {
	inputs.Add("vsan", func() telegraf.Input { return &VSan{} })
}
