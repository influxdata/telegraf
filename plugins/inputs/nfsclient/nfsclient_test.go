package nfsclient

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func getMountStatsPath() string {
	path := "./testdata/mountstats"
	if os.Getenv("MOUNT_PROC") != "" {
		path = os.Getenv("MOUNT_PROC")
	}

	return path
}

func TestNFSClientParsev3(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{Fullstat: true}
	nfsclient.nfs3Ops = map[string]bool{"READLINK": true, "GETATTR": false}
	nfsclient.nfs4Ops = map[string]bool{"READLINK": true, "GETATTR": false}
	data := strings.Fields("         READLINK: 500 501 502 503 504 505 506 507")
	err := nfsclient.parseStat("1.2.3.4:/storage/NFS", "/A", "3", data, &acc)
	require.NoError(t, err)

	fieldsOps := map[string]interface{}{
		"ops":           uint64(500),
		"trans":         uint64(501),
		"timeouts":      uint64(502),
		"bytes_sent":    uint64(503),
		"bytes_recv":    uint64(504),
		"queue_time":    uint64(505),
		"response_time": uint64(506),
		"total_time":    uint64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fieldsOps)
}

func TestNFSClientParsev4(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{Fullstat: true}
	nfsclient.nfs3Ops = map[string]bool{"DESTROY_SESSION": true, "GETATTR": false}
	nfsclient.nfs4Ops = map[string]bool{"DESTROY_SESSION": true, "GETATTR": false}
	data := strings.Fields("    DESTROY_SESSION: 500 501 502 503 504 505 506 507")
	err := nfsclient.parseStat("2.2.2.2:/nfsdata/", "/B", "4", data, &acc)
	require.NoError(t, err)

	fieldsOps := map[string]interface{}{
		"ops":           uint64(500),
		"trans":         uint64(501),
		"timeouts":      uint64(502),
		"bytes_sent":    uint64(503),
		"bytes_recv":    uint64(504),
		"queue_time":    uint64(505),
		"response_time": uint64(506),
		"total_time":    uint64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fieldsOps)
}

func TestNFSClientParseLargeValue(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{Fullstat: true}
	nfsclient.nfs3Ops = map[string]bool{"SETCLIENTID": true, "GETATTR": false}
	nfsclient.nfs4Ops = map[string]bool{"SETCLIENTID": true, "GETATTR": false}
	data := strings.Fields("    SETCLIENTID: 218 216 0 53568 12960 18446744073709531008 134 197")
	err := nfsclient.parseStat("2.2.2.2:/nfsdata/", "/B", "4", data, &acc)
	require.NoError(t, err)

	fieldsOps := map[string]interface{}{
		"ops":           uint64(218),
		"trans":         uint64(216),
		"timeouts":      uint64(0),
		"bytes_sent":    uint64(53568),
		"bytes_recv":    uint64(12960),
		"queue_time":    uint64(18446744073709531008),
		"response_time": uint64(134),
		"total_time":    uint64(197),
	}
	acc.AssertContainsFields(t, "nfs_ops", fieldsOps)
}

func TestNFSClientProcessStat(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	nfsclient.Fullstat = false

	file, _ := os.Open(getMountStatsPath())
	defer file.Close()

	scanner := bufio.NewScanner(file)

	err := nfsclient.processText(scanner, &acc)
	require.NoError(t, err)

	fieldsReadstat := map[string]interface{}{
		"ops":        uint64(600),
		"retrans":    uint64(1),
		"bytes":      uint64(1207),
		"rtt":        uint64(606),
		"exe":        uint64(607),
		"rtt_per_op": float64(1.01),
	}

	readTags := map[string]string{
		"serverexport": "1.2.3.4:/storage/NFS",
		"mountpoint":   "/A",
		"operation":    "READ",
	}

	acc.AssertContainsTaggedFields(t, "nfsstat", fieldsReadstat, readTags)

	fieldsWritestat := map[string]interface{}{
		"ops":        uint64(700),
		"retrans":    uint64(1),
		"bytes":      uint64(1407),
		"rtt":        uint64(706),
		"exe":        uint64(707),
		"rtt_per_op": float64(1.0085714285714287),
	}

	writeTags := map[string]string{
		"serverexport": "1.2.3.4:/storage/NFS",
		"mountpoint":   "/A",
		"operation":    "WRITE",
	}
	acc.AssertContainsTaggedFields(t, "nfsstat", fieldsWritestat, writeTags)
}

func TestNFSClientProcessFull(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSClient{}
	nfsclient.Fullstat = true

	file, _ := os.Open(getMountStatsPath())
	defer file.Close()

	scanner := bufio.NewScanner(file)

	err := nfsclient.processText(scanner, &acc)
	require.NoError(t, err)

	fieldsEvents := map[string]interface{}{
		"inoderevalidates":  uint64(301736),
		"dentryrevalidates": uint64(22838),
		"datainvalidates":   uint64(410979),
		"attrinvalidates":   uint64(26188427),
		"vfsopen":           uint64(27525),
		"vfslookup":         uint64(9140),
		"vfsaccess":         uint64(114420),
		"vfsupdatepage":     uint64(30785253),
		"vfsreadpage":       uint64(5308856),
		"vfsreadpages":      uint64(5364858),
		"vfswritepage":      uint64(30784819),
		"vfswritepages":     uint64(79832668),
		"vfsgetdents":       uint64(170),
		"vfssetattr":        uint64(64),
		"vfsflush":          uint64(18194),
		"vfsfsync":          uint64(29294718),
		"vfslock":           uint64(0),
		"vfsrelease":        uint64(18279),
		"congestionwait":    uint64(0),
		"setattrtrunc":      uint64(2),
		"extendwrite":       uint64(785551),
		"sillyrenames":      uint64(0),
		"shortreads":        uint64(0),
		"shortwrites":       uint64(0),
		"delay":             uint64(0),
		"pnfsreads":         uint64(0),
		"pnfswrites":        uint64(0),
	}
	fieldsBytes := map[string]interface{}{
		"normalreadbytes":  uint64(204440464584),
		"normalwritebytes": uint64(110857586443),
		"directreadbytes":  uint64(783170354688),
		"directwritebytes": uint64(296174954496),
		"serverreadbytes":  uint64(1134399088816),
		"serverwritebytes": uint64(407107155723),
		"readpages":        uint64(85749323),
		"writepages":       uint64(30784819),
	}
	fieldsXprtTCP := map[string]interface{}{
		"bind_count":    uint64(1),
		"connect_count": uint64(1),
		"connect_time":  uint64(0),
		"idle_time":     uint64(0),
		"rpcsends":      uint64(96172963),
		"rpcreceives":   uint64(96172963),
		"badxids":       uint64(0),
		"inflightsends": uint64(620878754),
		"backlogutil":   uint64(0),
	}

	acc.AssertContainsFields(t, "nfs_events", fieldsEvents)
	acc.AssertContainsFields(t, "nfs_bytes", fieldsBytes)
	acc.AssertContainsFields(t, "nfs_xprt_tcp", fieldsXprtTCP)
}
