package nfsclient

import (
	"bufio"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func readMountstatsFile() string {
	path := "./testdata/mountstats"
	if os.Getenv("MOUNT_PROC") != "" {
		path = os.Getenv("MOUNT_PROC")
	}

	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Print(err)
	}
	return string(bytes)
}

func TestNFSClientParsev3(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	nfs3Ops := map[string]bool{"READLINK": true, "GETATTR": false}
	nfs4Ops := map[string]bool{"READLINK": true, "GETATTR": false}
	data := strings.Fields("         READLINK: 500 501 502 503 504 505 506 507")
	nfsclient.parseStat("1.2.3.4:/storage/NFS", "/NFS", "3", data, true, nfs3Ops, nfs4Ops, &acc)

	fields_ops := map[string]interface{}{
		"READLINK_ops":           int64(500),
		"READLINK_trans":         int64(501),
		"READLINK_timeouts":      int64(502),
		"READLINK_bytes_sent":    int64(503),
		"READLINK_bytes_recv":    int64(504),
		"READLINK_queue_time":    int64(505),
		"READLINK_response_time": int64(506),
		"READLINK_total_time":    int64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fields_ops)
}

func TestNFSClientParsev4(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	nfs3Ops := map[string]bool{"DESTROY_SESSION": true, "GETATTR": false}
	nfs4Ops := map[string]bool{"DESTROY_SESSION": true, "GETATTR": false}
	data := strings.Fields("    DESTROY_SESSION: 500 501 502 503 504 505 506 507")
	nfsclient.parseStat("2.2.2.2:/nfsdata/", "/mnt", "4", data, true, nfs3Ops, nfs4Ops, &acc)

	fields_ops := map[string]interface{}{
		"DESTROY_SESSION_ops":           int64(500),
		"DESTROY_SESSION_trans":         int64(501),
		"DESTROY_SESSION_timeouts":      int64(502),
		"DESTROY_SESSION_bytes_sent":    int64(503),
		"DESTROY_SESSION_bytes_recv":    int64(504),
		"DESTROY_SESSION_queue_time":    int64(505),
		"DESTROY_SESSION_response_time": int64(506),
		"DESTROY_SESSION_total_time":    int64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fields_ops)
}

func TestNFSClientProcessStat(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	mountstatstext := readMountstatsFile()

	scanner := bufio.NewScanner(strings.NewReader(mountstatstext))

	nfsclient.processText(scanner, &acc)

	fields_readstat := map[string]interface{}{
		"read_ops":     int64(600),
		"read_retrans": int64(1),
		"read_bytes":   int64(1207),
		"read_rtt":     int64(606),
		"read_exe":     int64(607),
	}
	fields_writestat := map[string]interface{}{
		"write_ops":     int64(700),
		"write_retrans": int64(1),
		"write_bytes":   int64(1407),
		"write_rtt":     int64(706),
		"write_exe":     int64(707),
	}
	tags := map[string]string{
		"serverexport": "1.2.3.4:/storage/NFS",
		"mountpoint":   "/NFS",
	}
	acc.AssertContainsTaggedFields(t, "nfsstat_read", fields_readstat, tags)
	acc.AssertContainsTaggedFields(t, "nfsstat_write", fields_writestat, tags)
}

func TestNFSClientProcessFull(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	nfsclient.Fullstat = true
	mountstatstext := readMountstatsFile()
	scanner := bufio.NewScanner(strings.NewReader(mountstatstext))

	nfsclient.processText(scanner, &acc)

	fields_events := map[string]interface{}{
		"inoderevalidates":  int64(301736),
		"dentryrevalidates": int64(22838),
		"datainvalidates":   int64(410979),
		"attrinvalidates":   int64(26188427),
		"vfsopen":           int64(27525),
		"vfslookup":         int64(9140),
		"vfsaccess":         int64(114420),
		"vfsupdatepage":     int64(30785253),
		"vfsreadpage":       int64(5308856),
		"vfsreadpages":      int64(5364858),
		"vfswritepage":      int64(30784819),
		"vfswritepages":     int64(79832668),
		"vfsgetdents":       int64(170),
		"vfssetattr":        int64(64),
		"vfsflush":          int64(18194),
		"vfsfsync":          int64(29294718),
		"vfslock":           int64(0),
		"vfsrelease":        int64(18279),
		"congestionwait":    int64(0),
		"setattrtrunc":      int64(2),
		"extendwrite":       int64(785551),
		"sillyrenames":      int64(0),
		"shortreads":        int64(0),
		"shortwrites":       int64(0),
		"delay":             int64(0),
		"pnfsreads":         int64(0),
		"pnfswrites":        int64(0),
	}
	fields_bytes := map[string]interface{}{
		"normalreadbytes":  int64(204440464584),
		"normalwritebytes": int64(110857586443),
		"directreadbytes":  int64(783170354688),
		"directwritebytes": int64(296174954496),
		"serverreadbytes":  int64(1134399088816),
		"serverwritebytes": int64(407107155723),
		"readpages":        int64(85749323),
		"writepages":       int64(30784819),
	}
	fields_xprt_tcp := map[string]interface{}{
		"bind_count":    int64(1),
		"connect_count": int64(1),
		"connect_time":  int64(0),
		"idle_time":     int64(0),
		"rpcsends":      int64(96172963),
		"rpcreceives":   int64(96172963),
		"badxids":       int64(0),
		"inflightsends": int64(620878754),
		"backlogutil":   int64(0),
	}

	acc.AssertContainsFields(t, "nfs_events", fields_events)
	acc.AssertContainsFields(t, "nfs_bytes", fields_bytes)
	acc.AssertContainsFields(t, "nfs_xprt_tcp", fields_xprt_tcp)
}
