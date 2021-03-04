// +build linux

package nfsstat

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestGetNFSVers(t *testing.T) {
	tests := []struct {
		input     string
		nfsvers   int
		isnfsstat bool
	}{
		{
			input:     "proc3 22 0",
			nfsvers:   3,
			isnfsstat: true,
		},
		{
			input:     "proc4 60 0",
			nfsvers:   4,
			isnfsstat: true,
		},
		{
			input:     "proc2 10 0",
			nfsvers:   0,
			isnfsstat: false,
		},
		{
			input:     "",
			nfsvers:   0,
			isnfsstat: false,
		},
	}
	for _, td := range tests {
		out_v, out_e := getNFSVers(td.input)
		assert.Equal(t, out_v, td.nfsvers)
		assert.Equal(t, out_e, td.isnfsstat)
	}
}
func TestGetNFSv3Ops(t *testing.T) {
	nfsv3ops := []string{
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
	assert.Equal(t, nfsv3ops, getNFSv3Ops())
}
func TestGetNFSv4Ops(t *testing.T) {
	nfsv4ops := []string{
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
	assert.Equal(t, nfsv4ops, getNFSv4Ops())
}
func TestGetNFSOp(t *testing.T) {
	tests := []struct {
		index   int
		nfsvers int
		opname  string
		err     error
	}{
		{
			index:   0,
			nfsvers: 0,
			opname:  "",
			err:     ErrorInvalidNFSVersion,
		},
		{
			index:   0,
			nfsvers: 2,
			opname:  "",
			err:     ErrorInvalidNFSVersion,
		},
		{
			index:   0,
			nfsvers: 5,
			opname:  "",
			err:     ErrorInvalidNFSVersion,
		},
		{
			index:   -1,
			nfsvers: 4,
			opname:  "",
			err:     ErrorInvalidNFSOpIndex,
		},
		{
			index:   22,
			nfsvers: 3,
			opname:  "",
			err:     ErrorInvalidNFSOpIndex,
		},
		{
			index:   21,
			nfsvers: 3,
			opname:  "commit",
			err:     nil,
		},
		{
			index:   20,
			nfsvers: 3,
			opname:  "pathconf",
			err:     nil,
		},
		{
			index:   11,
			nfsvers: 3,
			opname:  "mknod",
			err:     nil,
		},
		{
			index:   0,
			nfsvers: 3,
			opname:  "null",
			err:     nil,
		},
		{
			index:   0,
			nfsvers: 4,
			opname:  "null",
			err:     nil,
		},
		{
			index:   67,
			nfsvers: 4,
			opname:  "removexattr",
			err:     nil,
		},
		{
			index:   68,
			nfsvers: 4,
			opname:  "",
			err:     ErrorInvalidNFSOpIndex,
		},
	}
	for _, td := range tests {
		o, e := getNFSOp(td.index, td.nfsvers)
		assert.Equal(t, e, td.err)
		assert.Equal(t, o, td.opname)
	}
}
func TestParseCounters(t *testing.T) {
	tests := []struct {
		cstr string
		cslc []int64
		err  error
	}{
		{
			cstr: "0 0 0",
			cslc: []int64{0, 0, 0},
			err:  nil,
		},
		{
			cstr: "1 49020 0",
			cslc: []int64{1, 49020, 0},
			err:  nil,
		},
		{
			cstr: "invalid input",
			cslc: []int64{},
			err:  &strconv.NumError{},
		},
		{
			cstr: "0 0 -1",
			cslc: []int64{},
			err:  ErrorInvalidCounterValue,
		},
	}
	for _, td := range tests {
		o, e := parseCounters(td.cstr)
		assert.IsType(t, e, td.err)
		assert.Equal(t, o, td.cslc)
	}
}
func TestParseStatLine(t *testing.T) {
	tests := []struct {
		line    string
		nfsvers int
		output  []int64
		err     error
	}{
		{
			line:    "",
			nfsvers: 0,
			output:  []int64{},
			err:     ErrorInvalidNFSVersion,
		},
		{
			line:    "proc9",
			nfsvers: 0,
			output:  []int64{},
			err:     ErrorInvalidNFSVersion,
		},
	}
	for _, td := range tests {
		o, e := parseStatLine(td.line, td.nfsvers)
		assert.IsType(t, e, td.err)
		assert.Equal(t, o, td.output)
	}
}
func TestProcessNFSstatLine(t *testing.T) {
	tests := []struct {
		line    string
		nfsvers int
		output  []NFSStatCounter
		err     error
	}{
		{
			line:    "",
			nfsvers: 0,
			output:  []NFSStatCounter{},
			err:     ErrorInvalidNFSVersion,
		},
		{
			line:    "proc3 22 0 97338 0 2 5 0 0 0 0 0 0 0 0 0 0 0 0 10 1734 4 1 0",
			nfsvers: 3,
			output: []NFSStatCounter{
				{nfsvers: "3", op: "null", val: 0},
				{nfsvers: "3", op: "getattr", val: 97338},
				{nfsvers: "3", op: "setattr", val: 0},
				{nfsvers: "3", op: "lookup", val: 2},
				{nfsvers: "3", op: "access", val: 5},
				{nfsvers: "3", op: "readlink", val: 0},
				{nfsvers: "3", op: "read", val: 0},
				{nfsvers: "3", op: "write", val: 0},
				{nfsvers: "3", op: "create", val: 0},
				{nfsvers: "3", op: "mkdir", val: 0},
				{nfsvers: "3", op: "symlink", val: 0},
				{nfsvers: "3", op: "mknod", val: 0},
				{nfsvers: "3", op: "remove", val: 0},
				{nfsvers: "3", op: "rmdir", val: 0},
				{nfsvers: "3", op: "rename", val: 0},
				{nfsvers: "3", op: "link", val: 0},
				{nfsvers: "3", op: "readdir", val: 0},
				{nfsvers: "3", op: "readdirplus", val: 10},
				{nfsvers: "3", op: "fsstat", val: 1734},
				{nfsvers: "3", op: "fsinfo", val: 4},
				{nfsvers: "3", op: "pathconf", val: 1},
				{nfsvers: "3", op: "commit", val: 0},
			},
			err: nil,
		},
		{
			line:    "proc4 60 0 2306167 335769 0 57724 0 10386201 2 10311817 1752 4700 0 0 0 67 0 67 4993103 44647923 17084206 1657 1217 584 5 2 211 3043 16963 1058 3700237 7743 0 0 0 0 0 0 0 4513 901 899 116812 0 901 0 0 0 0 1657 0 67 0 0 899 0 0 0 0 0 0",
			nfsvers: 4,
			output: []NFSStatCounter{
				{nfsvers: "4", op: "null", val: 0},
				{nfsvers: "4", op: "read", val: 2306167},
				{nfsvers: "4", op: "write", val: 335769},
				{nfsvers: "4", op: "commit", val: 0},
				{nfsvers: "4", op: "open", val: 57724},
				{nfsvers: "4", op: "open_confirm", val: 0},
				{nfsvers: "4", op: "open_noattr", val: 10386201},
				{nfsvers: "4", op: "open_downgrade", val: 2},
				{nfsvers: "4", op: "close", val: 10311817},
				{nfsvers: "4", op: "setattr", val: 1752},
				{nfsvers: "4", op: "fsinfo", val: 4700},
				{nfsvers: "4", op: "renew", val: 0},
				{nfsvers: "4", op: "setclientid", val: 0},
				{nfsvers: "4", op: "setclientid_confirm", val: 0},
				{nfsvers: "4", op: "lock", val: 67},
				{nfsvers: "4", op: "lockt", val: 0},
				{nfsvers: "4", op: "locku", val: 67},
				{nfsvers: "4", op: "access", val: 4993103},
				{nfsvers: "4", op: "getattr", val: 44647923},
				{nfsvers: "4", op: "lookup", val: 17084206},
				{nfsvers: "4", op: "lookup_root", val: 1657},
				{nfsvers: "4", op: "remove", val: 1217},
				{nfsvers: "4", op: "rename", val: 584},
				{nfsvers: "4", op: "link", val: 5},
				{nfsvers: "4", op: "symlink", val: 2},
				{nfsvers: "4", op: "create", val: 211},
				{nfsvers: "4", op: "pathconf", val: 3043},
				{nfsvers: "4", op: "statfs", val: 16963},
				{nfsvers: "4", op: "readlink", val: 1058},
				{nfsvers: "4", op: "readdir", val: 3700237},
				{nfsvers: "4", op: "server_caps", val: 7743},
				{nfsvers: "4", op: "delegreturn", val: 0},
				{nfsvers: "4", op: "getacl", val: 0},
				{nfsvers: "4", op: "setacl", val: 0},
				{nfsvers: "4", op: "fs_locations", val: 0},
				{nfsvers: "4", op: "release_lockowner", val: 0},
				{nfsvers: "4", op: "secinfo", val: 0},
				{nfsvers: "4", op: "fsid_present", val: 0},
				{nfsvers: "4", op: "exchange_id", val: 4513},
				{nfsvers: "4", op: "create_session", val: 901},
				{nfsvers: "4", op: "destroy_session", val: 899},
				{nfsvers: "4", op: "sequence", val: 116812},
				{nfsvers: "4", op: "get_lease_time", val: 0},
				{nfsvers: "4", op: "reclaim_complete", val: 901},
				{nfsvers: "4", op: "layoutget", val: 0},
				{nfsvers: "4", op: "getdeviceinfo", val: 0},
				{nfsvers: "4", op: "layoutcommit", val: 0},
				{nfsvers: "4", op: "layoutreturn", val: 0},
				{nfsvers: "4", op: "secinfo_no_name", val: 1657},
				{nfsvers: "4", op: "test_stateid", val: 0},
				{nfsvers: "4", op: "free_stateid", val: 67},
				{nfsvers: "4", op: "getdevicelist", val: 0},
				{nfsvers: "4", op: "bind_conn_to_session", val: 0},
				{nfsvers: "4", op: "destroy_clientid", val: 899},
				{nfsvers: "4", op: "seek", val: 0},
				{nfsvers: "4", op: "allocate", val: 0},
				{nfsvers: "4", op: "deallocate", val: 0},
				{nfsvers: "4", op: "layoutstats", val: 0},
				{nfsvers: "4", op: "clone", val: 0},
				{nfsvers: "4", op: "copy", val: 0},
			},
			err: nil,
		},
	}
	for _, td := range tests {
		o, e := processNFSstatLine(td.line, td.nfsvers)
		assert.IsType(t, e, td.err)
		assert.Equal(t, o, td.output)
	}
}
func TestProcessNFSStats(t *testing.T) {
	tests := []struct {
		lines  []string
		output []NFSStatCounter
		err    error
	}{
		{
			lines:  []string{"", ""},
			output: nil,
			err:    nil,
		},
		{
			lines: []string{
				"rpc 0 0 0",
			},
			output: nil,
			err:    nil,
		},
		{
			lines: []string{
				"proc3 22 0 97338 0 2 5 0 0 0 0 0 0 0 0 0 0 0 0 10 1734 4 1 0",
				"proc4 60 0 2306167 335769 0 57724 0 10386201 2 10311817 1752 4700 0 0 0 67 0 67 4993103 44647923 17084206 1657 1217 584 5 2 211 3043 16963 1058 3700237 7743 0 0 0 0 0 0 0 4513 901 899 116812 0 901 0 0 0 0 1657 0 67 0 0 899 0 0 0 0 0 0",
			},
			output: []NFSStatCounter{
				{nfsvers: "3", op: "null", val: 0},
				{nfsvers: "3", op: "getattr", val: 97338},
				{nfsvers: "3", op: "setattr", val: 0},
				{nfsvers: "3", op: "lookup", val: 2},
				{nfsvers: "3", op: "access", val: 5},
				{nfsvers: "3", op: "readlink", val: 0},
				{nfsvers: "3", op: "read", val: 0},
				{nfsvers: "3", op: "write", val: 0},
				{nfsvers: "3", op: "create", val: 0},
				{nfsvers: "3", op: "mkdir", val: 0},
				{nfsvers: "3", op: "symlink", val: 0},
				{nfsvers: "3", op: "mknod", val: 0},
				{nfsvers: "3", op: "remove", val: 0},
				{nfsvers: "3", op: "rmdir", val: 0},
				{nfsvers: "3", op: "rename", val: 0},
				{nfsvers: "3", op: "link", val: 0},
				{nfsvers: "3", op: "readdir", val: 0},
				{nfsvers: "3", op: "readdirplus", val: 10},
				{nfsvers: "3", op: "fsstat", val: 1734},
				{nfsvers: "3", op: "fsinfo", val: 4},
				{nfsvers: "3", op: "pathconf", val: 1},
				{nfsvers: "3", op: "commit", val: 0},
				{nfsvers: "4", op: "null", val: 0},
				{nfsvers: "4", op: "read", val: 2306167},
				{nfsvers: "4", op: "write", val: 335769},
				{nfsvers: "4", op: "commit", val: 0},
				{nfsvers: "4", op: "open", val: 57724},
				{nfsvers: "4", op: "open_confirm", val: 0},
				{nfsvers: "4", op: "open_noattr", val: 10386201},
				{nfsvers: "4", op: "open_downgrade", val: 2},
				{nfsvers: "4", op: "close", val: 10311817},
				{nfsvers: "4", op: "setattr", val: 1752},
				{nfsvers: "4", op: "fsinfo", val: 4700},
				{nfsvers: "4", op: "renew", val: 0},
				{nfsvers: "4", op: "setclientid", val: 0},
				{nfsvers: "4", op: "setclientid_confirm", val: 0},
				{nfsvers: "4", op: "lock", val: 67},
				{nfsvers: "4", op: "lockt", val: 0},
				{nfsvers: "4", op: "locku", val: 67},
				{nfsvers: "4", op: "access", val: 4993103},
				{nfsvers: "4", op: "getattr", val: 44647923},
				{nfsvers: "4", op: "lookup", val: 17084206},
				{nfsvers: "4", op: "lookup_root", val: 1657},
				{nfsvers: "4", op: "remove", val: 1217},
				{nfsvers: "4", op: "rename", val: 584},
				{nfsvers: "4", op: "link", val: 5},
				{nfsvers: "4", op: "symlink", val: 2},
				{nfsvers: "4", op: "create", val: 211},
				{nfsvers: "4", op: "pathconf", val: 3043},
				{nfsvers: "4", op: "statfs", val: 16963},
				{nfsvers: "4", op: "readlink", val: 1058},
				{nfsvers: "4", op: "readdir", val: 3700237},
				{nfsvers: "4", op: "server_caps", val: 7743},
				{nfsvers: "4", op: "delegreturn", val: 0},
				{nfsvers: "4", op: "getacl", val: 0},
				{nfsvers: "4", op: "setacl", val: 0},
				{nfsvers: "4", op: "fs_locations", val: 0},
				{nfsvers: "4", op: "release_lockowner", val: 0},
				{nfsvers: "4", op: "secinfo", val: 0},
				{nfsvers: "4", op: "fsid_present", val: 0},
				{nfsvers: "4", op: "exchange_id", val: 4513},
				{nfsvers: "4", op: "create_session", val: 901},
				{nfsvers: "4", op: "destroy_session", val: 899},
				{nfsvers: "4", op: "sequence", val: 116812},
				{nfsvers: "4", op: "get_lease_time", val: 0},
				{nfsvers: "4", op: "reclaim_complete", val: 901},
				{nfsvers: "4", op: "layoutget", val: 0},
				{nfsvers: "4", op: "getdeviceinfo", val: 0},
				{nfsvers: "4", op: "layoutcommit", val: 0},
				{nfsvers: "4", op: "layoutreturn", val: 0},
				{nfsvers: "4", op: "secinfo_no_name", val: 1657},
				{nfsvers: "4", op: "test_stateid", val: 0},
				{nfsvers: "4", op: "free_stateid", val: 67},
				{nfsvers: "4", op: "getdevicelist", val: 0},
				{nfsvers: "4", op: "bind_conn_to_session", val: 0},
				{nfsvers: "4", op: "destroy_clientid", val: 899},
				{nfsvers: "4", op: "seek", val: 0},
				{nfsvers: "4", op: "allocate", val: 0},
				{nfsvers: "4", op: "deallocate", val: 0},
				{nfsvers: "4", op: "layoutstats", val: 0},
				{nfsvers: "4", op: "clone", val: 0},
				{nfsvers: "4", op: "copy", val: 0},
			},
			err: nil,
		},
	}
	for _, td := range tests {
		o, e := processNFSStats(td.lines)
		assert.IsType(t, e, td.err)
		assert.Equal(t, o, td.output)
	}
}

func TestAssignStats(t *testing.T) {
	tests := []struct {
		counters   []NFSStatCounter
		metricdata []struct {
			operations int64
			tags       map[string]string
		}
	}{
		{
			counters: []NFSStatCounter{
				{nfsvers: "3", op: "null", val: 0},
				{nfsvers: "3", op: "getattr", val: 97338},
				{nfsvers: "3", op: "setattr", val: 0},
				{nfsvers: "3", op: "lookup", val: 2},
				{nfsvers: "3", op: "access", val: 5},
				{nfsvers: "3", op: "readlink", val: 0},
				{nfsvers: "3", op: "read", val: 0},
				{nfsvers: "3", op: "write", val: 0},
				{nfsvers: "3", op: "create", val: 0},
				{nfsvers: "3", op: "mkdir", val: 0},
				{nfsvers: "3", op: "symlink", val: 0},
				{nfsvers: "3", op: "mknod", val: 0},
				{nfsvers: "3", op: "remove", val: 0},
				{nfsvers: "3", op: "rmdir", val: 0},
				{nfsvers: "3", op: "rename", val: 0},
				{nfsvers: "3", op: "link", val: 0},
				{nfsvers: "3", op: "readdir", val: 0},
				{nfsvers: "3", op: "readdirplus", val: 10},
				{nfsvers: "3", op: "fsstat", val: 1734},
				{nfsvers: "3", op: "fsinfo", val: 4},
				{nfsvers: "3", op: "pathconf", val: 1},
				{nfsvers: "3", op: "commit", val: 0},
				{nfsvers: "4", op: "null", val: 0},
				{nfsvers: "4", op: "read", val: 2306167},
				{nfsvers: "4", op: "write", val: 335769},
				{nfsvers: "4", op: "commit", val: 0},
				{nfsvers: "4", op: "open", val: 57724},
				{nfsvers: "4", op: "open_confirm", val: 0},
				{nfsvers: "4", op: "open_noattr", val: 10386201},
				{nfsvers: "4", op: "open_downgrade", val: 2},
				{nfsvers: "4", op: "close", val: 10311817},
				{nfsvers: "4", op: "setattr", val: 1752},
				{nfsvers: "4", op: "fsinfo", val: 4700},
				{nfsvers: "4", op: "renew", val: 0},
				{nfsvers: "4", op: "setclientid", val: 0},
				{nfsvers: "4", op: "setclientid_confirm", val: 0},
				{nfsvers: "4", op: "lock", val: 67},
				{nfsvers: "4", op: "lockt", val: 0},
				{nfsvers: "4", op: "locku", val: 67},
				{nfsvers: "4", op: "access", val: 4993103},
				{nfsvers: "4", op: "getattr", val: 44647923},
				{nfsvers: "4", op: "lookup", val: 17084206},
				{nfsvers: "4", op: "lookup_root", val: 1657},
				{nfsvers: "4", op: "remove", val: 1217},
				{nfsvers: "4", op: "rename", val: 584},
				{nfsvers: "4", op: "link", val: 5},
				{nfsvers: "4", op: "symlink", val: 2},
				{nfsvers: "4", op: "create", val: 211},
				{nfsvers: "4", op: "pathconf", val: 3043},
				{nfsvers: "4", op: "statfs", val: 16963},
				{nfsvers: "4", op: "readlink", val: 1058},
				{nfsvers: "4", op: "readdir", val: 3700237},
				{nfsvers: "4", op: "server_caps", val: 7743},
				{nfsvers: "4", op: "delegreturn", val: 0},
				{nfsvers: "4", op: "getacl", val: 0},
				{nfsvers: "4", op: "setacl", val: 0},
				{nfsvers: "4", op: "fs_locations", val: 0},
				{nfsvers: "4", op: "release_lockowner", val: 0},
				{nfsvers: "4", op: "secinfo", val: 0},
				{nfsvers: "4", op: "fsid_present", val: 0},
				{nfsvers: "4", op: "exchange_id", val: 4513},
				{nfsvers: "4", op: "create_session", val: 901},
				{nfsvers: "4", op: "destroy_session", val: 899},
				{nfsvers: "4", op: "sequence", val: 116812},
				{nfsvers: "4", op: "get_lease_time", val: 0},
				{nfsvers: "4", op: "reclaim_complete", val: 901},
				{nfsvers: "4", op: "layoutget", val: 0},
				{nfsvers: "4", op: "getdeviceinfo", val: 0},
				{nfsvers: "4", op: "layoutcommit", val: 0},
				{nfsvers: "4", op: "layoutreturn", val: 0},
				{nfsvers: "4", op: "secinfo_no_name", val: 1657},
				{nfsvers: "4", op: "test_stateid", val: 0},
				{nfsvers: "4", op: "free_stateid", val: 67},
				{nfsvers: "4", op: "getdevicelist", val: 0},
				{nfsvers: "4", op: "bind_conn_to_session", val: 0},
				{nfsvers: "4", op: "destroy_clientid", val: 899},
				{nfsvers: "4", op: "seek", val: 0},
				{nfsvers: "4", op: "allocate", val: 0},
				{nfsvers: "4", op: "deallocate", val: 0},
				{nfsvers: "4", op: "layoutstats", val: 0},
				{nfsvers: "4", op: "clone", val: 0},
				{nfsvers: "4", op: "copy", val: 0},
			},
			metricdata: []struct {
				operations int64
				tags       map[string]string
			}{
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "null"},
				},
				{
					operations: 97338,
					tags:       map[string]string{"nfsvers": "3", "op": "getattr"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "setattr"},
				},
				{
					operations: 2,
					tags:       map[string]string{"nfsvers": "3", "op": "lookup"},
				},
				{
					operations: 5,
					tags:       map[string]string{"nfsvers": "3", "op": "access"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "readlink"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "read"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "write"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "create"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "mkdir"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "symlink"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "mknod"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "remove"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "rmdir"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "rename"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "link"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "readdir"},
				},
				{
					operations: 10,
					tags:       map[string]string{"nfsvers": "3", "op": "readdirplus"},
				},
				{
					operations: 1734,
					tags:       map[string]string{"nfsvers": "3", "op": "fsstat"},
				},
				{
					operations: 4,
					tags:       map[string]string{"nfsvers": "3", "op": "fsinfo"},
				},
				{
					operations: 1,
					tags:       map[string]string{"nfsvers": "3", "op": "pathconf"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "3", "op": "commit"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "null"},
				},
				{
					operations: 2306167,
					tags:       map[string]string{"nfsvers": "4", "op": "read"},
				},
				{
					operations: 335769,
					tags:       map[string]string{"nfsvers": "4", "op": "write"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "commit"},
				},
				{
					operations: 57724,
					tags:       map[string]string{"nfsvers": "4", "op": "open"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "open_confirm"},
				},
				{
					operations: 10386201,
					tags:       map[string]string{"nfsvers": "4", "op": "open_noattr"},
				},
				{
					operations: 2,
					tags:       map[string]string{"nfsvers": "4", "op": "open_downgrade"},
				},
				{
					operations: 10311817,
					tags:       map[string]string{"nfsvers": "4", "op": "close"},
				},
				{
					operations: 1752,
					tags:       map[string]string{"nfsvers": "4", "op": "setattr"},
				},
				{
					operations: 4700,
					tags:       map[string]string{"nfsvers": "4", "op": "fsinfo"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "renew"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "setclientid"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "setclientid_confirm"},
				},
				{
					operations: 67,
					tags:       map[string]string{"nfsvers": "4", "op": "lock"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "lockt"},
				},
				{
					operations: 67,
					tags:       map[string]string{"nfsvers": "4", "op": "locku"},
				},
				{
					operations: 4993103,
					tags:       map[string]string{"nfsvers": "4", "op": "access"},
				},
				{
					operations: 44647923,
					tags:       map[string]string{"nfsvers": "4", "op": "getattr"},
				},
				{
					operations: 17084206,
					tags:       map[string]string{"nfsvers": "4", "op": "lookup"},
				},
				{
					operations: 1657,
					tags:       map[string]string{"nfsvers": "4", "op": "lookup_root"},
				},
				{
					operations: 1217,
					tags:       map[string]string{"nfsvers": "4", "op": "remove"},
				},
				{
					operations: 584,
					tags:       map[string]string{"nfsvers": "4", "op": "rename"},
				},
				{
					operations: 5,
					tags:       map[string]string{"nfsvers": "4", "op": "link"},
				},
				{
					operations: 2,
					tags:       map[string]string{"nfsvers": "4", "op": "symlink"},
				},
				{
					operations: 211,
					tags:       map[string]string{"nfsvers": "4", "op": "create"},
				},
				{
					operations: 3043,
					tags:       map[string]string{"nfsvers": "4", "op": "pathconf"},
				},
				{
					operations: 16963,
					tags:       map[string]string{"nfsvers": "4", "op": "statfs"},
				},
				{
					operations: 1058,
					tags:       map[string]string{"nfsvers": "4", "op": "readlink"},
				},
				{
					operations: 3700237,
					tags:       map[string]string{"nfsvers": "4", "op": "readdir"},
				},
				{
					operations: 7743,
					tags:       map[string]string{"nfsvers": "4", "op": "server_caps"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "delegreturn"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "getacl"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "setacl"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "fs_locations"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "release_lockowner"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "secinfo"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "fsid_present"},
				},
				{
					operations: 4513,
					tags:       map[string]string{"nfsvers": "4", "op": "exchange_id"},
				},
				{
					operations: 901,
					tags:       map[string]string{"nfsvers": "4", "op": "create_session"},
				},
				{
					operations: 899,
					tags:       map[string]string{"nfsvers": "4", "op": "destroy_session"},
				},
				{
					operations: 116812,
					tags:       map[string]string{"nfsvers": "4", "op": "sequence"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "get_lease_time"},
				},
				{
					operations: 901,
					tags:       map[string]string{"nfsvers": "4", "op": "reclaim_complete"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "layoutget"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "getdeviceinfo"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "layoutcommit"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "layoutreturn"},
				},
				{
					operations: 1657,
					tags:       map[string]string{"nfsvers": "4", "op": "secinfo_no_name"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "test_stateid"},
				},
				{
					operations: 67,
					tags:       map[string]string{"nfsvers": "4", "op": "free_stateid"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "getdevicelist"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "bind_conn_to_session"},
				},
				{
					operations: 899,
					tags:       map[string]string{"nfsvers": "4", "op": "destroy_clientid"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "seek"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "allocate"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "deallocate"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "layoutstats"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "clone"},
				},
				{
					operations: 0,
					tags:       map[string]string{"nfsvers": "4", "op": "copy"},
				},
			},
		},
	}
	for _, td := range tests {
		var acc testutil.Accumulator
		err := assignStats(td.counters, &acc)
		assert.Nil(t, err)
		for _, md := range td.metricdata {
			acc.AssertContainsTaggedFields(t, "nfsstat", map[string]interface{}{"operations": md.operations}, md.tags)
		}
	}
}
