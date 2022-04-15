//go:build linux
// +build linux

package zfs

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type metricsVersion uint8

const (
	unknown metricsVersion = iota
	v1
	v2
)

type poolInfo struct {
	name       string
	ioFilename string
	version    metricsVersion
}

func probeVersion(kstatPath string) (metricsVersion, []string, error) {
	poolsDirs, err := filepath.Glob(fmt.Sprintf("%s/*/objset-*", kstatPath))

	// From the docs: the only possible returned error is ErrBadPattern, when pattern is malformed.
	// Because of this we need to determine how to fallback differently.
	if err != nil {
		return unknown, poolsDirs, err
	}

	if len(poolsDirs) > 0 {
		return v2, poolsDirs, nil
	}

	// Fallback to the old kstat in case of an older ZFS version.
	poolsDirs, err = filepath.Glob(fmt.Sprintf("%s/*/io", kstatPath))
	if err != nil {
		return unknown, poolsDirs, err
	}

	return v1, poolsDirs, nil
}

func getPools(kstatPath string) ([]poolInfo, error) {
	pools := make([]poolInfo, 0)
	version, poolsDirs, err := probeVersion(kstatPath)
	if err != nil {
		return nil, err
	}

	for _, poolDir := range poolsDirs {
		poolDirSplit := strings.Split(poolDir, "/")
		pool := poolDirSplit[len(poolDirSplit)-2]
		pools = append(pools, poolInfo{name: pool, ioFilename: poolDir, version: version})
	}

	return pools, nil
}

func getTags(pools []poolInfo) map[string]string {
	poolNames := ""
	knownPools := make(map[string]struct{})
	for _, entry := range pools {
		name := entry.name
		if _, ok := knownPools[name]; !ok {
			knownPools[name] = struct{}{}
			if poolNames != "" {
				poolNames += "::"
			}
			poolNames += name
		}
	}

	return map[string]string{"pools": poolNames}
}

func gather(lines []string, fileLines int) ([]string, []string, error) {
	if len(lines) != fileLines {
		return nil, nil, errors.New("expected lines in kstat does not match")
	}

	keys := strings.Fields(lines[1])
	values := strings.Fields(lines[2])
	if len(keys) != len(values) {
		return nil, nil, fmt.Errorf("key and value count don't match Keys:%v Values:%v", keys, values)
	}

	return keys, values, nil
}

func gatherV1(lines []string) (map[string]interface{}, error) {
	fileLines := 3
	keys, values, err := gather(lines, fileLines)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]interface{})
	for i := 0; i < len(keys); i++ {
		value, err := strconv.ParseInt(values[i], 10, 64)
		if err != nil {
			return nil, err
		}

		fields[keys[i]] = value
	}

	return fields, nil
}

// New way of collection. Each objset-* file in ZFS >= 2.1.x has a format looking like this:
// 36 1 0x01 7 2160 5214787391 73405258558961
// name                            type data
// dataset_name                    7    rpool/ROOT/pve-1
// writes                          4    409570
// nwritten                        4    2063419969
// reads                           4    22108699
// nread                           4    63067280992
// nunlinks                        4    13849
// nunlinked                       4    13848
//
// For explanation of the first line's values see https://github.com/openzfs/zfs/blob/master/module/os/linux/spl/spl-kstat.c#L61
func gatherV2(lines []string, tags map[string]string) (map[string]interface{}, error) {
	fileLines := 9
	_, _, err := gather(lines, fileLines)
	if err != nil {
		return nil, err
	}

	tags["dataset"] = strings.Fields(lines[2])[2]
	fields := make(map[string]interface{})
	for i := 3; i < len(lines); i++ {
		lineFields := strings.Fields(lines[i])
		fieldName := lineFields[0]
		fieldData := lineFields[2]
		value, err := strconv.ParseInt(fieldData, 10, 64)
		if err != nil {
			return nil, err
		}

		fields[fieldName] = value
	}

	return fields, nil
}

func gatherPoolStats(pool poolInfo, acc telegraf.Accumulator) error {
	lines, err := internal.ReadLines(pool.ioFilename)
	if err != nil {
		return err
	}

	var fields map[string]interface{}
	var gatherErr error
	tags := map[string]string{"pool": pool.name}
	switch pool.version {
	case v1:
		fields, gatherErr = gatherV1(lines)
	case v2:
		fields, gatherErr = gatherV2(lines, tags)
	case unknown:
		return errors.New("Unknown metrics version detected")
	}

	if gatherErr != nil {
		return err
	}

	acc.AddFields("zfs_pool", fields, tags)
	return nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		// vdev_cache_stats is deprecated
		// xuio_stats are ignored because as of Sep-2016, no known
		// consumers of xuio exist on Linux
		kstatMetrics = []string{"abdstats", "arcstats", "dnodestats", "dbufcachestats",
			"dmu_tx", "fm", "vdev_mirror_stats", "zfetchstats", "zil"}
	}

	kstatPath := z.KstatPath
	if len(kstatPath) == 0 {
		kstatPath = "/proc/spl/kstat/zfs"
	}

	pools, err := getPools(kstatPath)
	tags := getTags(pools)

	if z.PoolMetrics && err == nil {
		for _, pool := range pools {
			err := gatherPoolStats(pool, acc)
			if err != nil {
				return err
			}
		}
	}

	fields := make(map[string]interface{})
	for _, metric := range kstatMetrics {
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

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
