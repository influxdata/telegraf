package nfsstat

import (
	"strconv"
	"strings"

	"errors"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
)

var (
	ErrorInvalidNFSVersion   error = errors.New("invalid nfs version")
	ErrorInvalidNFSOpIndex   error = errors.New("invalid nfs operation index")
	ErrorInvalidCounterValue error = errors.New("invalid counter value")
)

const (
	prefix_nfs3 string = "proc3 22 "
	prefix_nfs4 string = "proc4 60 "
)

type NFSStat struct {
	Log telegraf.Logger
}

type NFSStatCounter struct {
	nfsvers string
	op      string
	val     int64
}

func (n *NFSStat) SampleConfig() string {
	return "[[inputs.nfsstat]]"
}

func (n *NFSStat) Description() string {
	return "Read global nfsclient stats"
}

func getNFSVers(s string) (int, bool) {
	if strings.HasPrefix(s, prefix_nfs3) {
		return 3, true
	}

	if strings.HasPrefix(s, prefix_nfs4) {
		return 4, true
	}

	return 0, false
}

func getNFSv3Ops() []string {
	return []string{
		"null",
		"getattr",
		"setattr",
		"lookup",
		"access",
		"readlink",
		"read",
		"write",
		"create",
		"mkdir",
		"symlink",
		"mknod",
		"remove",
		"rmdir",
		"rename",
		"link",
		"readdir",
		"readdirplus",
		"fsstat",
		"fsinfo",
		"pathconf",
		"commit",
	}
}

func getNFSv4Ops() []string {
	return []string{
		"null",
		"read",
		"write",
		"commit",
		"open",
		"open_confirm",
		"open_noattr",
		"open_downgrade",
		"close",
		"setattr",
		"fsinfo",
		"renew",
		"setclientid",
		"setclientid_confirm",
		"lock",
		"lockt",
		"locku",
		"access",
		"getattr",
		"lookup",
		"lookup_root",
		"remove",
		"rename",
		"link",
		"symlink",
		"create",
		"pathconf",
		"statfs",
		"readlink",
		"readdir",
		"server_caps",
		"delegreturn",
		"getacl",
		"setacl",
		"fs_locations",
		"release_lockowner",
		"secinfo",
		"fsid_present",
		"exchange_id",
		"create_session",
		"destroy_session",
		"sequence",
		"get_lease_time",
		"reclaim_complete",
		"layoutget",
		"getdeviceinfo",
		"layoutcommit",
		"layoutreturn",
		"secinfo_no_name",
		"test_stateid",
		"free_stateid",
		"getdevicelist",
		"bind_conn_to_session",
		"destroy_clientid",
		"seek",
		"allocate",
		"deallocate",
		"layoutstats",
		"clone",
		"copy",
		"offload_cancel",
		"lookupp",
		"layouterror",
		"copy_notify",
		"getxattr",
		"setxattr",
		"listxattrs",
		"removexattr",
	}
}

func getNFSOp(index int, nfsvers int) (string, error) {

	var nfsops []string
	switch nfsvers {
	case 3:
		nfsops = getNFSv3Ops()
	case 4:
		nfsops = getNFSv4Ops()
	default:
		return "", ErrorInvalidNFSVersion
	}

	// The operation index is defined by the slice with the operation
	// names. Catch potential programming errors if index is out of bounds
	if index < 0 || index >= len(nfsops) {
		return "", ErrorInvalidNFSOpIndex
	}

	return nfsops[index], nil
}

func parseCounters(s string) ([]int64, error) {
	ss := strings.Split(s, " ")
	sn := make([]int64, len(ss))
	for i, val := range ss {
		val_i, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			return []int64{}, err
		}
		if val_i < 0 {
			//This should never happen. Lines come from the kernel
			// and must not have negative or invalid numbers.
			return []int64{}, ErrorInvalidCounterValue
		}
		sn[i] = val_i
	}
	return sn, nil
}

func parseStatLine(line string, nfsvers int) ([]int64, error) {

	var s string

	switch nfsvers {
	case 3:
		s = line[len(prefix_nfs3):]
	case 4:
		s = line[len(prefix_nfs4):]
	default:
		return []int64{}, ErrorInvalidNFSVersion
	}

	stats, err := parseCounters(s)
	if err != nil {
		return []int64{}, err
	}
	return stats, nil
}

func processNFSstatLine(line string, nfsvers int) ([]NFSStatCounter, error) {

	var counters []NFSStatCounter
	nfsstat, err := parseStatLine(line, nfsvers)
	if err != nil {
		return []NFSStatCounter{}, err
	}

	for i, val := range nfsstat {
		op, err := getNFSOp(i, nfsvers)
		if err != nil {
			return []NFSStatCounter{}, err
		}
		counters = append(counters,
			NFSStatCounter{
				nfsvers: strconv.Itoa(nfsvers),
				op:      op,
				val:     val,
			})
	}

	return counters, nil
}

func processNFSStats(contents []string) ([]NFSStatCounter, error) {

	var counters []NFSStatCounter

	for _, line := range contents {
		nfsvers, isnfsstat := getNFSVers(line)
		if !isnfsstat {
			continue
		}
		lcounters, err := processNFSstatLine(line, nfsvers)
		if err != nil {
			return []NFSStatCounter{}, err
		}
		counters = append(counters, lcounters...)
	}
	return counters, nil
}

func assignStats(counters []NFSStatCounter, acc telegraf.Accumulator) error {

	var fields = make(map[string]interface{})
	for _, c := range counters {
		tags := map[string]string{
			"nfsvers": c.nfsvers,
			"op":      c.op,
		}
		fields["operations"] = c.val
		acc.AddFields("nfsstat", fields, tags)
	}

	return nil
}

func (n *NFSStat) Gather(acc telegraf.Accumulator) error {

	rawstats, err := ioutil.ReadFile("/proc/net/rpc/nfs")
	if err != nil {
		n.Log.Errorf("Failed to read /proc/net/rpc/nfs:  ", err)
		return err
	}

	counters, err := processNFSStats(strings.Split(string(rawstats), "\n"))
	if err != nil {
		return err
	}
	return assignStats(counters, acc)
}

func init() {
	inputs.Add("nfsstat", func() telegraf.Input {
		return &NFSStat{}
	})
}
