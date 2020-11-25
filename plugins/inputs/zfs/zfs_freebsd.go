// +build freebsd

package zfs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (z *Zfs) gatherPoolStats(acc telegraf.Accumulator) (string, error) {
	poolFields, err := z.getZpoolStats()
	if err != nil {
		return "", err
	}

	pools := []string{}
	for name := range poolFields {
		pools = append(pools, name)
	}

	if z.PoolMetrics {
		for name, fields := range poolFields {
			tags := map[string]string{
				"pool":   name,
				"health": fields["health"].(string),
			}

			delete(fields, "name")
			delete(fields, "health")

			acc.AddFields("zfs_pool", fields, tags)
		}
	}

	return strings.Join(pools, "::"), nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		kstatMetrics = []string{"arcstats", "zfetchstats", "vdev_cache_stats"}
	}

	poolNames, err := z.gatherPoolStats(acc)
	if err != nil {
		return err
	}
	tags := map[string]string{"pools": poolNames}
	tags["pools"] = poolNames

	fields := make(map[string]interface{})
	for _, metric := range kstatMetrics {
		stdout, err := z.sysctl(metric)
		if err != nil {
			return err
		}
		for _, line := range stdout {
			rawData := strings.Split(line, ": ")
			key := metric + "_" + strings.Split(rawData[0], ".")[4]
			value, _ := strconv.ParseInt(rawData[1], 10, 64)
			fields[key] = value
		}
	}
	acc.AddFields("zfs", fields, tags)
	return nil
}

func sysctl(metric string) ([]string, error) {
	return run("sysctl", []string{"-q", fmt.Sprintf("kstat.zfs.misc.%s", metric)}...)
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{
			sysctl: sysctl,
			zpool:  zpool,
		}
	})
}
