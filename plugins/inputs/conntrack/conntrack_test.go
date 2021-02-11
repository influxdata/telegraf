// +build linux

package conntrack

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func restoreDflts(savedFiles, savedDirs []string) {
	dfltFiles = savedFiles
	dfltDirs = savedDirs
}

func TestNoFilesFound(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)

	dfltFiles = []string{"baz.txt"}
	dfltDirs = []string{"./foo/bar"}
	c := &Conntrack{}
	acc := &testutil.Accumulator{}
	err := c.Gather(acc)

	assert.EqualError(t, err, "Conntrack input failed to collect metrics. "+
		"Is the conntrack kernel module loaded?")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := ioutil.TempFile(tmpdir, "ip_conntrack_count")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dfltDirs = []string{tmpdir}
	fname := path.Base(tmpFile.Name())
	dfltFiles = []string{fname}

	count := 1234321
	ioutil.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(count)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)
	acc.AssertContainsFields(t, inputName, map[string]interface{}{
		fname: float64(count)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	cntFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_count")
	maxFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_max")
	assert.NoError(t, err)
	defer os.Remove(cntFile.Name())
	defer os.Remove(maxFile.Name())

	dfltDirs = []string{tmpdir}
	cntFname := path.Base(cntFile.Name())
	maxFname := path.Base(maxFile.Name())
	dfltFiles = []string{cntFname, maxFname}

	count := 1234321
	max := 9999999
	ioutil.WriteFile(cntFile.Name(), []byte(strconv.Itoa(count)), 0660)
	ioutil.WriteFile(maxFile.Name(), []byte(strconv.Itoa(max)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)

	fix := func(s string) string {
		return strings.Replace(s, "nf_", "ip_", 1)
	}

	acc.AssertContainsFields(t, inputName,
		map[string]interface{}{
			fix(cntFname): float64(count),
			fix(maxFname): float64(max),
		})
}

func TestNfConntrackParse(t *testing.T) {
	fakeRow := []byte(`ipv4     2 udp      17 19 src=192.168.0.230 dst=8.8.8.8 sport=37421 dport=123 [UNREPLIED] src=8.8.8.8 dst=10.255.244.37 sport=123 dport=31224 mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 udp      17 19 src=192.168.0.181 dst=8.8.8.8 sport=56467 dport=123 src=8.8.8.8 dst=10.255.244.37 sport=123 dport=55189 mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 7 CLOSE src=192.168.0.72 dst=8.8.8.8 sport=28426 dport=443 src=8.8.8.8 dst=10.255.244.37 sport=443 dport=50394 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1 TIME_WAIT src=192.168.0.221 dst=8.8.8.8 sport=30596 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=47976 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1198 ESTABLISHED src=192.168.0.72 dst=8.8.8.8 sport=20162 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=1410 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 9 TIME_WAIT src=192.168.0.221 dst=8.8.8.8 sport=13396 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=48534 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1199 ESTABLISHED src=192.168.0.221 dst=8.8.8.8 sport=33444 dport=443 src=8.8.8.8 dst=10.255.244.37 sport=443 dport=6168 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1199 ESTABLISHED src=192.168.0.230 dst=8.8.8.8 sport=31540 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=22071 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 0 CLOSE src=192.168.0.221 dst=8.8.8.8 sport=46302 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=43508 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 7 TIME_WAIT src=192.168.0.19 dst=8.8.8.8 sport=14532 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=46470 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 4 TIME_WAIT src=192.168.0.221 dst=8.8.8.8 sport=11120 dport=80 src=8.8.8.8 dst=10.255.244.37 sport=80 dport=39246 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1191 ESTABLISHED src=192.168.0.221 dst=8.8.8.8 sport=55518 dport=443 src=8.8.8.8 dst=10.255.244.37 sport=443 dport=44639 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 1199 ESTABLISHED src=192.168.0.221 dst=8.8.8.8 sport=10934 dport=443 src=8.8.8.8 dst=10.255.244.37 sport=443 dport=31644 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2
ipv4     2 tcp      6 8 CLOSE src=192.168.0.221 dst=8.8.8.8 sport=49746 dport=5454 src=8.8.8.8 dst=10.255.244.37 sport=5454 dport=44524 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 zone=0 use=2`)

	expected := map[string]int64{
		"tcp_close":       2,
		"tcp_established": 5,
		"tcp_time_wait":   4,
		"udp_unreplied":   1,
		"udp":             1,
	}

	nf := newNfConntrack()
	io.Copy(nf, bytes.NewReader(fakeRow))
	if !assert.ObjectsAreEqualValues(expected, nf.counters) {
		t.Error("Invalid result in the parser")
	}
}
