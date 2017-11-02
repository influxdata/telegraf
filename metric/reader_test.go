package metric

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkMetricReader(b *testing.B) {
	metrics := make([]telegraf.Metric, 10)
	for i := 0; i < 10; i++ {
		metrics[i], _ = New("foo", map[string]string{},
			map[string]interface{}{"value": int64(1)}, time.Now())
	}
	for n := 0; n < b.N; n++ {
		r := NewReader(metrics)
		io.Copy(ioutil.Discard, r)
	}
}

func TestMetricReader(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	metrics := make([]telegraf.Metric, 10)
	for i := 0; i < 10; i++ {
		metrics[i], _ = New("foo", map[string]string{},
			map[string]interface{}{"value": int64(1)}, ts)
	}

	r := NewReader(metrics)

	buf := make([]byte, 35)
	for i := 0; i < 10; i++ {
		n, err := r.Read(buf)
		if err != nil {
			assert.True(t, err == io.EOF, err.Error())
		}
		assert.Equal(t, 33, n)
		assert.Equal(t, "foo value=1i 1481032190000000000\n", string(buf[0:n]))
	}

	// reader should now be done, and always return 0, io.EOF
	for i := 0; i < 10; i++ {
		n, err := r.Read(buf)
		assert.True(t, err == io.EOF, err.Error())
		assert.Equal(t, 0, n)
	}
}

func TestMetricReader_OverflowMetric(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m, _ := New("foo", map[string]string{},
		map[string]interface{}{"value": int64(10)}, ts)
	metrics := []telegraf.Metric{m}

	r := NewReader(metrics)
	buf := make([]byte, 5)

	tests := []struct {
		exp string
		err error
		n   int
	}{
		{
			"foo v",
			nil,
			5,
		},
		{
			"alue=",
			nil,
			5,
		},
		{
			"10i 1",
			nil,
			5,
		},
		{
			"48103",
			nil,
			5,
		},
		{
			"21900",
			nil,
			5,
		},
		{
			"00000",
			nil,
			5,
		},
		{
			"000\n",
			io.EOF,
			4,
		},
		{
			"",
			io.EOF,
			0,
		},
	}

	for _, test := range tests {
		n, err := r.Read(buf)
		assert.Equal(t, test.n, n)
		assert.Equal(t, test.exp, string(buf[0:n]))
		assert.Equal(t, test.err, err)
	}
}

// Regression test for when a metric is the same size as the buffer.
//
// Previously EOF would not be set until the next call to Read.
func TestMetricReader_MetricSizeEqualsBufferSize(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{"a": int64(1)}, ts)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, m1.Len())

	for {
		n, err := r.Read(buf)
		// Should never read 0 bytes unless at EOF, unless input buffer is 0 length
		if n == 0 {
			require.Equal(t, io.EOF, err)
			break
		}
		// Lines should be terminated with a LF
		if err == io.EOF {
			require.Equal(t, uint8('\n'), buf[n-1])
			break
		}
		require.NoError(t, err)
	}
}

// Regression test for when a metric requires to be split and one of the
// split metrics is exactly the size of the buffer.
//
// Previously an empty string would be returned on the next Read without error,
// and then next Read call would panic.
func TestMetricReader_SplitWithExactLengthSplit(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{"a": int64(1), "bb": int64(2)}, ts)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, 30)

	//  foo a=1i,bb=2i 1481032190000000000\n // len 35
	//
	// Requires this specific split order:
	//  foo a=1i 1481032190000000000\n  // len 29
	//  foo bb=2i 1481032190000000000\n // len 30

	for {
		n, err := r.Read(buf)
		// Should never read 0 bytes unless at EOF, unless input buffer is 0 length
		if n == 0 {
			require.Equal(t, io.EOF, err)
			break
		}
		// Lines should be terminated with a LF
		if err == io.EOF {
			require.Equal(t, uint8('\n'), buf[n-1])
			break
		}
		require.NoError(t, err)
	}
}

// Regression test for when a metric requires to be split and one of the
// split metrics is larger than the buffer.
//
// Previously the metric index would be set incorrectly causing a panic.
func TestMetricReader_SplitOverflowOversized(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"a":   int64(1),
			"bbb": int64(2),
		}, ts)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, 30)

	// foo a=1i,bbb=2i 1481032190000000000\n // len 36
	//
	// foo a=1i 1481032190000000000\n  // len 29
	// foo bbb=2i 1481032190000000000\n // len 31

	for {
		n, err := r.Read(buf)
		// Should never read 0 bytes unless at EOF, unless input buffer is 0 length
		if n == 0 {
			require.Equal(t, io.EOF, err)
			break
		}
		// Lines should be terminated with a LF
		if err == io.EOF {
			require.Equal(t, uint8('\n'), buf[n-1])
			break
		}
		require.NoError(t, err)
	}
}

// Regression test for when a split metric exactly fits in the buffer.
//
// Previously the metric would be overflow split when not required.
func TestMetricReader_SplitOverflowUneeded(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{"a": int64(1), "b": int64(2)}, ts)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, 29)

	// foo a=1i,b=2i 1481032190000000000\n // len 34
	//
	// foo a=1i 1481032190000000000\n  // len 29
	// foo b=2i 1481032190000000000\n // len 29

	for {
		n, err := r.Read(buf)
		// Should never read 0 bytes unless at EOF, unless input buffer is 0 length
		if n == 0 {
			require.Equal(t, io.EOF, err)
			break
		}
		// Lines should be terminated with a LF
		if err == io.EOF {
			require.Equal(t, uint8('\n'), buf[n-1])
			break
		}
		require.NoError(t, err)
	}
}

func TestMetricReader_OverflowMultipleMetrics(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m, _ := New("foo", map[string]string{},
		map[string]interface{}{"value": int64(10)}, ts)
	metrics := []telegraf.Metric{m, m.Copy()}

	r := NewReader(metrics)
	buf := make([]byte, 10)

	tests := []struct {
		exp string
		err error
		n   int
	}{
		{
			"foo value=",
			nil,
			10,
		},
		{
			"10i 148103",
			nil,
			10,
		},
		{
			"2190000000",
			nil,
			10,
		},
		{
			"000\n",
			nil,
			4,
		},
		{
			"foo value=",
			nil,
			10,
		},
		{
			"10i 148103",
			nil,
			10,
		},
		{
			"2190000000",
			nil,
			10,
		},
		{
			"000\n",
			io.EOF,
			4,
		},
		{
			"",
			io.EOF,
			0,
		},
	}

	for _, test := range tests {
		n, err := r.Read(buf)
		assert.Equal(t, test.n, n)
		assert.Equal(t, test.exp, string(buf[0:n]))
		assert.Equal(t, test.err, err)
	}
}

// test splitting a metric
func TestMetricReader_SplitMetric(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
			"value2": int64(10),
			"value3": int64(10),
			"value4": int64(10),
			"value5": int64(10),
			"value6": int64(10),
		},
		ts,
	)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, 60)

	tests := []struct {
		expRegex string
		err      error
		n        int
	}{
		{
			`foo value\d=10i,value\d=10i,value\d=10i 1481032190000000000\n`,
			nil,
			57,
		},
		{
			`foo value\d=10i,value\d=10i,value\d=10i 1481032190000000000\n`,
			io.EOF,
			57,
		},
		{
			"",
			io.EOF,
			0,
		},
	}

	for _, test := range tests {
		n, err := r.Read(buf)
		assert.Equal(t, test.n, n)
		re := regexp.MustCompile(test.expRegex)
		assert.True(t, re.MatchString(string(buf[0:n])), string(buf[0:n]))
		assert.Equal(t, test.err, err)
	}
}

// test an array with one split metric and one unsplit
func TestMetricReader_SplitMetric2(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
			"value2": int64(10),
			"value3": int64(10),
			"value4": int64(10),
			"value5": int64(10),
			"value6": int64(10),
		},
		ts,
	)
	m2, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
		},
		ts,
	)
	metrics := []telegraf.Metric{m1, m2}

	r := NewReader(metrics)
	buf := make([]byte, 60)

	tests := []struct {
		expRegex string
		err      error
		n        int
	}{
		{
			`foo value\d=10i,value\d=10i,value\d=10i 1481032190000000000\n`,
			nil,
			57,
		},
		{
			`foo value\d=10i,value\d=10i,value\d=10i 1481032190000000000\n`,
			nil,
			57,
		},
		{
			`foo value1=10i 1481032190000000000\n`,
			io.EOF,
			35,
		},
		{
			"",
			io.EOF,
			0,
		},
	}

	for _, test := range tests {
		n, err := r.Read(buf)
		assert.Equal(t, test.n, n)
		re := regexp.MustCompile(test.expRegex)
		assert.True(t, re.MatchString(string(buf[0:n])), string(buf[0:n]))
		assert.Equal(t, test.err, err)
	}
}

// test split that results in metrics that are still too long, which results in
// the reader falling back to regular overflow.
func TestMetricReader_SplitMetricTooLong(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
			"value2": int64(10),
		},
		ts,
	)
	metrics := []telegraf.Metric{m1}

	r := NewReader(metrics)
	buf := make([]byte, 30)

	tests := []struct {
		expRegex string
		err      error
		n        int
	}{
		{
			`foo value\d=10i,value\d=10i 1481`,
			nil,
			30,
		},
		{
			`032190000000000\n`,
			io.EOF,
			16,
		},
		{
			"",
			io.EOF,
			0,
		},
	}

	for _, test := range tests {
		n, err := r.Read(buf)
		assert.Equal(t, test.n, n)
		re := regexp.MustCompile(test.expRegex)
		assert.True(t, re.MatchString(string(buf[0:n])), string(buf[0:n]))
		assert.Equal(t, test.err, err)
	}
}

// test split with a changing buffer size in the middle of subsequent calls
// to Read
func TestMetricReader_SplitMetricChangingBuffer(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
			"value2": int64(10),
			"value3": int64(10),
		},
		ts,
	)
	m2, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
		},
		ts,
	)
	metrics := []telegraf.Metric{m1, m2}

	r := NewReader(metrics)

	tests := []struct {
		expRegex string
		err      error
		n        int
		buf      []byte
	}{
		{
			`foo value\d=10i 1481032190000000000\n`,
			nil,
			35,
			make([]byte, 36),
		},
		{
			`foo value\d=10i 148103219000000`,
			nil,
			30,
			make([]byte, 30),
		},
		{
			`0000\n`,
			nil,
			5,
			make([]byte, 30),
		},
		{
			`foo value\d=10i 1481032190000000000\n`,
			nil,
			35,
			make([]byte, 36),
		},
		{
			`foo value1=10i 1481032190000000000\n`,
			io.EOF,
			35,
			make([]byte, 36),
		},
		{
			"",
			io.EOF,
			0,
			make([]byte, 36),
		},
	}

	for _, test := range tests {
		n, err := r.Read(test.buf)
		assert.Equal(t, test.n, n, test.expRegex)
		re := regexp.MustCompile(test.expRegex)
		assert.True(t, re.MatchString(string(test.buf[0:n])), string(test.buf[0:n]))
		assert.Equal(t, test.err, err, test.expRegex)
	}
}

// test split with a changing buffer size in the middle of subsequent calls
// to Read
func TestMetricReader_SplitMetricChangingBuffer2(t *testing.T) {
	ts := time.Unix(1481032190, 0)
	m1, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
			"value2": int64(10),
		},
		ts,
	)
	m2, _ := New("foo", map[string]string{},
		map[string]interface{}{
			"value1": int64(10),
		},
		ts,
	)
	metrics := []telegraf.Metric{m1, m2}

	r := NewReader(metrics)

	tests := []struct {
		expRegex string
		err      error
		n        int
		buf      []byte
	}{
		{
			`foo value\d=10i 1481032190000000000\n`,
			nil,
			35,
			make([]byte, 36),
		},
		{
			`foo value\d=10i 148103219000000`,
			nil,
			30,
			make([]byte, 30),
		},
		{
			`0000\n`,
			nil,
			5,
			make([]byte, 30),
		},
		{
			`foo value1=10i 1481032190000000000\n`,
			io.EOF,
			35,
			make([]byte, 36),
		},
		{
			"",
			io.EOF,
			0,
			make([]byte, 36),
		},
	}

	for _, test := range tests {
		n, err := r.Read(test.buf)
		assert.Equal(t, test.n, n, test.expRegex)
		re := regexp.MustCompile(test.expRegex)
		assert.True(t, re.MatchString(string(test.buf[0:n])), string(test.buf[0:n]))
		assert.Equal(t, test.err, err, test.expRegex)
	}
}

func TestReader_Read(t *testing.T) {
	epoch := time.Unix(0, 0)

	type args struct {
		name   string
		tags   map[string]string
		fields map[string]interface{}
		t      time.Time
		mType  []telegraf.ValueType
	}
	tests := []struct {
		name     string
		args     args
		expected []byte
	}{
		{
			name: "escape backslashes in string field",
			args: args{
				name:   "cpu",
				tags:   map[string]string{},
				fields: map[string]interface{}{"value": `test\`},
				t:      epoch,
			},
			expected: []byte(`cpu value="test\\" 0`),
		},
		{
			name: "escape quote in string field",
			args: args{
				name:   "cpu",
				tags:   map[string]string{},
				fields: map[string]interface{}{"value": `test"`},
				t:      epoch,
			},
			expected: []byte(`cpu value="test\"" 0`),
		},
		{
			name: "escape quote and backslash in string field",
			args: args{
				name:   "cpu",
				tags:   map[string]string{},
				fields: map[string]interface{}{"value": `test\"`},
				t:      epoch,
			},
			expected: []byte(`cpu value="test\\\"" 0`),
		},
		{
			name: "escape multiple backslash in string field",
			args: args{
				name:   "cpu",
				tags:   map[string]string{},
				fields: map[string]interface{}{"value": `test\\`},
				t:      epoch,
			},
			expected: []byte(`cpu value="test\\\\" 0`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 512)
			m, err := New(tt.args.name, tt.args.tags, tt.args.fields, tt.args.t, tt.args.mType...)
			require.NoError(t, err)

			r := NewReader([]telegraf.Metric{m})
			num, err := r.Read(buf)
			if err != io.EOF {
				require.NoError(t, err)
			}
			line := string(buf[:num])
			// This is done so that we can use raw strings in the test spec
			noeol := strings.TrimRight(line, "\n")
			require.Equal(t, string(tt.expected), noeol)
			require.Equal(t, len(tt.expected)+1, num)
		})
	}
}

func TestMetricRoundtrip(t *testing.T) {
	const lp = `nstat,bu=linux,cls=server,dc=cer,env=production,host=hostname,name=netstat,sr=database IpExtInBcastOctets=12570626154i,IpExtInBcastPkts=95541226i,IpExtInCEPkts=0i,IpExtInCsumErrors=0i,IpExtInECT0Pkts=55674i,IpExtInECT1Pkts=0i,IpExtInMcastOctets=5928296i,IpExtInMcastPkts=174365i,IpExtInNoECTPkts=17965863529i,IpExtInNoRoutes=20i,IpExtInOctets=3334866321815i,IpExtInTruncatedPkts=0i,IpExtOutBcastOctets=0i,IpExtOutBcastPkts=0i,IpExtOutMcastOctets=0i,IpExtOutMcastPkts=0i,IpExtOutOctets=31397892391399i,TcpExtArpFilter=0i,TcpExtBusyPollRxPackets=0i,TcpExtDelayedACKLocked=14094i,TcpExtDelayedACKLost=302083i,TcpExtDelayedACKs=55486507i,TcpExtEmbryonicRsts=11879i,TcpExtIPReversePathFilter=0i,TcpExtListenDrops=1736i,TcpExtListenOverflows=0i,TcpExtLockDroppedIcmps=0i,TcpExtOfoPruned=0i,TcpExtOutOfWindowIcmps=8i,TcpExtPAWSActive=0i,TcpExtPAWSEstab=974i,TcpExtPAWSPassive=0i,TcpExtPruneCalled=0i,TcpExtRcvPruned=0i,TcpExtSyncookiesFailed=12593i,TcpExtSyncookiesRecv=0i,TcpExtSyncookiesSent=0i,TcpExtTCPACKSkippedChallenge=0i,TcpExtTCPACKSkippedFinWait2=0i,TcpExtTCPACKSkippedPAWS=806i,TcpExtTCPACKSkippedSeq=519i,TcpExtTCPACKSkippedSynRecv=0i,TcpExtTCPACKSkippedTimeWait=0i,TcpExtTCPAbortFailed=0i,TcpExtTCPAbortOnClose=22i,TcpExtTCPAbortOnData=36593i,TcpExtTCPAbortOnLinger=0i,TcpExtTCPAbortOnMemory=0i,TcpExtTCPAbortOnTimeout=674i,TcpExtTCPAutoCorking=494253233i,TcpExtTCPBacklogDrop=0i,TcpExtTCPChallengeACK=281i,TcpExtTCPDSACKIgnoredNoUndo=93354i,TcpExtTCPDSACKIgnoredOld=336i,TcpExtTCPDSACKOfoRecv=0i,TcpExtTCPDSACKOfoSent=7i,TcpExtTCPDSACKOldSent=302073i,TcpExtTCPDSACKRecv=215884i,TcpExtTCPDSACKUndo=7633i,TcpExtTCPDeferAcceptDrop=0i,TcpExtTCPDirectCopyFromBacklog=0i,TcpExtTCPDirectCopyFromPrequeue=0i,TcpExtTCPFACKReorder=1320i,TcpExtTCPFastOpenActive=0i,TcpExtTCPFastOpenActiveFail=0i,TcpExtTCPFastOpenCookieReqd=0i,TcpExtTCPFastOpenListenOverflow=0i,TcpExtTCPFastOpenPassive=0i,TcpExtTCPFastOpenPassiveFail=0i,TcpExtTCPFastRetrans=350681i,TcpExtTCPForwardRetrans=142168i,TcpExtTCPFromZeroWindowAdv=4317i,TcpExtTCPFullUndo=29502i,TcpExtTCPHPAcks=10267073000i,TcpExtTCPHPHits=5629837098i,TcpExtTCPHPHitsToUser=0i,TcpExtTCPHystartDelayCwnd=285127i,TcpExtTCPHystartDelayDetect=12318i,TcpExtTCPHystartTrainCwnd=69160570i,TcpExtTCPHystartTrainDetect=3315799i,TcpExtTCPLossFailures=109i,TcpExtTCPLossProbeRecovery=110819i,TcpExtTCPLossProbes=233995i,TcpExtTCPLossUndo=5276i,TcpExtTCPLostRetransmit=397i,TcpExtTCPMD5NotFound=0i,TcpExtTCPMD5Unexpected=0i,TcpExtTCPMemoryPressures=0i,TcpExtTCPMinTTLDrop=0i,TcpExtTCPOFODrop=0i,TcpExtTCPOFOMerge=7i,TcpExtTCPOFOQueue=15196i,TcpExtTCPOrigDataSent=29055119435i,TcpExtTCPPartialUndo=21320i,TcpExtTCPPrequeueDropped=0i,TcpExtTCPPrequeued=0i,TcpExtTCPPureAcks=1236441827i,TcpExtTCPRcvCoalesce=225590473i,TcpExtTCPRcvCollapsed=0i,TcpExtTCPRenoFailures=0i,TcpExtTCPRenoRecovery=0i,TcpExtTCPRenoRecoveryFail=0i,TcpExtTCPRenoReorder=0i,TcpExtTCPReqQFullDoCookies=0i,TcpExtTCPReqQFullDrop=0i,TcpExtTCPRetransFail=41i,TcpExtTCPSACKDiscard=0i,TcpExtTCPSACKReneging=0i,TcpExtTCPSACKReorder=4307i,TcpExtTCPSYNChallenge=244i,TcpExtTCPSackFailures=1698i,TcpExtTCPSackMerged=184668i,TcpExtTCPSackRecovery=97369i,TcpExtTCPSackRecoveryFail=381i,TcpExtTCPSackShiftFallback=2697079i,TcpExtTCPSackShifted=760299i,TcpExtTCPSchedulerFailed=0i,TcpExtTCPSlowStartRetrans=9276i,TcpExtTCPSpuriousRTOs=959i,TcpExtTCPSpuriousRtxHostQueues=2973i,TcpExtTCPSynRetrans=200970i,TcpExtTCPTSReorder=15221i,TcpExtTCPTimeWaitOverflow=0i,TcpExtTCPTimeouts=70127i,TcpExtTCPToZeroWindowAdv=4317i,TcpExtTCPWantZeroWindowAdv=2133i,TcpExtTW=24809813i,TcpExtTWKilled=0i,TcpExtTWRecycled=0i 1496460785000000000
nstat,bu=linux,cls=server,dc=cer,env=production,host=hostname,name=snmp,sr=database IcmpInAddrMaskReps=0i,IcmpInAddrMasks=90i,IcmpInCsumErrors=0i,IcmpInDestUnreachs=284401i,IcmpInEchoReps=9i,IcmpInEchos=1761912i,IcmpInErrors=407i,IcmpInMsgs=2047767i,IcmpInParmProbs=0i,IcmpInRedirects=0i,IcmpInSrcQuenchs=0i,IcmpInTimeExcds=46i,IcmpInTimestampReps=0i,IcmpInTimestamps=1309i,IcmpMsgInType0=9i,IcmpMsgInType11=46i,IcmpMsgInType13=1309i,IcmpMsgInType17=90i,IcmpMsgInType3=284401i,IcmpMsgInType8=1761912i,IcmpMsgOutType0=1761912i,IcmpMsgOutType14=1248i,IcmpMsgOutType3=108709i,IcmpMsgOutType8=9i,IcmpOutAddrMaskReps=0i,IcmpOutAddrMasks=0i,IcmpOutDestUnreachs=108709i,IcmpOutEchoReps=1761912i,IcmpOutEchos=9i,IcmpOutErrors=0i,IcmpOutMsgs=1871878i,IcmpOutParmProbs=0i,IcmpOutRedirects=0i,IcmpOutSrcQuenchs=0i,IcmpOutTimeExcds=0i,IcmpOutTimestampReps=1248i,IcmpOutTimestamps=0i,IpDefaultTTL=64i,IpForwDatagrams=0i,IpForwarding=2i,IpFragCreates=0i,IpFragFails=0i,IpFragOKs=0i,IpInAddrErrors=0i,IpInDelivers=17658795773i,IpInDiscards=0i,IpInHdrErrors=0i,IpInReceives=17659269339i,IpInUnknownProtos=0i,IpOutDiscards=236976i,IpOutNoRoutes=1009i,IpOutRequests=23466783734i,IpReasmFails=0i,IpReasmOKs=0i,IpReasmReqds=0i,IpReasmTimeout=0i,TcpActiveOpens=23308977i,TcpAttemptFails=3757543i,TcpCurrEstab=280i,TcpEstabResets=184792i,TcpInCsumErrors=0i,TcpInErrs=232i,TcpInSegs=17536573089i,TcpMaxConn=-1i,TcpOutRsts=4051451i,TcpOutSegs=29836254873i,TcpPassiveOpens=176546974i,TcpRetransSegs=878085i,TcpRtoAlgorithm=1i,TcpRtoMax=120000i,TcpRtoMin=200i,UdpInCsumErrors=0i,UdpInDatagrams=24441661i,UdpInErrors=0i,UdpLiteInCsumErrors=0i,UdpLiteInDatagrams=0i,UdpLiteInErrors=0i,UdpLiteNoPorts=0i,UdpLiteOutDatagrams=0i,UdpLiteRcvbufErrors=0i,UdpLiteSndbufErrors=0i,UdpNoPorts=17660i,UdpOutDatagrams=51807896i,UdpRcvbufErrors=0i,UdpSndbufErrors=236922i 1496460785000000000
`
	metrics, err := Parse([]byte(lp))
	require.NoError(t, err)
	r := NewReader(metrics)
	buf := make([]byte, 128)
	_, err = r.Read(buf)
	require.NoError(t, err)
	metrics, err = Parse(buf)
	require.NoError(t, err)
}
