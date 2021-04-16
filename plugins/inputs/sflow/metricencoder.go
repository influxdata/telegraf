package sflow

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func makeMetrics(p *V5Format) ([]telegraf.Metric, error) {
	now := time.Now()
	metrics := []telegraf.Metric{}
	tags := map[string]string{
		"agent_address": p.AgentAddress.String(),
	}
	fields := map[string]interface{}{}
	for _, sample := range p.Samples {
		tags["input_ifindex"] = strconv.FormatUint(uint64(sample.SampleData.InputIfIndex), 10)
		tags["output_ifindex"] = strconv.FormatUint(uint64(sample.SampleData.OutputIfIndex), 10)
		tags["sample_direction"] = sample.SampleData.SampleDirection
		tags["source_id_index"] = strconv.FormatUint(uint64(sample.SampleData.SourceIDIndex), 10)
		tags["source_id_type"] = strconv.FormatUint(uint64(sample.SampleData.SourceIDType), 10)
		fields["drops"] = sample.SampleData.Drops
		fields["sampling_rate"] = sample.SampleData.SamplingRate

		for _, flowRecord := range sample.SampleData.FlowRecords {
			if flowRecord.FlowData != nil {
				tags2 := flowRecord.FlowData.GetTags()
				fields2 := flowRecord.FlowData.GetFields()
				for k, v := range tags {
					tags2[k] = v
				}
				for k, v := range fields {
					fields2[k] = v
				}
				m := metric.New("sflow", tags2, fields2, now)
				metrics = append(metrics, m)
			}
		}
	}
	return metrics, nil
}
