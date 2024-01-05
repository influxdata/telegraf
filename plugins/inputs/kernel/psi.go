//go:build linux

package kernel

import (
	"fmt"
	"time"

	"github.com/prometheus/procfs"

	"github.com/influxdata/telegraf"
)

// Gather PSI metrics
func (k *Kernel) gatherPressure(acc telegraf.Accumulator) error {
	for _, resource := range []string{"cpu", "memory", "io"} {
		now := time.Now()
		psiStats, err := k.procfs.PSIStatsForResource(resource)
		if err != nil {
			return fmt.Errorf("failed to read %s pressure: %w", resource, err)
		}

		stats := map[string]*procfs.PSILine{
			"some": psiStats.Some,
			"full": psiStats.Full,
		}

		for _, typ := range []string{"some", "full"} {
			if resource == "cpu" && typ == "full" {
				// resource=cpu,type=full is omitted because it is always zero
				continue
			}

			tags := map[string]string{
				"resource": resource,
				"type":     typ,
			}
			stat := stats[typ]

			acc.AddCounter("pressure", map[string]interface{}{
				"total": stat.Total,
			}, tags, now)
			acc.AddGauge("pressure", map[string]interface{}{
				"avg10":  stat.Avg10,
				"avg60":  stat.Avg60,
				"avg300": stat.Avg300,
			}, tags, now)
		}
	}
	return nil
}
