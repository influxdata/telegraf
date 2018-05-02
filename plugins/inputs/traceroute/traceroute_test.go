package traceroute

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TR_Line_Test struct {
	Line         string
	Entries      []string
	NumberOfHops int
	HopInfo      []TracerouteHopInfo
}

type TR_Column_Test struct {
	Text      string
	CarryOver [3]string
	Fqdn      string
	Ip        string
	Asn       string
	Rtt       float64
}

var (
	LinuxTracerouteOutput = `
traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets
 1  165.227.32.254 (165.227.32.254)  1.206 ms 165.227.32.253 (165.227.32.253)  1.188 ms 165.227.32.254 (165.227.32.254)  1.143 msg
 2  138.197.249.78 (138.197.249.78)  0.985 ms 138.197.249.86 (138.197.249.86)  0.939 ms 138.197.249.90 (138.197.249.90)  1.181 ms
 3  72.14.219.10 (72.14.219.10)  0.818 ms 162.243.190.33 (162.243.190.33)  0.952 ms  0.941 ms
 4  108.170.250.225 (108.170.250.225)  0.825 ms  0.970 ms 108.170.250.241 (108.170.250.241)  1.007 ms
 5  108.170.226.217 (108.170.226.217)  0.995 ms 108.170.226.219 (108.170.226.219)  1.033 ms 108.170.226.217 (108.170.226.217)  1.003 ms
 6  dfw06s38-in-f14.1e100.net (172.217.0.238)  1.187 ms  0.722 ms  0.545 ms
`
	LinuxTracerouteHeader = `traceroute to google.com (172.217.0.238), 30 hops max, 60 byte packets`
	LinuxTracerouteFqdn   = `google.com`
	LinuxTracerouteIp     = `172.217.0.238`
)

func MockHostTracerouter(timeout float64, args ...string) (string, error) {
	return LinuxTracerouteOutput, nil
}

func TestFindNumberOfHops(t *testing.T) {
	numHops := findNumberOfHops(LinuxTracerouteOutput)
	assert.Equal(t, 6, numHops, "6 hops made by packet")
}

var SampleTracerouteLine = `12  54.239.110.174 (54.239.110.174)  22.609 ms 54.239.110.130 (54.239.110.130)  26.629 ms 54.239.110.183 (54.239.110.183)  34.258 ms`

func TestGetHopNumber(t *testing.T) {
	hopNum, err := findHopNumber(SampleTracerouteLine)
	assert.NoError(t, err)
	assert.Equal(t, 12, hopNum, "Traceroute line is the 12th hop")
}

var TR_Line_Test_Suite = []TR_Line_Test{
	Normal_TR_Line_Test,
	SomeVoid_TR_Line_Test,
	NoHost_TR_Line_Test,
	AllVoid_TR_Line_Test,
	NoHost_ASN_1_TR_Line_Test,
	LowerCaseAs_TR_Line_Test,
}
var (
	NormalTracerouteLine      = ` 6  yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms  1.244 ms  0.417 ms`
	NormalTracerouteEntries   = []string{"yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms", "1.244 ms", "0.417 ms"}
	NormalTracerouteHopNumber = 6
	NormalTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 0,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			Asn:       "",
			RTT:       1.480,
		},
		TracerouteHopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 1,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			Asn:       "",
			RTT:       1.244,
		},
		TracerouteHopInfo{
			HopNumber: NormalTracerouteHopNumber,
			ColumnNum: 2,
			Fqdn:      "yyz10s03-in-f3.1e100.net",
			Ip:        "172.217.0.227",
			Asn:       "",
			RTT:       0.417,
		},
	}
	Normal_TR_Line_Test = TR_Line_Test{
		Line:         NormalTracerouteLine,
		Entries:      NormalTracerouteEntries,
		NumberOfHops: NormalTracerouteHopNumber,
		HopInfo:      NormalTracerouteHopInfo,
	}
)
var (
	SomeVoidTracerouteLine      = `14  54.239.110.152 (54.239.110.152)  27.198 ms * 54.239.110.247 (54.239.110.247)  37.625 ms`
	SomeVoidTracerouteEntries   = []string{"54.239.110.152 (54.239.110.152)  27.198 ms", "*", "54.239.110.247 (54.239.110.247)  37.625 ms"}
	SomeVoidTracerouteHopNumber = 14
	SomeVoidTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: SomeVoidTracerouteHopNumber,
			ColumnNum: 0,
			Fqdn:      "54.239.110.152",
			Ip:        "54.239.110.152",
			Asn:       "",
			RTT:       27.198,
		},
		TracerouteHopInfo{
			HopNumber: SomeVoidTracerouteHopNumber,
			ColumnNum: 2,
			Fqdn:      "54.239.110.247",
			Ip:        "54.239.110.247",
			Asn:       "",
			RTT:       37.625,
		},
	}
	SomeVoid_TR_Line_Test = TR_Line_Test{
		Line:         SomeVoidTracerouteLine,
		Entries:      SomeVoidTracerouteEntries,
		NumberOfHops: SomeVoidTracerouteHopNumber,
		HopInfo:      SomeVoidTracerouteHopInfo,
	}
)

var (
	NoHostTracerouteLine      = `10  129.250.2.81  186.767 ms`
	NoHostTracerouteEntries   = []string{"129.250.2.81  186.767 ms"}
	NoHostTracerouteHopNumber = 10
	NoHostTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: NoHostTracerouteHopNumber,
			ColumnNum: 0,
			Fqdn:      "129.250.2.81",
			Ip:        "129.250.2.81",
			Asn:       "",
			RTT:       186.767,
		},
	}
	NoHost_TR_Line_Test = TR_Line_Test{
		Line:         NoHostTracerouteLine,
		Entries:      NoHostTracerouteEntries,
		NumberOfHops: NoHostTracerouteHopNumber,
		HopInfo:      NoHostTracerouteHopInfo,
	}
)

var (
	NoHost_ASN_1_TR_Line      = `15  77.238.190.3 [AS34010]  155.664 ms 77.238.190.2 [AS34010]  155.539 ms 77.238.190.5 [AS34010]  157.304 ms`
	NoHost_ASN_1_TR_Entries   = []string{"77.238.190.3 [AS34010]  155.664 ms", "77.238.190.2 [AS34010]  155.539 ms", "77.238.190.5 [AS34010]  157.304 ms"}
	NoHost_ASN_1_TR_HopNumber = 15
	NoHost_ASN_1_TR_HopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_1_TR_HopNumber,
			ColumnNum: 0,
			Fqdn:      "77.238.190.3",
			Ip:        "77.238.190.3",
			Asn:       "AS34010",
			RTT:       155.664,
		},
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_1_TR_HopNumber,
			ColumnNum: 1,
			Fqdn:      "77.238.190.2",
			Ip:        "77.238.190.2",
			Asn:       "AS34010",
			RTT:       155.539,
		},
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_1_TR_HopNumber,
			ColumnNum: 2,
			Fqdn:      "77.238.190.5",
			Ip:        "77.238.190.5",
			Asn:       "AS34010",
			RTT:       157.304,
		},
	}
	NoHost_ASN_1_TR_Line_Test = TR_Line_Test{
		Line:         NoHost_ASN_1_TR_Line,
		Entries:      NoHost_ASN_1_TR_Entries,
		NumberOfHops: NoHost_ASN_1_TR_HopNumber,
		HopInfo:      NoHost_ASN_1_TR_HopInfo,
	}
	NoHost_ASN_2_NumberOfHops = 14
	NoHost_ASN_2_TR_Line_Test = TR_Line_Test{
		Line: "14  49.255.198.125 [*]  188.903 ms 101.0.127.233 [AS38880/AS38220/AS55803]  187.293 ms  182.836 ms",
		Entries: []string{
			"49.255.198.125 [*]  188.903 ms",
			"101.0.127.233 [AS38880/AS38220/AS55803]  187.293 m",
			"182.836 ms",
		},
		NumberOfHops: 14,
		HopInfo: []TracerouteHopInfo{
			TracerouteHopInfo{
				HopNumber: NoHost_ASN_2_NumberOfHops,
				ColumnNum: 0,
				Fqdn:      "49.255.198.125",
				Ip:        "49.255.198.125",
				Asn:       "",
				RTT:       188.903,
			},
			TracerouteHopInfo{
				HopNumber: NoHost_ASN_2_NumberOfHops,
				ColumnNum: 1,
				Fqdn:      "101.0.127.233",
				Ip:        "101.0.127.233",
				Asn:       "AS38880/AS38220/AS55803",
				RTT:       187.293,
			},
			TracerouteHopInfo{
				HopNumber: NoHost_ASN_2_NumberOfHops,
				ColumnNum: 2,
				Fqdn:      "101.0.127.233",
				Ip:        "101.0.127.233",
				Asn:       "AS38880/AS38220/AS55803",
				RTT:       187.293,
			},
		},
	}
)

var (
	NoHost_ASN_TR_2_Line      = `17  101.0.127.49 [AS38880/AS38220/AS55803]  183.849 ms 101.0.126.74 [AS55803/AS38880/AS38220]  184.038 ms  180.053 ms`
	NoHost_ASN_TR_2_Entries   = []string{"101.0.127.49 [AS38880/AS38220/AS55803]  183.849 ms", "101.0.126.74 [AS55803/AS38880/AS38220]  184.038 ms", "180.053 ms"}
	NoHost_ASN_TR_2_HopNumber = 17
	NoHost_ASN_TR_2_HopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_TR_2_HopNumber,
			ColumnNum: 0,
			Fqdn:      "101.0.127.49",
			Ip:        "101.0.127.49",
			Asn:       "AS38880/AS38220/AS55803",
			RTT:       183.849,
		},
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_TR_2_HopNumber,
			ColumnNum: 1,
			Fqdn:      "101.0.126.74",
			Ip:        "101.0.126.74",
			Asn:       "AS55803/AS38880/AS38220",
			RTT:       184.038,
		},
		TracerouteHopInfo{
			HopNumber: NoHost_ASN_TR_2_HopNumber,
			ColumnNum: 2,
			Fqdn:      "101.0.126.74",
			Ip:        "101.0.126.74",
			Asn:       "AS55803/AS38880/AS38220",
			RTT:       180.053,
		},
	}
)

var (
	AllVoidTracerouteLine      = `5  * * *`
	AllVoidTracerouteEntries   = []string{"*", "*", "*"}
	AllVoidTracerouteHopNumber = 5
	AllVoid_TR_Line_Test       = TR_Line_Test{
		Line:         AllVoidTracerouteLine,
		Entries:      AllVoidTracerouteEntries,
		NumberOfHops: AllVoidTracerouteHopNumber,
		HopInfo:      []TracerouteHopInfo{},
	}
)

var (
	LowerCaseAsTracerouteLine      = `6  206.248.155.168 [as13768]  86.202 ms  68.356 ms  68.281 ms`
	LowerCaseAsTracerouteEntries   = []string{"206.248.155.168 [as13768]  86.202 ms", "68.356 ms", "68.281 ms"}
	LowerCaseAsTracerouteHopNumber = 6
	LowerCaseAsTracerouteHopInfo   = []TracerouteHopInfo{
		TracerouteHopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 0,
			Fqdn:      "206.248.155.168",
			Ip:        "206.248.155.168",
			Asn:       "as13768",
			RTT:       86.202,
		},
		TracerouteHopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 1,
			Fqdn:      "206.248.155.168",
			Ip:        "206.248.155.168",
			Asn:       "as13768",
			RTT:       68.356,
		},
		TracerouteHopInfo{
			HopNumber: LowerCaseAsTracerouteHopNumber,
			ColumnNum: 2,
			Fqdn:      "206.248.155.168",
			Ip:        "206.248.155.168",
			Asn:       "as13768",
			RTT:       68.281,
		},
	}
	LowerCaseAs_TR_Line_Test = TR_Line_Test{
		Line:         LowerCaseAsTracerouteLine,
		Entries:      LowerCaseAsTracerouteEntries,
		NumberOfHops: LowerCaseAsTracerouteHopNumber,
		HopInfo:      LowerCaseAsTracerouteHopInfo,
	}
)

func TestFindColumnEntries(t *testing.T) {
	var entries []string
	for _, tr_line_test := range TR_Line_Test_Suite {
		entries = findColumnEntries(tr_line_test.Line)
		assert.Equal(t, len(tr_line_test.Entries), len(entries), "# entries")
		assert.True(t, reflect.DeepEqual(tr_line_test.Entries, entries), "Expected: %s, Actual: %s", tr_line_test.Entries, entries)
	}

	entries = findColumnEntries(NormalTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(NormalTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, NormalTracerouteEntries)

	entries = findColumnEntries(SomeVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(SomeVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, SomeVoidTracerouteEntries)

	entries = findColumnEntries(NoHostTracerouteLine)
	assert.Equal(t, 1, len(entries), "1 entry")
	assert.True(t, reflect.DeepEqual(NoHostTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, NoHostTracerouteEntries)

	entries = findColumnEntries(AllVoidTracerouteLine)
	assert.Equal(t, 3, len(entries), "3 entries")
	assert.True(t, reflect.DeepEqual(AllVoidTracerouteEntries, entries), "Expected: %s, Actual: %s", entries, AllVoidTracerouteEntries)

}

var TR_Column_Test_Suite = []TR_Column_Test{
	IpFqdn_TR_Column_Test,
	HttpFqdn_TR_Column_Test,
	CarryOver_TR_Column_Test,
	NoHost_TR_Column_Test,
	ASN_1_TR_Column_Test,
}
var (
	IpFqdnColumnEntry     = `12  54.239.110.174 (54.239.110.174)  22.609 ms`
	IpFqdn_TR_Column_Test = TR_Column_Test{
		Text:      IpFqdnColumnEntry,
		CarryOver: [3]string{"", "", ""},
		Fqdn:      "54.239.110.174",
		Ip:        "54.239.110.174",
		Asn:       "",
		Rtt:       22.609,
	}
	HttpFqdnColumnEntry     = `yyz10s03-in-f3.1e100.net (172.217.0.227)  1.480 ms`
	HttpFqdn_TR_Column_Test = TR_Column_Test{
		Text:      HttpFqdnColumnEntry,
		CarryOver: [3]string{"some", "thing", "inconsequential"},
		Fqdn:      "yyz10s03-in-f3.1e100.net",
		Ip:        "172.217.0.227",
		Asn:       "",
		Rtt:       1.480,
	}
	CarryOverColumnEntry     = `0.417 ms`
	CarryOver_Params         = [3]string{"wildmadagascar.org", "75.101.140.9", "AS16509"}
	CarryOver_TR_Column_Test = TR_Column_Test{
		Text:      CarryOverColumnEntry,
		CarryOver: CarryOver_Params,
		Fqdn:      CarryOver_Params[0],
		Ip:        CarryOver_Params[1],
		Asn:       CarryOver_Params[2],
		Rtt:       0.417,
	}
	NoHostColumnEntry     = `3  192.168.1.1  2.854 ms`
	NoHost_TR_Column_Test = TR_Column_Test{
		Text:      NoHostColumnEntry,
		CarryOver: [3]string{"Ja", "Pa", "Dog"},
		Fqdn:      "192.168.1.1",
		Ip:        "",
		Asn:       "",
		Rtt:       2.854,
	}
	ASN_1_TR_Column_Test = TR_Column_Test{
		Text:      "66.163.66.70 (66.163.66.70) [AS6327]  65.254 ms",
		CarryOver: [3]string{"", "", ""},
		Fqdn:      "66.163.66.70",
		Ip:        "66.163.66.70",
		Asn:       "AS6327",
		Rtt:       65.254,
	}
)

func TestProcessTracerouteColumnEntry(t *testing.T) {
	var fqdn, ip, asn string
	var rtt float32
	var err error
	acceptableDelta := 0.0005

	for _, tr_column_test := range TR_Column_Test_Suite {
		fqdn, ip, asn, rtt, err = processTracerouteColumnEntry(tr_column_test.Text, 1, tr_column_test.CarryOver[0], tr_column_test.CarryOver[1], tr_column_test.CarryOver[2])
		assert.NoError(t, err)
		assert.Equal(t, tr_column_test.Fqdn, fqdn, "fqdn")
		assert.Equal(t, tr_column_test.Ip, ip, "ip")
		assert.Equal(t, tr_column_test.Asn, asn, "asn")
		assert.InDelta(t, tr_column_test.Rtt, rtt, acceptableDelta, "rtt")
	}
}

func TestProcessTracerouteHopLine(t *testing.T) {
	var (
		hopInfo []TracerouteHopInfo
		err     error
	)
	for _, tr_line_test := range TR_Line_Test_Suite {
		hopInfo, err = processTracerouteHopLine(tr_line_test.Line)
		assert.NoError(t, err)
		expectedHopInfo := tr_line_test.HopInfo
		assert.True(t, reflect.DeepEqual(expectedHopInfo, hopInfo), "Expected: %s Actual: %s", expectedHopInfo, hopInfo)
	}

}

func TestProcessTracerouteHeaderLine(t *testing.T) {
	fqdn, ip := processTracerouteHeaderLine(LinuxTracerouteHeader)
	assert.Equal(t, LinuxTracerouteFqdn, fqdn, "fqdn")
	assert.Equal(t, LinuxTracerouteIp, ip, "ip")
}
