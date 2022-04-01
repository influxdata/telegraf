//go:build !windows
// +build !windows

// Package lustre2 (doesn't aim for Windows)
// Lustre 2.x Telegraf plugin
// Lustre (http://lustre.org/) is an open-source, parallel file system
// for HPC environments. It stores statistics about its activity in /proc
package lustre2

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type tags struct {
	name, job, client string
}

// Lustre proc files can change between versions, so we want to future-proof
// by letting people choose what to look at.
type Lustre2 struct {
	OstProcfiles []string `toml:"ost_procfiles"`
	MdsProcfiles []string `toml:"mds_procfiles"`

	// allFields maps and OST name to the metric fields associated with that OST
	allFields map[tags]map[string]interface{}
}

/* The wanted fields would be a []string if not for the
lines that start with read_bytes/write_bytes and contain
   both the byte count and the function call count
*/
type mapping struct {
	inProc   string // What to look for at the start of a line in /proc/fs/lustre/*
	field    uint32 // which field to extract from that line
	reportAs string // What measurement name to use
}

var wantedOstFields = []*mapping{
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

var wantedOstJobstatsFields = []*mapping{
	{ // The read line has several fields, so we need to differentiate what they are
		inProc:   "read",
		field:    3,
		reportAs: "jobstats_read_calls",
	},
	{
		inProc:   "read",
		field:    7,
		reportAs: "jobstats_read_min_size",
	},
	{
		inProc:   "read",
		field:    9,
		reportAs: "jobstats_read_max_size",
	},
	{
		inProc:   "read",
		field:    11,
		reportAs: "jobstats_read_bytes",
	},
	{ // Different inProc for newer versions
		inProc:   "read_bytes",
		field:    3,
		reportAs: "jobstats_read_calls",
	},
	{
		inProc:   "read_bytes",
		field:    7,
		reportAs: "jobstats_read_min_size",
	},
	{
		inProc:   "read_bytes",
		field:    9,
		reportAs: "jobstats_read_max_size",
	},
	{
		inProc:   "read_bytes",
		field:    11,
		reportAs: "jobstats_read_bytes",
	},
	{ // We need to do the same for the write fields
		inProc:   "write",
		field:    3,
		reportAs: "jobstats_write_calls",
	},
	{
		inProc:   "write",
		field:    7,
		reportAs: "jobstats_write_min_size",
	},
	{
		inProc:   "write",
		field:    9,
		reportAs: "jobstats_write_max_size",
	},
	{
		inProc:   "write",
		field:    11,
		reportAs: "jobstats_write_bytes",
	},
	{ // Different inProc for newer versions
		inProc:   "write_bytes",
		field:    3,
		reportAs: "jobstats_write_calls",
	},
	{
		inProc:   "write_bytes",
		field:    7,
		reportAs: "jobstats_write_min_size",
	},
	{
		inProc:   "write_bytes",
		field:    9,
		reportAs: "jobstats_write_max_size",
	},
	{
		inProc:   "write_bytes",
		field:    11,
		reportAs: "jobstats_write_bytes",
	},
	{
		inProc:   "getattr",
		field:    3,
		reportAs: "jobstats_ost_getattr",
	},
	{
		inProc:   "setattr",
		field:    3,
		reportAs: "jobstats_ost_setattr",
	},
	{
		inProc:   "punch",
		field:    3,
		reportAs: "jobstats_punch",
	},
	{
		inProc:   "sync",
		field:    3,
		reportAs: "jobstats_ost_sync",
	},
	{
		inProc:   "destroy",
		field:    3,
		reportAs: "jobstats_destroy",
	},
	{
		inProc:   "create",
		field:    3,
		reportAs: "jobstats_create",
	},
	{
		inProc:   "statfs",
		field:    3,
		reportAs: "jobstats_ost_statfs",
	},
	{
		inProc:   "get_info",
		field:    3,
		reportAs: "jobstats_get_info",
	},
	{
		inProc:   "set_info",
		field:    3,
		reportAs: "jobstats_set_info",
	},
	{
		inProc:   "quotactl",
		field:    3,
		reportAs: "jobstats_quotactl",
	},
}

var wantedMdsFields = []*mapping{
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

var wantedMdtJobstatsFields = []*mapping{
	{
		inProc:   "open",
		field:    3,
		reportAs: "jobstats_open",
	},
	{
		inProc:   "close",
		field:    3,
		reportAs: "jobstats_close",
	},
	{
		inProc:   "mknod",
		field:    3,
		reportAs: "jobstats_mknod",
	},
	{
		inProc:   "link",
		field:    3,
		reportAs: "jobstats_link",
	},
	{
		inProc:   "unlink",
		field:    3,
		reportAs: "jobstats_unlink",
	},
	{
		inProc:   "mkdir",
		field:    3,
		reportAs: "jobstats_mkdir",
	},
	{
		inProc:   "rmdir",
		field:    3,
		reportAs: "jobstats_rmdir",
	},
	{
		inProc:   "rename",
		field:    3,
		reportAs: "jobstats_rename",
	},
	{
		inProc:   "getattr",
		field:    3,
		reportAs: "jobstats_getattr",
	},
	{
		inProc:   "setattr",
		field:    3,
		reportAs: "jobstats_setattr",
	},
	{
		inProc:   "getxattr",
		field:    3,
		reportAs: "jobstats_getxattr",
	},
	{
		inProc:   "setxattr",
		field:    3,
		reportAs: "jobstats_setxattr",
	},
	{
		inProc:   "statfs",
		field:    3,
		reportAs: "jobstats_statfs",
	},
	{
		inProc:   "sync",
		field:    3,
		reportAs: "jobstats_sync",
	},
	{
		inProc:   "samedir_rename",
		field:    3,
		reportAs: "jobstats_samedir_rename",
	},
	{
		inProc:   "crossdir_rename",
		field:    3,
		reportAs: "jobstats_crossdir_rename",
	},
}

func (l *Lustre2) GetLustreProcStats(fileglob string, wantedFields []*mapping) error {
	files, err := filepath.Glob(fileglob)
	if err != nil {
		return err
	}

	fieldSplitter := regexp.MustCompile(`[ :]+`)

	for _, file := range files {

		/* From /proc/fs/lustre/obdfilter/<ost_name>/stats and similar
		 * extract the object store target name,
		 * and for per-client files under
		 * /proc/fs/lustre/obdfilter/<ost_name>/exports/<client_nid>/stats
		 * and similar the client NID
		 * Assumption: the target name is fourth to last
		 * for per-client files and second to last otherwise
		 * and the client NID is always second to last,
		 * which is true in Lustre 2.1->2.14
		 */
		path := strings.Split(file, "/")
		var name, client string
		if strings.Contains(file, "/exports/") {
			name = path[len(path)-4]
			client = path[len(path)-2]
		} else {
			name = path[len(path)-2]
			client = ""
		}

		//lines, err := internal.ReadLines(file)
		wholeFile, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		jobs := strings.Split(string(wholeFile), "- ")
		for _, job := range jobs {
			lines := strings.Split(job, "\n")
			jobid := ""

			// figure out if the data should be tagged with job_id here
			parts := strings.Fields(lines[0])
			if strings.TrimSuffix(parts[0], ":") == "job_id" {
				jobid = parts[1]
			}

			for _, line := range lines {
				// skip any empty lines
				if len(line) < 1 {
					continue
				}

				parts := fieldSplitter.Split(line, -1)
				if len(parts[0]) == 0 {
					parts = parts[1:]
				}

				var fields map[string]interface{}
				fields, ok := l.allFields[tags{name, jobid, client}]
				if !ok {
					fields = make(map[string]interface{})
					l.allFields[tags{name, jobid, client}] = fields
				}

				for _, wanted := range wantedFields {
					var data uint64
					if parts[0] == wanted.inProc {
						wantedField := wanted.field
						// if not set, assume field[1]. Shouldn't be field[0], as
						// that's a string
						if wantedField == 0 {
							wantedField = 1
						}
						data, err = strconv.ParseUint(strings.TrimSuffix(parts[wantedField], ","), 10, 64)
						if err != nil {
							return err
						}
						reportName := wanted.inProc
						if wanted.reportAs != "" {
							reportName = wanted.reportAs
						}
						fields[reportName] = data
					}
				}
			}
		}
	}
	return nil
}

// Gather reads stats from all lustre targets
func (l *Lustre2) Gather(acc telegraf.Accumulator) error {
	//l.allFields = make(map[string]map[string]interface{})
	l.allFields = make(map[tags]map[string]interface{})

	if len(l.OstProcfiles) == 0 {
		// read/write bytes are in obdfilter/<ost_name>/stats
		err := l.GetLustreProcStats("/proc/fs/lustre/obdfilter/*/stats", wantedOstFields)
		if err != nil {
			return err
		}
		// cache counters are in osd-ldiskfs/<ost_name>/stats
		err = l.GetLustreProcStats("/proc/fs/lustre/osd-ldiskfs/*/stats", wantedOstFields)
		if err != nil {
			return err
		}
		// per job statistics are in obdfilter/<ost_name>/job_stats
		err = l.GetLustreProcStats("/proc/fs/lustre/obdfilter/*/job_stats", wantedOstJobstatsFields)
		if err != nil {
			return err
		}
	}

	if len(l.MdsProcfiles) == 0 {
		// Metadata server stats
		err := l.GetLustreProcStats("/proc/fs/lustre/mdt/*/md_stats", wantedMdsFields)
		if err != nil {
			return err
		}

		// Metadata target job stats
		err = l.GetLustreProcStats("/proc/fs/lustre/mdt/*/job_stats", wantedMdtJobstatsFields)
		if err != nil {
			return err
		}
	}

	for _, procfile := range l.OstProcfiles {
		ostFields := wantedOstFields
		if strings.HasSuffix(procfile, "job_stats") {
			ostFields = wantedOstJobstatsFields
		}
		err := l.GetLustreProcStats(procfile, ostFields)
		if err != nil {
			return err
		}
	}
	for _, procfile := range l.MdsProcfiles {
		mdtFields := wantedMdsFields
		if strings.HasSuffix(procfile, "job_stats") {
			mdtFields = wantedMdtJobstatsFields
		}
		err := l.GetLustreProcStats(procfile, mdtFields)
		if err != nil {
			return err
		}
	}

	for tgs, fields := range l.allFields {
		tags := map[string]string{
			"name": tgs.name,
		}
		if len(tgs.job) > 0 {
			tags["jobid"] = tgs.job
		}
		if len(tgs.client) > 0 {
			tags["client"] = tgs.client
		}
		acc.AddFields("lustre2", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("lustre2", func() telegraf.Input {
		return &Lustre2{}
	})
}
