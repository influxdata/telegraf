// +build linux

package conntrack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type Conntrack struct {
	ps           system.PS
	Path         string
	Dirs         []string
	Files        []string
	CollectStats bool `toml:"collect_stats"`
	PerCPU       bool `toml:"percpu"`
}

const (
	inputName = "conntrack"
)

var dfltDirs = []string{
	"/proc/sys/net/ipv4/netfilter",
	"/proc/sys/net/netfilter",
}

var dfltFiles = []string{
	"ip_conntrack_count",
	"ip_conntrack_max",
	"nf_conntrack_count",
	"nf_conntrack_max",
}

func NewConntrack(ps system.PS) *Conntrack {
	return &Conntrack{
		ps:           ps,
		CollectStats: true,
		PerCPU:       false,
	}
}

func (c *Conntrack) setDefaults() {
	if len(c.Dirs) == 0 {
		c.Dirs = dfltDirs
	}

	if len(c.Files) == 0 {
		c.Files = dfltFiles
	}
}

func (c *Conntrack) Description() string {
	return "Collects conntrack stats from the configured directories and files."
}

var sampleConfig = `
   ## The following defaults would work with multiple versions of conntrack.
   ## Note the nf_ and ip_ filename prefixes are mutually exclusive across
   ## kernel versions, as are the directory locations.

   ## Superset of filenames to look for within the conntrack dirs.
   ## Missing files will be ignored.
   files = ["ip_conntrack_count","ip_conntrack_max",
            "nf_conntrack_count","nf_conntrack_max"]

   ## Directories to search within for the conntrack files above.
   ## Missing directories will be ignored.
   dirs = ["/proc/sys/net/ipv4/netfilter","/proc/sys/net/netfilter"]
   ## If true, collect conntrack stats
   collect_stats = false
   ## Whether to report per-cpu stats or not
   percpu = false
`

func (c *Conntrack) SampleConfig() string {
	return sampleConfig
}

func (c *Conntrack) Gather(acc telegraf.Accumulator) error {
	c.setDefaults()

	var metricKey string
	fields := make(map[string]interface{})

	for _, dir := range c.Dirs {
		for _, file := range c.Files {
			// NOTE: no system will have both nf_ and ip_ prefixes,
			// so we're safe to branch on suffix only.
			parts := strings.SplitN(file, "_", 2)
			if len(parts) < 2 {
				continue
			}
			metricKey = "ip_" + parts[1]

			fName := filepath.Join(dir, file)
			if _, err := os.Stat(fName); err != nil {
				continue
			}

			contents, err := ioutil.ReadFile(fName)
			if err != nil {
				acc.AddError(fmt.Errorf("E! failed to read file '%s': %v", fName, err))
				continue
			}

			v := strings.TrimSpace(string(contents))
			fields[metricKey], err = strconv.ParseFloat(v, 64)
			if err != nil {
				acc.AddError(fmt.Errorf("E! failed to parse metric, expected number but "+
					" found '%s': %v", v, err))
			}
		}
	}

	if c.CollectStats {
		stats, err := c.ps.NetConntrack(c.PerCPU)
		if err != nil {
			acc.AddError(fmt.Errorf("E! failed to retrieve conntrack statistics: %v", err))
		}

		for i, sts := range stats {
			var tags map[string]string
			if c.PerCPU {
				tags = map[string]string{
					"cpu": fmt.Sprintf("cpu%d", i),
				}
			} else {
				tags = map[string]string{
					"cpu": "all",
				}
			}

			statFields := map[string]interface{}{
				"entries":        sts.Entries,       // entries in the conntrack table
				"searched":       sts.Searched,      // conntrack table lookups performed
				"found":          sts.Found,         // searched entries which were successful
				"new":            sts.New,           // entries added which were not expected before
				"invalid":        sts.Invalid,       // packets seen which can not be tracked
				"ignore":         sts.Ignore,        // packets seen which are already connected to an entry
				"delete":         sts.Delete,        // entries which were removed
				"delete_list":    sts.DeleteList,    // entries which were put to dying list
				"insert":         sts.Insert,        // entries inserted into the list
				"insert_failed":  sts.InsertFailed,  // insertion attempted but failed (same entry exists)
				"drop":           sts.Drop,          // packets dropped due to conntrack failure
				"early_drop":     sts.EarlyDrop,     // dropped entries to make room for new ones, if maxsize reached
				"icmp_error":     sts.IcmpError,     // Subset of invalid. Packets that can't be tracked d/t error
				"expect_new":     sts.ExpectNew,     // Entries added after an expectation was already present
				"expect_create":  sts.ExpectCreate,  // Expectations added
				"expect_delete":  sts.ExpectDelete,  // Expectations deleted
				"search_restart": sts.SearchRestart, // onntrack table lookups restarted due to hashtable resizes
			}
			acc.AddCounter(inputName, statFields, tags)
		}
	}

	if len(fields) == 0 {
		return fmt.Errorf("Conntrack input failed to collect metrics. " +
			"Is the conntrack kernel module loaded?")
	}

	acc.AddFields(inputName, fields, nil)
	return nil
}

func init() {
	inputs.Add(inputName, func() telegraf.Input { return &Conntrack{} })
}
