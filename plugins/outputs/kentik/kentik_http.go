package kentik

import (
	"log"
	"strconv"

	"github.com/kentik/libkflow/flow"
)

const (
	METRIC_NAME   = "metric_name"
	METRIC_PREFIX = "c_"
)

type KentikMetric struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     uint64            `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func ToFlow(customStrings map[string]uint32, customInts map[string]uint32, met *KentikMetric) *flow.Flow {
	in := flow.Flow{
		TimestampNano: met.Timestamp,
		InBytes:       1,
		InPkts:        met.Value,
		OutBytes:      0,
		OutPkts:       0,
		InputPort:     1,
		OutputPort:    1,
		L4DstPort:     32000,
		Protocol:      16, // use this number for metrics.
		SampleRate:    1,
		SampleAdj:     true,
		Customs:       []flow.Custom{},
	}

	if cid, ok := customStrings[METRIC_PREFIX+METRIC_NAME]; ok {
		in.Customs = append(in.Customs, flow.Custom{
			ID:   cid,
			Type: flow.Str,
			Str:  met.Metric,
		})
	}

	for n, v := range met.Tags {
		if cid, ok := customStrings[METRIC_PREFIX+n]; ok {
			in.Customs = append(in.Customs, flow.Custom{
				ID:   cid,
				Type: flow.Str,
				Str:  v,
			})
		} else if cid, ok := customInts[METRIC_PREFIX+n]; ok {
			intv, _ := strconv.Atoi(v)
			in.Customs = append(in.Customs, flow.Custom{
				ID:   cid,
				Type: flow.U32,
				U32:  uint32(intv),
			})

		}
	}

	return &in
}

func (met *KentikMetric) Print() {
	log.Printf("Kentik: %s %d %d %v", met.Metric, met.Value, met.Timestamp, met.Tags)
}
