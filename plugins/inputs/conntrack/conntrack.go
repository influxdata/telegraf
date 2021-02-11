// +build linux

package conntrack

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Conntrack struct {
	Path       string
	Dirs       []string
	Files      []string
	Conntrack  bool
	ConnTable  []string
	ConnStates []string
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

var dfltConnTable = []string{"/usr/sbin/conntrack", "-L", "-o", "extended"}

func (c *Conntrack) setDefaults() {
	if len(c.Dirs) == 0 {
		c.Dirs = dfltDirs
	}

	if len(c.Files) == 0 {
		c.Files = dfltFiles
	}

	if len(c.ConnTable) == 0 {
		c.ConnTable = dfltConnTable
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
   ## Missing directrories will be ignored.
   dirs = ["/proc/sys/net/ipv4/netfilter","/proc/sys/net/netfilter"]

   ## Connections tracking table from Linux kernel
   ## This metrics are only for servers with the nf_conntrack kerne module
   ## and the conntrack command installed.

   ## Use conntrack to enable the nf_conntrack metrics
   conntrack = false
   ## Change the location path as needed
   conntable = ["/usr/sbin/conntrack", "-L", "-o", "extended"]
`

func (c *Conntrack) SampleConfig() string {
	return sampleConfig
}

func (c *Conntrack) Gather(acc telegraf.Accumulator) error {
	c.setDefaults()
	if err := c.gatherCounters(acc); err != nil {
		return err
	}
	return c.gatherConnStates(acc)
}

func (c *Conntrack) gatherCounters(acc telegraf.Accumulator) error {
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

	if len(fields) == 0 {
		return fmt.Errorf("Conntrack input failed to collect metrics. " +
			"Is the conntrack kernel module loaded?")
	}

	acc.AddFields(inputName, fields, nil)
	return nil
}

func (c *Conntrack) gatherConnStates(acc telegraf.Accumulator) error {
	if !c.Conntrack {
		return nil
	}

	fields := make(map[string]interface{})

	nf := newNfConntrack()

	var cmd *exec.Cmd
	if len(c.ConnTable) == 1 {
		cmd = exec.Command(c.ConnTable[0])
	} else {
		cmd = exec.Command(c.ConnTable[0], c.ConnTable[1:]...)
	}
	cmd.Stdout = nf
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	for k, v := range nf.counters {
		fields[k] = v
	}

	acc.AddFields(inputName, fields, nil)

	return nil
}

func init() {
	inputs.Add(inputName, func() telegraf.Input { return &Conntrack{} })
}
