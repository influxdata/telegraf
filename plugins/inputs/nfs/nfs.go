package nfs

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NFS struct {
	Iostat   bool
	Fullstat bool
}

var sampleConfig = `
  ## Read fewer metrics (iostat)
  iostat = true

  ## Read all metrics
  fullstat = true
`

func (n *NFS) SampleConfig() string {
	return sampleConfig
}

func (n *NFS) Description() string {
	return "Read NFS metrics from /proc/self/mountstats"
}

var eventsFields = []string{
	"inoderevalidates",
	"dentryrevalidates",
	"datainvalidates",
	"attrinvalidates",
	"vfsopen",
	"vfslookup",
	"vfspermission",
	"vfsupdatepage",
	"vfsreadpage",
	"vfsreadpages",
	"vfswritepage",
	"vfswritepages",
	"vfsreaddir",
	"vfssetattr",
	"vfsflush",
	"vfsfsync",
	"vfslock",
	"vfsrelease",
	"congestionwait",
	"setattrtrunc",
	"extendwrite",
	"sillyrenames",
	"shortreads",
	"shortwrites",
	"delay",
	"pnfsreads",
	"pnfswrites",
}

var bytesFields = []string{
	"normalreadbytes",
	"normalwritebytes",
	"directreadbytes",
	"directwritebytes",
	"serverreadbytes",
	"serverwritebytes",
	"readpages",
	"writepages",
}

var xprtudpFields = []string{
	//        "port",
	"bind_count",
	"rpcsends",
	"rpcreceives",
	"badxids",
	"inflightsends",
	"backlogutil",
}

var xprttcpFields = []string{
	//        "port",
	"bind_count",
	"connect_count",
	"connect_time",
	"idle_time",
	"rpcsends",
	"rpcreceives",
	"badxids",
	"inflightsends",
	"backlogutil",
}

var nfs3Fields = []string{
	"NULL",
	"GETATTR",
	"SETATTR",
	"LOOKUP",
	"ACCESS",
	"READLINK",
	"READ",
	"WRITE",
	"CREATE",
	"MKDIR",
	"SYMLINK",
	"MKNOD",
	"REMOVE",
	"RMDIR",
	"RENAME",
	"LINK",
	"READDIR",
	"READDIRPLUS",
	"FSSTAT",
	"FSINFO",
	"PATHCONF",
	"COMMIT",
}

var nfs4Fields = []string{
	"NULL",
	"READ",
	"WRITE",
	"COMMIT",
	"OPEN",
	"OPEN_CONFIRM",
	"OPEN_NOATTR",
	"OPEN_DOWNGRADE",
	"CLOSE",
	"SETATTR",
	"FSINFO",
	"RENEW",
	"SETCLIENTID",
	"SETCLIENTID_CONFIRM",
	"LOCK",
	"LOCKT",
	"LOCKU",
	"ACCESS",
	"GETATTR",
	"LOOKUP",
	"LOOKUP_ROOT",
	"REMOVE",
	"RENAME",
	"LINK",
	"SYMLINK",
	"CREATE",
	"PATHCONF",
	"STATFS",
	"READLINK",
	"READDIR",
	"SERVER_CAPS",
	"DELEGRETURN",
	"GETACL",
	"SETACL",
	"FS_LOCATIONS",
	"RELEASE_LOCKOWNER",
	"SECINFO",
	"FSID_PRESENT",
	"EXCHANGE_ID",
	"CREATE_SESSION",
	"DESTROY_SESSION",
	"SEQUENCE",
	"GET_LEASE_TIME",
	"RECLAIM_COMPLETE",
	"LAYOUTGET",
	"GETDEVICEINFO",
	"LAYOUTCOMMIT",
	"LAYOUTRETURN",
	"SECINFO_NO_NAME",
	"TEST_STATEID",
	"FREE_STATEID",
	"GETDEVICELIST",
	"BIND_CONN_TO_SESSION",
	"DESTROY_CLIENTID",
	"SEEK",
	"ALLOCATE",
	"DEALLOCATE",
	"LAYOUTSTATS",
	"CLONE",
}

var nfsopFields = []string{
	"ops",
	"trans",
	"timeouts",
	"bytes_sent",
	"bytes_recv",
	"queue_time",
	"response_time",
	"total_time",
}

func convert(line []string) []float64 {
	var nline []float64
	for _, l := range line[1:] {
		f, _ := strconv.ParseFloat(l, 64)
		nline = append(nline, f)
	}
	return nline
}

func In(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func (n *NFS) parseStat(mountpoint string, version string, line []string, acc telegraf.Accumulator) error {
	tags := map[string]string{"mountpoint": mountpoint}
	nline := convert(line)
	first := strings.Replace(line[0], ":", "", 1)

	var fields = make(map[string]interface{})

	if version == "3" || version == "4" {
		if In(nfs3Fields, first) {
			if first == "READ" {
				fields["read_ops"] = nline[0]
				fields["read_retrans"] = (nline[1] - nline[0])
				fields["read_bytes"] = (nline[3] + nline[4])
				fields["read_rtt"] = nline[6]
				fields["read_exe"] = nline[7]
				acc.AddFields("nfsstat_read", fields, tags)
			} else if first == "WRITE" {
				fields["write_ops"] = nline[0]
				fields["write_retrans"] = (nline[1] - nline[0])
				fields["write_bytes"] = (nline[3] + nline[4])
				fields["write_rtt"] = nline[6]
				fields["write_exe"] = nline[7]
				acc.AddFields("nfsstat_write", fields, tags)
			}
		}
	}
	return nil
}

func (n *NFS) parseData(mountpoint string, version string, line []string, acc telegraf.Accumulator) error {
	tags := map[string]string{"mountpoint": mountpoint}
	nline := convert(line)
	first := strings.Replace(line[0], ":", "", 1)

	var fields = make(map[string]interface{})

	if first == "events" {
		for i, t := range eventsFields {
			fields[t] = nline[i]
		}
		acc.AddFields("nfs_events", fields, tags)
	} else if first == "bytes" {
		for i, t := range bytesFields {
			fields[t] = nline[i]
		}
		acc.AddFields("nfs_bytes", fields, tags)
	} else if first == "xprt" {
		switch line[1] {
		case "tcp":
			{
				for i, t := range xprttcpFields {
					fields[t] = nline[i+2]
				}
				acc.AddFields("nfs_xprttcp", fields, tags)
			}
		case "udp":
			{
				for i, t := range xprtudpFields {
					fields[t] = nline[i+2]
				}
				acc.AddFields("nfs_xprtudp", fields, tags)
			}
		}
	} else if version == "3" {
		if In(nfs3Fields, first) {
			for i, t := range nline {
				item := fmt.Sprintf("%s_%s", first, nfsopFields[i])
				fields[item] = t
			}
			acc.AddFields("nfs_ops", fields, tags)
		}
	} else if version == "4" {
		if In(nfs4Fields, first) {
			for i, t := range nline {
				item := fmt.Sprintf("%s_%s", first, nfsopFields[i])
				fields[item] = t
			}
			acc.AddFields("nfs_ops", fields, tags)
		}
	}
	return nil
}

func (n *NFS) processText(scanner *bufio.Scanner, acc telegraf.Accumulator) error {
	var device string
	var version string
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if In(line, "fstype") && In(line, "nfs") {
			device = fmt.Sprintf("%s %s", line[1], line[4])
		} else if In(line, "(nfs)") {
			version = strings.Split(line[5], "/")[1]
		}
		if len(line) > 0 {
			if n.Iostat == true {
				n.parseStat(device, version, line, acc)
			}
			if n.Fullstat == true {
				n.parseData(device, version, line, acc)
			}
		}
	}
	return nil
}

func (n *NFS) Gather(acc telegraf.Accumulator) error {
	var outerr error

	file, err := os.Open("/proc/self/mountstats")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	n.processText(scanner, acc)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return outerr
}

func init() {
	inputs.Add("nfs", func() telegraf.Input {
		return &NFS{}
	})
}
