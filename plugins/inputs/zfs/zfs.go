package zfs

type Sysctl func(metric string) ([]string, error)
type Zpool func() ([]string, error)

type Zfs struct {
	KstatPath    string
	KstatMetrics []string
	PoolMetrics  bool
	sysctl       Sysctl
	zpool        Zpool
}

var sampleConfig = `
  ## ZFS kstat path. Ignored on FreeBSD
  ## If not specified, then default is:
  # kstatPath = "/proc/spl/kstat/zfs"

  ## By default, telegraf gather all zfs stats
  ## If not specified, then default is:
  # kstatMetrics = ["arcstats", "zfetchstats", "vdev_cache_stats"]
  ## For Linux, the default is:
  # kstatMetrics = ["abdstats", "arcstats", "dnodestats", "dbufcachestats",
  #   "dmu_tx", "fm", "vdev_mirror_stats", "zfetchstats", "zil"]
  ## By default, don't gather zpool stats
  # poolMetrics = false
`

func (z *Zfs) SampleConfig() string {
	return sampleConfig
}

func (z *Zfs) Description() string {
	return "Read metrics of ZFS from arcstats, zfetchstats, vdev_cache_stats, and pools"
}
