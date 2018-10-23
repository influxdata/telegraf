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

type poolInfo struct {
	name       string
	ioFilename string
}

type objsetInfo struct {
	pool     string
	filename string
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

func getTags(pools []poolInfo) map[string]string {
	var poolNames string

	for _, pool := range pools {
		if len(poolNames) != 0 {
			poolNames += "::"
		}
		poolNames += pool.name
	}

	return map[string]string{"pools": poolNames}
}

func gatherPoolStats(pool poolInfo, acc telegraf.Accumulator) error {
	lines, err := internal.ReadLines(pool.ioFilename)
	if err != nil {
		return err
	}

	if len(lines) != 3 {
		return err
	}

	keys := strings.Fields(lines[1])
	values := strings.Fields(lines[2])

	keyCount := len(keys)

	if keyCount != len(values) {
		return fmt.Errorf("key and value count don't match Keys:%v Values:%v", keys, values)
	}

	tag := map[string]string{"pool": pool.name}
	fields := make(map[string]interface{})
	for i := 0; i < keyCount; i++ {
		value, err := strconv.ParseUint(values[i], 10, 64)
		if err != nil {
			return err
		}
		fields[keys[i]] = value
	}

	// get health if available
	lines, err = internal.ReadLines(filepath.Join(filepath.Dir(pool.ioFilename), "state"))
	if err == nil && len(lines) > 0 {
		tag["health"] = strings.TrimSpace(lines[0])
	}

	acc.AddFields("zfs_pool", fields, tag)

	return nil
}

func getObjsets(kstatPath string) []objsetInfo {
	objsets := make([]objsetInfo, 0)
	objsetPaths, _ := filepath.Glob(kstatPath + "/*/objset-*")

	for _, objsetPath := range objsetPaths {
		objsetDirSplit := strings.Split(objsetPath, "/")
		pool := objsetDirSplit[len(objsetDirSplit)-2]
		objsets = append(objsets, objsetInfo{pool: pool, filename: objsetPath})
	}
	return objsets
}

func gatherObjsetStats(objset objsetInfo, acc telegraf.Accumulator) error {
	lines, err := internal.ReadLines(objset.filename)
	if err != nil {
		return err
	}

	tags := make(map[string]string)
	tags["pool"] = objset.pool
	fields := make(map[string]interface{})
	for i, line := range lines {
		if i == 0 || i == 1 {
			continue
		}
		if len(line) < 1 {
			continue
		}
		k, v, err := parseKstatFields(line)
		if err != nil {
			continue
		}
		if k == "dataset_name" {
			tags["dataset_name"] = v.(string)
		} else {
			fields[k] = v
		}
	}
	acc.AddFields("zfs_objset", fields, tags)

	return nil
}

// constants from https://github.com/zfsonlinux/zfs/blob/master/lib/libspl/include/sys/kstat.h
// kept as strings for comparison thus avoiding conversion to int
//noinspection GoSnakeCaseUsage
const (
	KSTAT_DATA_CHAR   = "0"
	KSTAT_DATA_INT32  = "1"
	KSTAT_DATA_UINT32 = "2"
	KSTAT_DATA_INT64  = "3"
	KSTAT_DATA_UINT64 = "4"
	KSTAT_DATA_LONG   = "5"
	KSTAT_DATA_ULONG  = "6"
	KSTAT_DATA_STRING = "7"
)

func parseKstatFields(line string) (string, interface{}, error) {
	// Linux kstat fields are space separated, but with the addition of string
	// fields in 2018, it is not possible to use a simple, space-separated parser.
	tokens := strings.Fields(line)
	if len(tokens) < 3 {
		return "", "", errors.New("insufficient tokens")
	}
	key := tokens[0]
	switch tokens[1] {
	case KSTAT_DATA_CHAR:
		return key, tokens[2], nil
	case KSTAT_DATA_STRING:
		if len(tokens) == 3 {
			return key, tokens[2], nil
		} else {
			stringTokens := strings.SplitN(line, KSTAT_DATA_STRING, 2)
			return key, strings.TrimLeft(stringTokens[1], " "), nil
		}
	case KSTAT_DATA_INT32, KSTAT_DATA_INT64, KSTAT_DATA_LONG:
		value, _ := strconv.ParseInt(tokens[2], 0, 64)
		return key, value, nil
	case KSTAT_DATA_UINT32, KSTAT_DATA_UINT64, KSTAT_DATA_ULONG:
		value, _ := strconv.ParseUint(tokens[2], 0, 64)
		return key, value, nil
	}
	return key, tokens[2], errors.New("unknown kstat type")
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		// vdev_cache_stats is deprecated
		// xuio_stats are ignored because as of Sep-2016, no known
		// consumers of xuio exist on Linux
		kstatMetrics = []string{"abdstats", "arcstats", "dnodestats", "dbufstats",
			"dmu_tx", "fm", "vdev_mirror_stats", "zfetchstats", "zil"}
	}

	kstatPath := z.KstatPath
	if len(kstatPath) == 0 {
		kstatPath = "/proc/spl/kstat/zfs"
	}

	pools := getPools(kstatPath)
	tags := getTags(pools)

	if z.PoolMetrics {
		for _, pool := range pools {
			err := gatherPoolStats(pool, acc)
			if err != nil {
				return err
			}
		}
	}
	if z.ObjsetMetrics {
		objsets := getObjsets(kstatPath)
		for _, objset := range objsets {
			err := gatherObjsetStats(objset, acc)
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
			k, v, err := parseKstatFields(line)
			if err != nil {
				continue
			}
			key := metric + "_" + k
			// trim redundant_redundant metrics
			if metric == "zil" || metric == "dmu_tx" || metric == "dnodestats" {
				key = k
			}
			fields[key] = v
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
