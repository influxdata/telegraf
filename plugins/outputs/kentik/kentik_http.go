package kentik

import (
	"github.com/kentik/libkflow/flow"
)

type KentikMetric struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     uint64            `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func ToFlow(customs map[string]uint32, met *KentikMetric) *flow.Flow {
	in := flow.Flow{
		TimestampNano: met.Timestamp,
		InBytes:       met.Value,
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

	for n, v := range met.Tags {
		if cid, ok := customs[n]; ok {
			in.Customs = append(in.Customs, flow.Custom{
				ID:   cid,
				Type: flow.Str,
				Str:  v,
			})
		}
	}

	return &in
}
