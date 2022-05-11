package sflow_a10

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func makeMetricsForCounters(p *V5Format, d *PacketDecoder) ([]telegraf.Metric, error) {
	now := time.Now()
	metrics := []telegraf.Metric{}

	for i := 0; i < len(p.Samples); i++ {
		sample := p.Samples[i]

		if sample.SampleCounterData == nil || sample.SampleCounterData.CounterRecords == nil {
			// this is the case when we get a counter with a tag that doesn't exist on our XML file
			// or when we get a "special" counter like 260/271/272 for which we don't collect counter values
			// or we get a sample type that we're not interested in (like "flow")
			d.debug(fmt.Sprintf("  nil sampleCounterData or SampleCounterData.CounterRecords for sampleType %v: %#v", sample.SampleType, sample.SampleCounterData))
			continue
		}

		// this is for 293 and 294
		// as per A10, each packet of either counter block 293 or 294 is just a single sample of 293 or 294
		if !sample.SampleCounterData.NeedsIpAndPort() {
			for j := 0; j < len(sample.SampleCounterData.CounterRecords); j++ {
				counterRecord := sample.SampleCounterData.CounterRecords[j]
				if counterRecord.CounterData == nil {
					d.debug(fmt.Sprintf("  nil CounterData tag is %x for sourceID %x", counterRecord.CounterFormat&4095, sample.SampleCounterData.SourceID))
					continue
				}
				counterFields := counterRecord.CounterData.GetFields()
				counterTags := map[string]string{"agent_address": p.AgentAddress.String()}
				if len(counterFields) > 0 {
					m, err := metric.New("sflow_a10", counterTags, counterFields, now)
					if err != nil {
						d.debug(fmt.Sprintf("  error sending new metric to telegraf %s", err))
						return nil, err
					}

					d.debug(fmt.Sprintf("  sending 293 or 294 metric to telegraf %s", m))
					metrics = append(metrics, m)
				}
			}
			return metrics, nil
		}

		key := createMapKey(sample.SampleCounterData.SourceID, p.AgentAddress.String())

		ipValue, ipExists := d.IPMap.Get(key)
		portValue, portExists := d.PortMap.Get(key)

		if !ipExists || !portExists {
			d.debug(fmt.Sprintf("  sourceID %x and key %v does not exist in IPMap or PortMap", sample.SampleCounterData.SourceID, key))
			continue
		}

		ipDimensions := ipValue.([]IPDimension)
		portDimensions := portValue.(*PortDimension)

		if err := validate(ipDimensions, portDimensions); err != nil {
			//d.debug(fmt.Sprintf("  error in Validate, error is %s, map value is %v whereas counter source ID is %x and key is %v", err, dimensions, sample.SampleCounterData.SourceID, key))
			continue
		}

		for j := 0; j < len(sample.SampleCounterData.CounterRecords); j++ {
			counterRecord := sample.SampleCounterData.CounterRecords[j]

			if counterRecord.CounterData == nil {
				d.debug(fmt.Sprintf("  nil CounterData tag is %x for sourceID %x", counterRecord.CounterFormat&4095, sample.SampleCounterData.SourceID))
				continue
			}

			counterFields := counterRecord.CounterData.GetFields()

			counterTags := counterRecord.CounterData.GetTags(ipDimensions, portDimensions)

			err := appendCommonTags(p, counterTags)
			if err != nil {
				return metrics, err
			}

			if len(counterFields) > 0 {
				m, err := metric.New("sflow_a10", counterTags, counterFields, now)
				if err != nil {
					d.debug(fmt.Sprintf("  error sending new metric to telegraf %s", err))
					return nil, err
				}

				metrics = append(metrics, m)
			}
		}
	}
	return metrics, nil
}

func appendCommonTags(p *V5Format, counterDefinedTags map[string]string) error {
	tags := map[string]string{
		"agent_address": p.AgentAddress.String(),
	}

	for k, v := range tags {
		if _, exists := counterDefinedTags[k]; exists {
			return fmt.Errorf("tag %s exists on counterTags with value %s", k, counterDefinedTags[k])
		}
		counterDefinedTags[k] = v
	}
	return nil
}

// validate returns true if IP and Port Dimensions are valid
func validate(ipDimensions []IPDimension, portDimensions *PortDimension) error {
	if portDimensions == nil {
		return fmt.Errorf("PortDimension is nil")
	} else if ipDimensions == nil {
		return fmt.Errorf("IPDimensions is nil")
	} else if len(ipDimensions) == 0 {
		return fmt.Errorf("IPDimensions has zero length")
	}
	return nil
}
