package sflow

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func makeMetrics(p *v5Format) []telegraf.Metric {
	now := time.Now()
	metrics := make([]telegraf.Metric, 0)
	tags := map[string]string{
		"agent_address": p.AgentAddress.String(),
	}
	fields := make(map[string]interface{}, 2)
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
				tags2 := flowRecord.FlowData.getTags()
				fields2 := flowRecord.FlowData.getFields()
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
	return metrics
}
