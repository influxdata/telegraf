package dovecot

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDovecot(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	var acc testutil.Accumulator
	tags := map[string]string{"server": "dovecot.test", "domain": "domain.test"}
	buf := bytes.NewBufferString(sampleStats)

	var doms = map[string]bool{
		"domain.test": true,
	}

	err := gatherStats(buf, &acc, doms, "dovecot.test")
	require.NoError(t, err)

	fields := map[string]interface{}{
		"reset_timestamp":        time.Unix(1453969886, 0),
		"last_update":            time.Unix(1454603963, 39864),
		"num_logins":             int64(7503897),
		"num_cmds":               int64(52595715),
		"num_connected_sessions": int64(1204),
		"user_cpu":               1.00831175372e+08,
		"sys_cpu":                8.3849071112e+07,
		"clock_time":             4.3260019315281835e+15,
		"min_faults":             int64(763950011),
		"maj_faults":             int64(1112443),
		"vol_cs":                 int64(4120386897),
		"invol_cs":               int64(3685239306),
		"disk_input":             int64(41679480946688),
		"disk_output":            int64(1819070669176832),
		"read_count":             int64(2368906465),
		"read_bytes":             int64(2957928122981169),
		"write_count":            int64(3545389615),
		"write_bytes":            int64(1666822498251286),
		"mail_lookup_path":       int64(24396105),
		"mail_lookup_attr":       int64(302845),
		"mail_read_count":        int64(20155768),
		"mail_read_bytes":        int64(669946617705),
		"mail_cache_hits":        int64(1557255080),
	}

	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)

}

const sampleStats = `domain	reset_timestamp	last_update	num_logins	num_cmds	num_connected_sessions	user_cpu	sys_cpu	clock_time	min_faults	maj_faults	vol_cs	invol_cs	disk_input	disk_output	read_count	read_bytes	write_count	write_bytes	mail_lookup_path	mail_lookup_attr	mail_read_count	mail_read_bytes	mail_cache_hits
domain.bad	1453970076	1454603947.383029	10749	33828	0	177988.524000	148071.772000	7531838964717.193706	212491179	2125	2190386067	112779200	74487934976	3221808119808	2469948401	5237602841760	1091171292	2951966459802	15363	0	2922	136403379	334372
domain.test	1453969886	1454603963.039864	7503897	52595715	1204	100831175.372000	83849071.112000	4326001931528183.495762	763950011	1112443	4120386897	3685239306	41679480946688	1819070669176832	2368906465	2957928122981169	3545389615	1666822498251286	24396105	302845	20155768	669946617705	1557255080`
