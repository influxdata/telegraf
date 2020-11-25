// +build linux

package zfs

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type poolInfo struct {
	name       string
	ioFilename string
}

func getPools(kstatPath string) []poolInfo {
	pools := make([]poolInfo, 0)
	poolsDirs, _ := filepath.Glob(kstatPath + "/*/io")

	for _, poolDir := range poolsDirs {
		poolDirSplit := strings.Split(poolDir, "/")
		pool := poolDirSplit[len(poolDirSplit)-2]
		pools = append(pools, poolInfo{name: pool, ioFilename: poolDir})
	}

	return pools
}

func getSinglePoolKstat(pool poolInfo) (map[string]interface{}, error) {
	fields := make(map[string]interface{})

	lines, err := internal.ReadLines(pool.ioFilename)
	if err != nil {
		return fields, err
	}

	if len(lines) != 3 {
		return fields, err
	}

	keys := strings.Fields(lines[1])
	values := strings.Fields(lines[2])

	keyCount := len(keys)

	if keyCount != len(values) {
		return fields, fmt.Errorf("Key and value count don't match Keys:%v Values:%v", keys, values)
	}

	for i := 0; i < keyCount; i++ {
		value, err := strconv.ParseInt(values[i], 10, 64)
		if err != nil {
			return fields, err
		}
		fields[keys[i]] = value
	}

	return fields, nil
}

func (z *Zfs) getKstatMetrics() []string {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		// vdev_cache_stats is deprecated
		// xuio_stats are ignored because as of Sep-2016, no known
		// consumers of xuio exist on Linux
		kstatMetrics = []string{"abdstats", "arcstats", "dnodestats", "dbufcachestats",
			"dmu_tx", "fm", "vdev_mirror_stats", "zfetchstats", "zil"}
	}
	return kstatMetrics
}

func (z *Zfs) getKstatPath() string {
	kstatPath := z.KstatPath
	if len(kstatPath) == 0 {
		kstatPath = "/proc/spl/kstat/zfs"
	}
	return kstatPath
}

func (z *Zfs) gatherZfsKstats(acc telegraf.Accumulator, poolNames string) error {
	tags := map[string]string{"pools": poolNames}
	fields := make(map[string]interface{})
	kstatPath := z.getKstatPath()

	for _, metric := range z.getKstatMetrics() {
		lines, err := internal.ReadLines(kstatPath + "/" + metric)
		if err != nil {
			continue
		}
		for i, line := range lines {
			if i == 0 || i == 1 {
				continue
			}
			if len(line) < 1 {
				continue
			}
			rawData := strings.Split(line, " ")
			key := metric + "_" + rawData[0]
			if metric == "zil" || metric == "dmu_tx" || metric == "dnodestats" {
				key = rawData[0]
			}
			rawValue := rawData[len(rawData)-1]
			value, _ := strconv.ParseInt(rawValue, 10, 64)
			fields[key] = value
		}
	}
	acc.AddFields("zfs", fields, tags)
	return nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {

	//Gather pools metrics from kstats
	poolFields, err := z.getZpoolStats()
	if err != nil {
		return err
	}

	poolNames := []string{}
	pools := getPools(z.getKstatPath())
	for _, pool := range pools {
		poolNames = append(poolNames, pool.name)

		if z.PoolMetrics {

			//Merge zpool list with kstats
			fields, err := getSinglePoolKstat(pool)
			if err != nil {
				return err
			} else {
				for k, v := range poolFields[pool.name] {
					fields[k] = v
				}
				tags := map[string]string{
					"pool":   pool.name,
					"health": fields["health"].(string),
				}

				delete(fields, "name")
				delete(fields, "health")

				acc.AddFields("zfs_pool", fields, tags)
			}
		}
	}

	return z.gatherZfsKstats(acc, strings.Join(poolNames, "::"))
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{zpool: zpool}
	})
}
