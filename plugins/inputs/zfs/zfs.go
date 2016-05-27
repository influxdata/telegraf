package zfs

type Zfs struct {
	KstatPath    string
	KstatMetrics []string
	PoolMetrics  bool
}

var sampleConfig = `
  ## ZFS kstat path. Ignored on FreeBSD
  ## If not specified, then default is:
  kstatPath = "/proc/spl/kstat/zfs"

  ## By default, telegraf gather all zfs stats
  ## If not specified, then default is:
  kstatMetrics = ["arcstats", "zfetchstats", "vdev_cache_stats"]

  ## By default, don't gather zpool stats
  poolMetrics = false
`

func (z *Zfs) SampleConfig() string {
	return sampleConfig
}

func (z *Zfs) Description() string {
	return "Read metrics of ZFS from arcstats, zfetchstats and vdev_cache_stats"
}
