package dovecot

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/textproto"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDovecotIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	var acc testutil.Accumulator

	// Test type=global server=unix
	addr := "/tmp/socket"
	wait := make(chan int)
	go func() {
		defer close(wait)

		la, err := net.ResolveUnixAddr("unix", addr)
		require.NoError(t, err)

		l, err := net.ListenUnix("unix", la)
		require.NoError(t, err)
		defer l.Close()
		defer os.Remove(addr)

		wait <- 0
		conn, err := l.Accept()
		require.NoError(t, err)
		defer conn.Close()

		readertp := textproto.NewReader(bufio.NewReader(conn))
		_, err = readertp.ReadLine()
		require.NoError(t, err)

		buf := bytes.NewBufferString(sampleGlobal)
		_, err = io.Copy(conn, buf)
		require.NoError(t, err)
	}()

	// Wait for server to start
	<-wait

	d := &Dovecot{Servers: []string{addr}, Type: "global"}
	err := d.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{"server": addr, "type": "global"}
	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)

	// Test type=global
	tags = map[string]string{"server": "dovecot.test", "type": "global"}
	buf := bytes.NewBufferString(sampleGlobal)

	err = gatherStats(buf, &acc, "dovecot.test", "global")
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)

	// Test type=domain
	tags = map[string]string{"server": "dovecot.test", "type": "domain", "domain": "domain.test"}
	buf = bytes.NewBufferString(sampleDomain)

	err = gatherStats(buf, &acc, "dovecot.test", "domain")
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)

	// Test type=ip
	tags = map[string]string{"server": "dovecot.test", "type": "ip", "ip": "192.168.0.100"}
	buf = bytes.NewBufferString(sampleIP)

	err = gatherStats(buf, &acc, "dovecot.test", "ip")
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)

	// Test type=user
	fields = map[string]interface{}{
		"reset_timestamp":  time.Unix(1453969886, 0),
		"last_update":      time.Unix(1454603963, 39864),
		"num_logins":       int64(7503897),
		"num_cmds":         int64(52595715),
		"user_cpu":         1.00831175372e+08,
		"sys_cpu":          8.3849071112e+07,
		"clock_time":       4.3260019315281835e+15,
		"min_faults":       int64(763950011),
		"maj_faults":       int64(1112443),
		"vol_cs":           int64(4120386897),
		"invol_cs":         int64(3685239306),
		"disk_input":       int64(41679480946688),
		"disk_output":      int64(1819070669176832),
		"read_count":       int64(2368906465),
		"read_bytes":       int64(2957928122981169),
		"write_count":      int64(3545389615),
		"write_bytes":      int64(1666822498251286),
		"mail_lookup_path": int64(24396105),
		"mail_lookup_attr": int64(302845),
		"mail_read_count":  int64(20155768),
		"mail_read_bytes":  int64(669946617705),
		"mail_cache_hits":  int64(1557255080),
	}

	tags = map[string]string{"server": "dovecot.test", "type": "user", "user": "user.1@domain.test"}
	buf = bytes.NewBufferString(sampleUser)

	err = gatherStats(buf, &acc, "dovecot.test", "user")
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t, "dovecot", fields, tags)
}

const sampleGlobal = `reset_timestamp	last_update	num_logins	num_cmds	num_connected_sessions	user_cpu	sys_cpu	clock_time	min_faults	maj_faults	vol_cs	invol_cs	disk_input	disk_output	read_count	read_bytes	write_count	write_bytes	mail_lookup_path	mail_lookup_attr	mail_read_count	mail_read_bytes	mail_cache_hits
1453969886	1454603963.039864	7503897	52595715	1204	100831175.372000	83849071.112000	4326001931528183.495762	763950011	1112443	4120386897	3685239306	41679480946688	1819070669176832	2368906465	2957928122981169	3545389615	1666822498251286	24396105	302845	20155768	669946617705	1557255080`

const sampleDomain = `domain	reset_timestamp	last_update	num_logins	num_cmds	num_connected_sessions	user_cpu	sys_cpu	clock_time	min_faults	maj_faults	vol_cs	invol_cs	disk_input	disk_output	read_count	read_bytes	write_count	write_bytes	mail_lookup_path	mail_lookup_attr	mail_read_count	mail_read_bytes	mail_cache_hits
domain.test	1453969886	1454603963.039864	7503897	52595715	1204	100831175.372000	83849071.112000	4326001931528183.495762	763950011	1112443	4120386897	3685239306	41679480946688	1819070669176832	2368906465	2957928122981169	3545389615	1666822498251286	24396105	302845	20155768	669946617705	1557255080`

const sampleIP = `ip	reset_timestamp	last_update	num_logins	num_cmds	num_connected_sessions	user_cpu	sys_cpu	clock_time	min_faults	maj_faults	vol_cs	invol_cs	disk_input	disk_output	read_count	read_bytes	write_count	write_bytes	mail_lookup_path	mail_lookup_attr	mail_read_count	mail_read_bytes	mail_cache_hits
192.168.0.100	1453969886	1454603963.039864	7503897	52595715	1204	100831175.372000	83849071.112000	4326001931528183.495762	763950011	1112443	4120386897	3685239306	41679480946688	1819070669176832	2368906465	2957928122981169	3545389615	1666822498251286	24396105	302845	20155768	669946617705	1557255080`

const sampleUser = `user	reset_timestamp	last_update	num_logins	num_cmds	user_cpu	sys_cpu	clock_time	min_faults	maj_faults	vol_cs	invol_cs	disk_input	disk_output	read_count	read_bytes	write_count	write_bytes	mail_lookup_path	mail_lookup_attr	mail_read_count	mail_read_bytes	mail_cache_hits
user.1@domain.test	1453969886	1454603963.039864	7503897	52595715	100831175.372000	83849071.112000	4326001931528183.495762	763950011	1112443	4120386897	3685239306	41679480946688	1819070669176832	2368906465	2957928122981169	3545389615	1666822498251286	24396105	302845	20155768	669946617705	1557255080`
