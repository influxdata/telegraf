//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package conntrack

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

//go:embed sample.conf
var sampleConfig string

type Conntrack struct {
	ps      system.PS
	Path    string
	Dirs    []string
	Files   []string
	Collect []string
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

func (c *Conntrack) setDefaults() {
	if len(c.Dirs) == 0 {
		c.Dirs = dfltDirs
	}

	if len(c.Files) == 0 {
		c.Files = dfltFiles
	}
}

func (*Conntrack) SampleConfig() string {
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

			contents, err := os.ReadFile(fName)
			if err != nil {
				acc.AddError(fmt.Errorf("failed to read file '%s': %v", fName, err))
				continue
			}

			v := strings.TrimSpace(string(contents))
			fields[metricKey], err = strconv.ParseFloat(v, 64)
			if err != nil {
				acc.AddError(fmt.Errorf("failed to parse metric, expected number but "+
					" found '%s': %v", v, err))
			}
		}
	}

	var all bool
	var perCPU bool

	for _, collect := range c.Collect {
		if collect == "all" {
			all = true
		}
		if collect == "percpu" {
			perCPU = true
		}
	}

	if all || perCPU {
		stats, err := c.ps.NetConntrack(perCPU)
		if err != nil {
			acc.AddError(fmt.Errorf("failed to retrieve conntrack statistics: %v", err))
		}

		if len(stats) == 0 {
			acc.AddError(fmt.Errorf("conntrack input failed to collect stats"))
		}

		for i, sts := range stats {
			cpuTag := "all"
			if perCPU {
				cpuTag = fmt.Sprintf("cpu%d", i)
			}
			tags := map[string]string{
				"cpu": cpuTag,
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
