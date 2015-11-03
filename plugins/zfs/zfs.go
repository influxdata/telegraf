package zfs

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/telegraf/plugins"
	"github.com/shirou/gopsutil/common"
)

type Zfs struct {
	KstatPath    string
	KstatMetrics []string
}

var sampleConfig = `
  # ZFS kstat path
  # If not specified, then default is:
  # kstatPath = "/proc/spl/kstat/zfs"
  #
  # By default, telegraf gather all zfs stats
  # If not specified, then default is:
  # kstatMetrics = ["arcstats", "zfetchstats", "vdev_cache_stats"]
`

func (z *Zfs) SampleConfig() string {
	return sampleConfig
}

func (z *Zfs) Description() string {
	return "Read metrics of ZFS from arcstats, zfetchstats and vdev_cache_stats"
}

func getTags(kstatPath string) map[string]string {
	var pools string
	poolsDirs, _ := filepath.Glob(kstatPath + "/*/io")
	for _, poolDir := range poolsDirs {
		poolDirSplit := strings.Split(poolDir, "/")
		pool := poolDirSplit[len(poolDirSplit)-2]
		if len(pools) != 0 {
			pools += "::"
		}
		pools += pool
	}
	return map[string]string{"pools": pools}
}

func (z *Zfs) Gather(acc plugins.Accumulator) error {
	kstatMetrics := z.KstatMetrics
	if len(kstatMetrics) == 0 {
		kstatMetrics = []string{"arcstats", "zfetchstats", "vdev_cache_stats"}
	}

	kstatPath := z.KstatPath
	if len(kstatPath) == 0 {
		kstatPath = "/proc/spl/kstat/zfs"
	}

	tags := getTags(kstatPath)

	for _, metric := range kstatMetrics {
		lines, err := common.ReadLines(kstatPath + "/" + metric)
		if err != nil {
			panic(err)
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
			rawValue := rawData[len(rawData)-1]
			value, _ := strconv.ParseInt(rawValue, 10, 64)
			acc.Add(key, value, tags)
		}
	}
	return nil
}

func init() {
	plugins.Add("zfs", func() plugins.Plugin {
		return &Zfs{}
	})
}
