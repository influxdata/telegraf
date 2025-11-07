//go:build linux

package zfs

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	unknown metricsVersion = iota
	v1
	v2

	// Procstat type-column values according to
	// https://github.com/openzfs/zfs/blob/master/include/os/linux/spl/sys/kstat.h#L54
	kstatDataChar   = 0
	kstatDataInt32  = 1
	kstatDataUint32 = 2
	kstatDataInt64  = 3
	kstatDataUint64 = 4
	kstatDataLong   = 5
	kstatDataULong  = 6
	kstatDataString = 7
)

type metricsVersion uint8

type poolInfo struct {
	name       string
	ioFilename string
	version    metricsVersion
}

type helper struct{} //nolint:unused // not used for "linux" OS, needed for Zfs struct

func (z *Zfs) Init() error {
	// Set defaults
	if z.KstatPath == "" {
		z.KstatPath = "/proc/spl/kstat/zfs"
	}

	if len(z.KstatMetrics) == 0 {
		// vdev_cache_stats is deprecated
		// xuio_stats are ignored because as of Sep-2016, no known
		// consumers of xuio exist on Linux
		z.KstatMetrics = []string{
			"abdstats",
			"arcstats",
			"dnodestats",
			"dbufcachestats",
			"dmu_tx",
			"fm",
			"vdev_mirror_stats",
			"zfetchstats",
			"zil",
		}
	}

	// Check settings
	// We need to check the kstat metrics _after_ assigning the default to
	// allow explicitly disabling the kstat metrics via an empty string. For
	// processing we need to remove the empty string to not confuse the code.
	z.KstatMetrics = slices.DeleteFunc(z.KstatMetrics, func(m string) bool { return m == "" })
	for _, m := range z.KstatMetrics {
		switch m {
		case "abdstats", "arcstats", "dnodestats", "dbufcachestats", "dmu_tx",
			"fm", "vdev_mirror_stats", "zfetchstats", "zil":
			// Do nothing, those are valid
		default:
			return fmt.Errorf("invalid kstat metric %q", m)
		}
	}

	return nil
}

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	pools, err := getPools(z.KstatPath)
	tags := getTags(pools)

	if z.PoolMetrics && err == nil {
		for _, pool := range pools {
			if err := z.gatherPoolStats(pool, acc); err != nil {
				return err
			}
		}
	}

	fields := make(map[string]interface{})
	for _, metric := range z.KstatMetrics {
		fn := filepath.Join(z.KstatPath, metric)
		lines, err := internal.ReadLines(fn)
		if err != nil {
			continue
		}

		data, err := z.processProcFile(lines)
		if err != nil {
			return fmt.Errorf("gathering metric %q from %q failed: %w", metric, fn, err)
		}
		for k, v := range data {
			switch metric {
			case "zil", "dmu_tx", "dnodestats":
				// Keep key as is
			default:
				// Prefix key with metric name
				k = metric + "_" + k
			}
			fields[k] = v
		}
	}
	acc.AddFields("zfs", fields, tags)
	return nil
}

func getPools(path string) ([]poolInfo, error) {
	pools := make([]poolInfo, 0)
	version, poolsDirs, err := probeVersion(path)
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

func probeVersion(path string) (metricsVersion, []string, error) {
	poolsDirs, err := filepath.Glob(path + "/*/objset-*")

	// From the docs: the only possible returned error is ErrBadPattern, when pattern is malformed.
	// Because of this we need to determine how to fallback differently.
	if err != nil {
		return unknown, poolsDirs, err
	}

	if len(poolsDirs) > 0 {
		return v2, poolsDirs, nil
	}

	// Fallback to the old kstat in case of an older ZFS version.
	poolsDirs, err = filepath.Glob(path + "/*/io")
	if err != nil {
		return unknown, poolsDirs, err
	}

	return v1, poolsDirs, nil
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

func (z *Zfs) gatherPoolStats(pool poolInfo, acc telegraf.Accumulator) error {
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
		fields, gatherErr = z.gatherV2(lines, tags)
	case unknown:
		return errors.New("unknown metrics version detected")
	}

	if gatherErr != nil {
		return fmt.Errorf("collecting pool stats from %q failed: %w", pool.ioFilename, gatherErr)
	}

	acc.AddFields("zfs_pool", fields, tags)
	return nil
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

func gather(lines []string, fileLines int) (keys, values []string, err error) {
	if len(lines) < fileLines {
		return nil, nil, errors.New("expected lines in kstat does not match")
	}

	keys = strings.Fields(lines[1])
	values = strings.Fields(lines[2])
	if len(keys) != len(values) {
		return nil, nil, fmt.Errorf("key and value count don't match Keys:%v Values:%v", keys, values)
	}

	return keys, values, nil
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
func (z *Zfs) gatherV2(lines []string, tags map[string]string) (map[string]interface{}, error) {
	fields, err := z.processProcFile(lines)
	if err != nil {
		return nil, err
	}
	if len(fields) < 7 {
		return nil, fmt.Errorf("expected 7 lines but got %d", len(fields))
	}

	// Extract the dataset name as a tag and remove it from the fields
	dsnRaw, found := fields["dataset_name"]
	if !found {
		return nil, errors.New("dataset name not found in data")
	}
	dsn, ok := dsnRaw.(string)
	if !ok {
		return nil, fmt.Errorf("invalid type %T for dataset name %v", dsnRaw, dsnRaw)
	}
	tags["dataset"] = dsn
	delete(fields, "dataset_name")

	return fields, nil
}

func (z *Zfs) processProcFile(lines []string) (map[string]interface{}, error) {
	// Ignore the first lines as it contains data in a different format
	// The second line (index 1) does contain the column header and should read
	// name				type	data
	header := strings.Fields(lines[1])
	if len(header) != 3 || header[0] != "name" || header[1] != "type" || header[2] != "data" {
		return nil, fmt.Errorf("invalid header %q", lines[1])
	}

	// Extract the data
	data := make(map[string]interface{}, len(lines)-2)
	for i, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return data, fmt.Errorf("invalid data in line %d: %s", i+3, line)
		}
		name := fields[0]
		ftype, err := strconv.Atoi(fields[1])
		if err != nil {
			z.Log.Warnf("cannot parse type %q for field %q; falling back to integer", fields[1], name)
			ftype = kstatDataInt64
		}

		switch ftype {
		case kstatDataChar, kstatDataString:
			data[name] = fields[2]
		case kstatDataInt32, kstatDataInt64:
			value, err := strconv.ParseInt(fields[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing field %q with %q failed: %w", name, fields[2], err)
			}
			data[name] = value
		case kstatDataUint32, kstatDataUint64:
			value, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing field %q with %q failed: %w", name, fields[2], err)
			}
			// For backward compatibility in the metric field-types
			if z.UseNativeTypes {
				data[name] = value
			} else {
				data[name] = int64(value)
			}
		case kstatDataLong, kstatDataULong:
			value, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("parsing field %q with %q failed: %w", name, fields[2], err)
			}
			data[name] = value
		default:
			z.Log.Errorf("field %q with %q has unknown type %d", name, fields[2], ftype)
		}
	}

	return data, nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
