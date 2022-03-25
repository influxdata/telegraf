package syslogng

import (
	"bytes"
	"testing"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func SyslogNgCtlRunnerTest(output string) func(string, config.Duration, bool) (*bytes.Buffer, error) {
	return func(string, config.Duration, bool) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestParseFullOutput(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &SyslogNg{
		run: SyslogNgCtlRunnerTest(fullOutput),
	}
	err := v.Gather(acc)

	require.NoError(t, err)
	require.True(t, acc.HasMeasurement("syslog-ng"))

	require.Len(t, acc.Metrics, 26)
	require.Equal(t, 26, acc.NFields())

	acc.AssertContainsFields(t, "syslog-ng", parsedFullOutput)
}

var parsedFullOutput = map[string]interface{}{
	"number": 0,
}

var fullOutput = `SourceName;SourceId;SourceInstance;State;Type;Number
destination;xferlog;;a;processed;0
destination;security;;a;processed;0
src.internal;src#4;;a;processed;23
src.internal;src#4;;a;stamp;1648129533
destination;debuglog;;a;processed;82
global;msg_clones;;a;processed;3
destination;cron;;a;processed;64
global;internal_source;;a;dropped;0
global;internal_source;;a;queued;0
global;sdata_updates;;a;processed;1530016
destination;d_foo2;;a;processed;1530016
destination;authlog;;a;processed;3
destination;messages;;a;processed;22
global;scratch_buffers_count;;a;queued;9
destination;console;;a;processed;0
center;;queued;a;processed;1530187
src.syslog;s_foo;afsocket_sd.(stream,AF_INET(127.0.0.1:6514));a;connections;0
destination;slip;;a;processed;0
global;payload_reallocs;;a;processed;43
destination;ppp;;a;processed;0
destination;lpd-errs;;a;processed;0
destination;allusers;;a;processed;0
center;;received;a;processed;1530280
destination;maillog;;a;processed;0
global;internal_queue_length;;a;processed;0
source;src;;a;processed;264
source;s_foo;;a;processed;1530016
global;scratch_buffers_bytes;;a;queued;512`
