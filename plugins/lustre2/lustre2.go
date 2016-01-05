/*
Lustre 2.x telegraf plugin

Lustre (http://lustre.org/) is an open-source, parallel file system
for HPC environments. It stores statistics about its activity in
/proc

*/
package lustre2

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/plugins"
)

// Lustre proc files can change between versions, so we want to future-proof
// by letting people choose what to look at.
type Lustre2 struct {
	Ost_procfiles []string
	Mds_procfiles []string
}

var sampleConfig = `
  # An array of /proc globs to search for Lustre stats
  # If not specified, the default will work on Lustre 2.5.x
  #
  # ost_procfiles = ["/proc/fs/lustre/obdfilter/*/stats", "/proc/fs/lustre/osd-ldiskfs/*/stats"]
  # mds_procfiles = ["/proc/fs/lustre/mdt/*/md_stats"]
`

/* The wanted fields would be a []string if not for the
lines that start with read_bytes/write_bytes and contain
   both the byte count and the function call count
*/
type mapping struct {
	inProc   string // What to look for at the start of a line in /proc/fs/lustre/*
	field    uint32 // which field to extract from that line
	reportAs string // What measurement name to use
	tag      string // Additional tag to add for this metric
}

var wanted_ost_fields = []*mapping{
	{
		inProc:   "write_bytes",
		field:    6,
		reportAs: "write_bytes",
	},
	{ // line starts with 'write_bytes', but value write_calls is in second column
		inProc:   "write_bytes",
		field:    1,
		reportAs: "write_calls",
	},
	{
		inProc:   "read_bytes",
		field:    6,
		reportAs: "read_bytes",
	},
	{ // line starts with 'read_bytes', but value read_calls is in second column
		inProc:   "read_bytes",
		field:    1,
		reportAs: "read_calls",
	},
	{
		inProc: "cache_hit",
	},
	{
		inProc: "cache_miss",
	},
	{
		inProc: "cache_access",
	},
}

var wanted_mds_fields = []*mapping{
	{
		inProc: "open",
	},
	{
		inProc: "close",
	},
	{
		inProc: "mknod",
	},
	{
		inProc: "link",
	},
	{
		inProc: "unlink",
	},
	{
		inProc: "mkdir",
	},
	{
		inProc: "rmdir",
	},
	{
		inProc: "rename",
	},
	{
		inProc: "getattr",
	},
	{
		inProc: "setattr",
	},
	{
		inProc: "getxattr",
	},
	{
		inProc: "setxattr",
	},
	{
		inProc: "statfs",
	},
	{
		inProc: "sync",
	},
	{
		inProc: "samedir_rename",
	},
	{
		inProc: "crossdir_rename",
	},
}

func (l *Lustre2) GetLustreProcStats(fileglob string, wanted_fields []*mapping, acc plugins.Accumulator) error {
	files, err := filepath.Glob(fileglob)
	if err != nil {
		return err
	}

	for _, file := range files {
		/* Turn /proc/fs/lustre/obdfilter/<ost_name>/stats and similar
		 * into just the object store target name
		 * Assumpion: the target name is always second to last,
		 * which is true in Lustre 2.1->2.5
		 */
		path := strings.Split(file, "/")
		name := path[len(path)-2]
		tags := map[string]string{
			"name": name,
		}

		lines, err := internal.ReadLines(file)
		if err != nil {
			return err
		}

		fields := make(map[string]interface{})
		for _, line := range lines {
			parts := strings.Fields(line)
			for _, wanted := range wanted_fields {
				var data uint64
				if parts[0] == wanted.inProc {
					wanted_field := wanted.field
					// if not set, assume field[1]. Shouldn't be field[0], as
					// that's a string
					if wanted_field == 0 {
						wanted_field = 1
					}
					data, err = strconv.ParseUint((parts[wanted_field]), 10, 64)
					if err != nil {
						return err
					}
					report_name := wanted.inProc
					if wanted.reportAs != "" {
						report_name = wanted.reportAs
					}
					fields[report_name] = data
				}
			}
		}
		acc.AddFields("lustre2", fields)
	}
	return nil
}

// SampleConfig returns sample configuration message
func (l *Lustre2) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Lustre2 plugin
func (l *Lustre2) Description() string {
	return "Read metrics from local Lustre service on OST, MDS"
}

// Gather reads stats from all lustre targets
func (l *Lustre2) Gather(acc plugins.Accumulator) error {

	if len(l.Ost_procfiles) == 0 {
		// read/write bytes are in obdfilter/<ost_name>/stats
		err := l.GetLustreProcStats("/proc/fs/lustre/obdfilter/*/stats", wanted_ost_fields, acc)
		if err != nil {
			return err
		}
		// cache counters are in osd-ldiskfs/<ost_name>/stats
		err = l.GetLustreProcStats("/proc/fs/lustre/osd-ldiskfs/*/stats", wanted_ost_fields, acc)
		if err != nil {
			return err
		}
	}

	if len(l.Mds_procfiles) == 0 {
		// Metadata server stats
		err := l.GetLustreProcStats("/proc/fs/lustre/mdt/*/md_stats", wanted_mds_fields, acc)
		if err != nil {
			return err
		}
	}

	for _, procfile := range l.Ost_procfiles {
		err := l.GetLustreProcStats(procfile, wanted_ost_fields, acc)
		if err != nil {
			return err
		}
	}
	for _, procfile := range l.Mds_procfiles {
		err := l.GetLustreProcStats(procfile, wanted_mds_fields, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	plugins.Add("lustre2", func() plugins.Plugin {
		return &Lustre2{}
	})
}
