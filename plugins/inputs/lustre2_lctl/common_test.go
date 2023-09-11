//go:build linux

package lustre2_lctl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseJobstats(t *testing.T) {
	datav215 := `job_stats:
	- job_id:          yhcontrol.20575
	  snapshot_time:   1694133006
	  open:            { samples:           3, unit:  reqs }
	  close:           { samples:           0, unit:  reqs }
	  read_bytes:      { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }
	  write_bytes:     { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }
	- job_id:          1311947
	  snapshot_time:   1694133002
	  open:            { samples:           0, unit:  reqs }
	  getattr:         { samples:         174, unit:  reqs }
	  read_bytes:      { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }
	  write_bytes:     { samples:           0, unit:  reqs, min:       0, max:       0, sum:               0 }`

	datav217 := `job_stats:
	  - job_id:          thmc.0
	  snapshot_time   : 1896533.424837288 secs.nsecs
	  start_time      : 379104.213451005 secs.nsecs
	  elapsed_time    : 1517429.211386283 secs.nsecs
		open:            { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		statfs:          { samples:     9183627, unit: usecs, min:        0, max:     4387, sum:        115333492, sumsq:         8448670428 }
		read:            { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		write:           { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		read_bytes:      { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }
		write_bytes:     { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }
	  - job_id:          542307
	  snapshot_time   : 1958181.453162577 secs.nsecs
	  start_time      : 1958181.447941155 secs.nsecs
	  elapsed_time    : 0.005221422 secs.nsecs
	    open:            { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		getattr:         { samples:           6, unit: usecs, min:        0, max:       19, sum:               53, sumsq:                821 }
		read:            { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		write:           { samples:           0, unit: usecs, min:        0, max:        0, sum:                0, sumsq:                  0 }
		read_bytes:      { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }
		write_bytes:     { samples:           0, unit: bytes, min:        0, max:        0, sum:                0, sumsq:                  0 }`

	expectedv215 := make(map[string][]*Jobstat)
	expectedv215["yhcontrol.20575"] = []*Jobstat{
		{
			Operation: "open",
			Unit:      "reqs",
			Samples:   3,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "close",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "read_bytes",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write_bytes",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
	}

	expectedv215["1311947"] = []*Jobstat{
		{
			Operation: "open",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "getattr",
			Unit:      "reqs",
			Samples:   174,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "read_bytes",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write_bytes",
			Unit:      "reqs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
	}

	resultv215 := parseJobStats(datav215)

	if diff := cmp.Diff(expectedv215, resultv215, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}

	expectedv217 := make(map[string][]*Jobstat)
	expectedv217["thmc.0"] = []*Jobstat{
		{
			Operation: "open",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "statfs",
			Unit:      "usecs",
			Samples:   9183627,
			Min:       0,
			Max:       4387,
			Sum:       115333492,
			Sumsq:     8448670428,
		},
		{
			Operation: "read",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "read_bytes",
			Unit:      "bytes",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write_bytes",
			Unit:      "bytes",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
	}

	expectedv217["542307"] = []*Jobstat{
		{
			Operation: "open",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "getattr",
			Unit:      "usecs",
			Samples:   6,
			Min:       0,
			Max:       19,
			Sum:       53,
			Sumsq:     821,
		},
		{
			Operation: "read",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write",
			Unit:      "usecs",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "read_bytes",
			Unit:      "bytes",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
		{
			Operation: "write_bytes",
			Unit:      "bytes",
			Samples:   0,
			Min:       0,
			Max:       0,
			Sum:       0,
			Sumsq:     0,
		},
	}

	resultv217 := parseJobStats(datav217)

	if diff := cmp.Diff(expectedv217, resultv217, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}
}

func TestParseStats(t *testing.T) {

	dataMDTv215 := `snapshot_time             1694140455.278503266 secs.nsecs
	open                      137391283844 samples [reqs]
	close                     97376107699 samples [reqs] 1 1 97376107699`

	dataMDTv217 := `snapshot_time             1964295.787757337 secs.nsecs
	start_time                0.000000000 secs.nsecs
	elapsed_time              1964295.787757337 secs.nsecs
	open                      293759658 samples [usecs] 5 65525205 38223033629 34619065558020419
	close                     675832392 samples [usecs] 3 905437 15359059495 111766227630681`

	dataOSTv215 := `snapshot_time             1694140610.229551637 secs.nsecs
	read_bytes                1590792638 samples [bytes] 4096 4194304 612611321090048
	sync                      7713634 samples [reqs]`

	dataOSTv217 := `snapshot_time             1694140643.273314489 secs.nsecs
	write_bytes               8 samples [bytes] 167 202945 388227 58058611167
	read                      8449 samples [usecs] 2 385149 80065680 5387531243688`

	expectedMDT215 := []*Stat{
		{
			Operation: "open",
			Unit:      "reqs",
			Samples:   uint64(137391283844),
		},
		{
			Operation: "close",
			Unit:      "reqs",
			Samples:   uint64(97376107699),
			Min:       uint64(1),
			Max:       uint64(1),
			Sum:       uint64(97376107699),
		},
	}
	statsMDT215 := parseStats(dataMDTv215)

	if diff := cmp.Diff(expectedMDT215, statsMDT215, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}

	expectedMDT217 := []*Stat{
		{
			Operation: "open",
			Unit:      "usecs",
			Samples:   uint64(293759658),
			Min:       uint64(5),
			Max:       uint64(65525205),
			Sum:       uint64(38223033629),
			Sumsq:     uint64(34619065558020419),
		},
		{
			Operation: "close",
			Unit:      "usecs",
			Samples:   uint64(675832392),
			Min:       uint64(3),
			Max:       uint64(905437),
			Sum:       uint64(15359059495),
			Sumsq:     uint64(111766227630681),
		},
	}
	statsMDT217 := parseStats(dataMDTv217)
	if diff := cmp.Diff(expectedMDT217, statsMDT217, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}

	expectedOST215 := []*Stat{
		{
			Operation: "read_bytes",
			Unit:      "bytes",
			Samples:   uint64(1590792638),
			Min:       uint64(4096),
			Max:       uint64(4194304),
			Sum:       uint64(612611321090048),
		},
		{
			Operation: "sync",
			Unit:      "reqs",
			Samples:   uint64(7713634),
		},
	}

	statsOST215 := parseStats(dataOSTv215)
	if diff := cmp.Diff(expectedOST215, statsOST215, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}

	expectedOST217 := []*Stat{
		{
			Operation: "write_bytes",
			Unit:      "bytes",
			Samples:   uint64(8),
			Min:       uint64(167),
			Max:       uint64(202945),
			Sum:       uint64(388227),
			Sumsq:     uint64(58058611167),
		},
		{
			Operation: "read",
			Unit:      "usecs",
			Samples:   uint64(8449),
			Min:       uint64(2),
			Max:       uint64(385149),
			Sum:       uint64(80065680),
			Sumsq:     uint64(5387531243688),
		},
	}
	statsOST217 := parseStats(dataOSTv217)
	if diff := cmp.Diff(expectedOST217, statsOST217, nil); diff != "" {
		t.Fatalf("[]string\n--- expected\n+++ actual\n%s", diff)
	}
}
