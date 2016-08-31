package nfsclient

import (
	"bufio"
	"github.com/influxdata/telegraf/testutil"
	"strings"
	"testing"
)

const mountstatstext = `device rootfs mounted on / with fstype rootfs
device proc mounted on /proc with fstype proc
device sysfs mounted on /sys with fstype sysfs
device devtmpfs mounted on /dev with fstype devtmpfs
device devpts mounted on /dev/pts with fstype devpts
device tmpfs mounted on /dev/shm with fstype tmpfs
device /dev/loop0 mounted on /dev/.initramfs/live with fstype iso9660
device /dev/loop6 mounted on / with fstype ext4
device /proc/bus/usb mounted on /proc/bus/usb with fstype usbfs
device none mounted on /proc/sys/fs/binfmt_misc with fstype binfmt_misc
device /tmp mounted on /tmp with fstype tmpfs
device /home mounted on /home with fstype tmpfs
device /var mounted on /var with fstype tmpfs
device /etc mounted on /etc with fstype tmpfs
device /dev/ram1 mounted on /root with fstype ext2
device cgroup mounted on /cgroup/cpuset with fstype cgroup
device cgroup mounted on /cgroup/cpu with fstype cgroup
device cgroup mounted on /cgroup/cpuacct with fstype cgroup
device cgroup mounted on /cgroup/memory with fstype cgroup
device cgroup mounted on /cgroup/devices with fstype cgroup
device cgroup mounted on /cgroup/freezer with fstype cgroup
device cgroup mounted on /cgroup/net_cls with fstype cgroup
device cgroup mounted on /cgroup/blkio with fstype cgroup
device sunrpc mounted on /var/lib/nfs/rpc_pipefs with fstype rpc_pipefs
device /etc/auto.misc mounted on /misc with fstype autofs
device -hosts mounted on /net with fstype autofs
device 1.2.3.4:/storage/NFS mounted on /NFS with fstype nfs statvers=1.1
    opts:   rw,vers=3,rsize=32768,wsize=32768,namlen=255,acregmin=60,acregmax=60,acdirmin=60,acdirmax=60,hard,nolock,noacl,nordirplus,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=1.2.3.4,mountvers=3,mountport=49193,mountproto=tcp,local_lock=all
    age:    1136770
    caps:   caps=0x3fe6,wtmult=512,dtsize=8192,bsize=0,namlen=255
    sec:    flavor=1,pseudoflavor=1
    events: 301736 22838 410979 26188427 27525 9140 114420 30785253 5308856 5364858 30784819 79832668 170 64 18194 29294718 0 18279 0 2 785551 0 0 0 0 0 0 
    bytes:  204440464584 110857586443 783170354688 296174954496 1134399088816 407107155723 85749323 30784819 
    RPC iostats version: 1.0  p/v: 100003/3 (nfs)
    xprt:   tcp 733 1 1 0 0 96172963 96172963 0 620878754 0 690 196347132 524706275
    per-op statistics
            NULL: 0 0 0 0 0 0 0 0
         GETATTR: 100 101 102 103 104 105 106 107
         SETATTR: 200 201 202 203 204 205 206 207
          LOOKUP: 300 301 302 303 304 305 306 307
          ACCESS: 400 401 402 403 404 405 406 407
        READLINK: 500 501 502 503 504 505 506 507
            READ: 600 601 602 603 604 605 606 607
           WRITE: 700 701 702 703 704 705 706 707
          CREATE: 800 801 802 803 804 805 806 807
           MKDIR: 900 901 902 903 904 905 906 907
         SYMLINK: 1000 1001 1002 1003 1004 1005 1006 1007 
           MKNOD: 1100 1101 1102 1103 1104 1105 1106 1107 
          REMOVE: 1200 1201 1202 1203 1204 1205 1206 1207 
           RMDIR: 1300 1301 1302 1303 1304 1305 1306 1307 
          RENAME: 1400 1401 1402 1403 1404 1405 1406 1407 
            LINK: 1500 1501 1502 1503 1504 1505 1506 1507 
         READDIR: 1600 1601 1602 1603 1604 1605 1606 1607 
     READDIRPLUS: 1700 1701 1702 1703 1704 1705 1706 1707 
          FSSTAT: 1800 1801 1802 1803 1804 1805 1806 1807 
          FSINFO: 1900 1901 1902 1903 1904 1905 1906 1907 
        PATHCONF: 2000 2001 2002 2003 2004 2005 2006 2007 
          COMMIT: 2100 2101 2102 2103 2104 2105 2106 2107 

device 2.2.2.2:/nfsdata/ mounted on /mnt with fstype nfs4 statvers=1.1
    opts:    rw,vers=4,rsize=1048576,wsize=1048576,namlen=255,acregmin=3,acregmax=60,
            acdirmin=30,acdirmax=60,hard,proto=tcp,port=0,timeo=600,retrans=2,sec=sys,
            clientaddr=3.3.3.3,minorversion=0,local_lock=none
    age:    19
    caps:    caps=0xfff7,wtmult=512,dtsize=32768,bsize=0,namlen=255
    nfsv4:    bm0=0xfdffafff,bm1=0xf9be3e,acl=0x0
    sec:    flavor=1,pseudoflavor=1
    events:    0 168232 0 0 0 10095 217808 0 2 9797 0 9739 0 0 19739 19739 0 19739 0 0 0 0 0 0 0 0 0
    bytes:    1612840960 0 0 0 627536112 0 158076 0
    RPC iostats version: 1.0  p/v: 100003/4 (nfs)
    xprt:    tcp 737 0 1 0 0 69698 69697 0 81817 0 2 1082 12119
    per-op statistics
            NULL: 0 0 0 0 0 0 0 0
            READ: 9797 9797 0 1000 2000 71 7953 8200
           WRITE: 0 0 0 0 0 0 0 0
          COMMIT: 0 0 0 0 0 0 0 0
            OPEN: 19740 19740 0 4737600 7343280 505 3449 4172
    OPEN_CONFIRM: 10211 10211 0 1552072 694348 74 836 1008
     OPEN_NOATTR: 0 0 0 0 0 0 0 0
    OPEN_DOWNGRADE: 0 0 0 0 0 0 0 0
           CLOSE: 19739 19739 0 3316152 2605548 334 3045 3620
         SETATTR: 0 0 0 0 0 0 0 0
          FSINFO: 1 1 0 132 108 0 0 0
           RENEW: 0 0 0 0 0 0 0 0
     SETCLIENTID: 0 0 0 0 0 0 0 0
    SETCLIENTID_CONFIRM: 0 0 0 0 0 0 0 0
            LOCK: 0 0 0 0 0 0 0 0
           LOCKT: 0 0 0 0 0 0 0 0
           LOCKU: 0 0 0 0 0 0 0 0
          ACCESS: 96 96 0 14584 19584 0 8 10
         GETATTR: 1 1 0 132 188 0 0 0
          LOOKUP: 10095 10095 0 1655576 2382420 36 898 1072
     LOOKUP_ROOT: 0 0 0 0 0 0 0 0
          REMOVE: 0 0 0 0 0 0 0 0
          RENAME: 0 0 0 0 0 0 0 0
            LINK: 0 0 0 0 0 0 0 0
         SYMLINK: 0 0 0 0 0 0 0 0
          CREATE: 0 0 0 0 0 0 0 0
        PATHCONF: 1 1 0 128 72 0 0 0
          STATFS: 0 0 0 0 0 0 0 0
        READLINK: 0 0 0 0 0 0 0 0
         READDIR: 0 0 0 0 0 0 0 0
     SERVER_CAPS: 2 2 0 256 176 0 0 0
     DELEGRETURN: 0 0 0 0 0 0 0 0
          GETACL: 0 0 0 0 0 0 0 0
          SETACL: 0 0 0 0 0 0 0 0
    FS_LOCATIONS: 0 0 0 0 0 0 0 0
    RELEASE_LOCKOWNER: 0 0 0 0 0 0 0 0
         SECINFO: 0 0 0 0 0 0 0 0
     EXCHANGE_ID: 0 0 0 0 0 0 0 0
    CREATE_SESSION: 0 0 0 0 0 0 0 0
    DESTROY_SESSION: 500 501 502 503 504 505 506 507
        SEQUENCE: 0 0 0 0 0 0 0 0
    GET_LEASE_TIME: 0 0 0 0 0 0 0 0
    RECLAIM_COMPLETE: 0 0 0 0 0 0 0 0
       LAYOUTGET: 0 0 0 0 0 0 0 0
    GETDEVICEINFO: 0 0 0 0 0 0 0 0
    LAYOUTCOMMIT: 0 0 0 0 0 0 0 0
    LAYOUTRETURN: 0 0 0 0 0 0 0 0

`

func TestNFSCLIENTParsev3(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSCLIENT{}
	data := strings.Fields("         READLINK: 500 501 502 503 504 505 506 507")
	nfsclient.parseData("1.2.3.4:/storage/NFS", "/NFS", "3", data, &acc)

	fields_ops := map[string]interface{}{
		"READLINK_ops":           float64(500),
		"READLINK_trans":         float64(501),
		"READLINK_timeouts":      float64(502),
		"READLINK_bytes_sent":    float64(503),
		"READLINK_bytes_recv":    float64(504),
		"READLINK_queue_time":    float64(505),
		"READLINK_response_time": float64(506),
		"READLINK_total_time":    float64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fields_ops)
}

func TestNFSCLIENTParsev4(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSCLIENT{}
	data := strings.Fields("    DESTROY_SESSION: 500 501 502 503 504 505 506 507")
	nfsclient.parseData("2.2.2.2:/nfsdata/", "/mnt", "4", data, &acc)

	fields_ops := map[string]interface{}{
		"DESTROY_SESSION_ops":           float64(500),
		"DESTROY_SESSION_trans":         float64(501),
		"DESTROY_SESSION_timeouts":      float64(502),
		"DESTROY_SESSION_bytes_sent":    float64(503),
		"DESTROY_SESSION_bytes_recv":    float64(504),
		"DESTROY_SESSION_queue_time":    float64(505),
		"DESTROY_SESSION_response_time": float64(506),
		"DESTROY_SESSION_total_time":    float64(507),
	}
	acc.AssertContainsFields(t, "nfs_ops", fields_ops)
}

func TestNFSCLIENTProcessStat(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSCLIENT{}
	scanner := bufio.NewScanner(strings.NewReader(mountstatstext))

	nfsclient.processText(scanner, &acc)

	fields_readstat := map[string]interface{}{
		"read_ops":     float64(600),
		"read_retrans": float64(1),
		"read_bytes":   float64(1207),
		"read_rtt":     float64(606),
		"read_exe":     float64(607),
	}
	fields_writestat := map[string]interface{}{
		"write_ops":     float64(700),
		"write_retrans": float64(1),
		"write_bytes":   float64(1407),
		"write_rtt":     float64(706),
		"write_exe":     float64(707),
	}
	tags := map[string]string{
		"serverexport": "1.2.3.4:/storage/NFS",
		"mountpoint": "/NFS",
	}
	acc.AssertContainsTaggedFields(t, "nfsstat_read", fields_readstat, tags)
	acc.AssertContainsTaggedFields(t, "nfsstat_write", fields_writestat, tags)
}

func TestNFSCLIENTProcessFull(t *testing.T) {
	var acc testutil.Accumulator

	nfsclient := NFSCLIENT{}
	nfsclient.Fullstat = true
	scanner := bufio.NewScanner(strings.NewReader(mountstatstext))

	nfsclient.processText(scanner, &acc)

	fields_events := map[string]interface{}{
		"inoderevalidates":  float64(301736),
		"dentryrevalidates": float64(22838),
		"datainvalidates":   float64(410979),
		"attrinvalidates":   float64(26188427),
		"vfsopen":           float64(27525),
		"vfslookup":         float64(9140),
		"vfspermission":     float64(114420),
		"vfsupdatepage":     float64(30785253),
		"vfsreadpage":       float64(5308856),
		"vfsreadpages":      float64(5364858),
		"vfswritepage":      float64(30784819),
		"vfswritepages":     float64(79832668),
		"vfsreaddir":        float64(170),
		"vfssetattr":        float64(64),
		"vfsflush":          float64(18194),
		"vfsfsync":          float64(29294718),
		"vfslock":           float64(0),
		"vfsrelease":        float64(18279),
		"congestionwait":    float64(0),
		"setattrtrunc":      float64(2),
		"extendwrite":       float64(785551),
		"sillyrenames":      float64(0),
		"shortreads":        float64(0),
		"shortwrites":       float64(0),
		"delay":             float64(0),
		"pnfsreads":         float64(0),
		"pnfswrites":        float64(0),
	}
	fields_bytes := map[string]interface{}{
		"normalreadbytes":  float64(204440464584),
		"normalwritebytes": float64(110857586443),
		"directreadbytes":  float64(783170354688),
		"directwritebytes": float64(296174954496),
		"serverreadbytes":  float64(1134399088816),
		"serverwritebytes": float64(407107155723),
		"readpages":        float64(85749323),
		"writepages":       float64(30784819),
	}
	fields_xprttcp := map[string]interface{}{
		//        "port": float64(733),
		"bind_count":    float64(1),
		"connect_count": float64(1),
		"connect_time":  float64(0),
		"idle_time":     float64(0),
		"rpcsends":      float64(96172963),
		"rpcreceives":   float64(96172963),
		"badxids":       float64(0),
		"inflightsends": float64(620878754),
		"backlogutil":   float64(0),
	}

	acc.AssertContainsFields(t, "nfs_events", fields_events)
	acc.AssertContainsFields(t, "nfs_bytes", fields_bytes)
	acc.AssertContainsFields(t, "nfs_xprttcp", fields_xprttcp)
}
