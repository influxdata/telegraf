package rfc5425

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/go-syslog/rfc5424"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	descr    string
	input    string
	results  []Result
	pResults []Result
}

var testCases []testCase

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // Number of letter indices fitting in 63 bits
)

func getRandomString(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func getStringAddress(str string) *string {
	return &str
}

func getUint8Address(x uint8) *uint8 {
	return &x
}

func getTimestampError(col int) error {
	return fmt.Errorf("expecting a RFC3339MICRO timestamp or a nil value [col %d]", col)
}

func getParsingError(col int) error {
	return fmt.Errorf("parsing error [col %d]", col)
}

func getTestCases() []testCase {
	maxPriority := uint8(191)
	maxVersion := uint16(999)
	maxTimestamp := "2018-12-31T23:59:59.999999-23:59"
	maxHostname := "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"
	maxAppname := "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"
	maxProcID := "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"
	maxMsgID := "abcdefghilmnopqrstuvzabcdefghilm"
	message7681 := getRandomString(7681)

	return []testCase{
		{
			"empty",
			"",
			[]Result{
				Result{Error: fmt.Errorf("found %s, expecting a %s", EOF, MSGLEN)},
			},
			[]Result{
				Result{Error: fmt.Errorf("found %s, expecting a %s", EOF, MSGLEN)},
			},
		},
		{
			"1st ok/2nd mf", // mf means malformed syslog message
			"16 <1>1 - - - - - -17 <2>12 A B C D E -",
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					MessageError: getTimestampError(6),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Message:      (&rfc5424.SyslogMessage{}).SetPriority(2).SetVersion(12),
					MessageError: getTimestampError(6),
				},
			},
		},
		{
			"1st ok/2nd ko", // ko means wrong token
			"16 <1>1 - - - - - -xaaa",
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", Token{ILLEGAL, []byte("x")}, MSGLEN),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", Token{ILLEGAL, []byte("x")}, MSGLEN),
				},
			},
		},
		{
			"1st ml/2nd ko",
			"16 <1>1 A B C D E -xaaa",
			// results w/o best effort
			[]Result{
				Result{
					MessageError: getTimestampError(5),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message:      (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
					MessageError: getTimestampError(5),
				},
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", Token{ILLEGAL, []byte("x")}, MSGLEN),
				},
			},
		},
		{
			"1st ok//utf8",
			"23 <1>1 - - - - - - hellø", // msglen MUST be the octet count
			//results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(1).
						SetVersion(1).
						SetMessage("hellø"),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(1).
						SetVersion(1).
						SetMessage("hellø"),
				},
			},
		},
		{
			"1st ko//incomplete SYSLOGMSG",
			"16 <1>1",
			// results w/o best effort
			[]Result{
				Result{
					Error: fmt.Errorf(`found %s after "%s", expecting a %s containing %d octets`, EOF, "<1>1", SYSLOGMSG, 16),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message:      (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
					MessageError: getParsingError(4),
					Error:        fmt.Errorf(`found %s after "%s", expecting a %s containing %d octets`, EOF, "<1>1", SYSLOGMSG, 16),
				},
			},
		},
		{
			"1st ko//missing WS found ILLEGAL",
			"16<1>1",
			// results w/o best effort
			[]Result{
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", Token{ILLEGAL, []byte("<")}, WS),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", Token{ILLEGAL, []byte("<")}, WS),
				},
			},
		},
		{
			"1st ko//missing WS found EOF",
			"1",
			// results w/o best effort
			[]Result{
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", EOF, WS),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", EOF, WS),
				},
			},
		},
		{
			"1st ok/2nd ok/3rd ok",
			"48 <1>1 2003-10-11T22:14:15.003Z host.local - - - -25 <3>1 - host.local - - - -38 <2>1 - host.local su - - - κόσμε",
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(1).
						SetVersion(1).
						SetTimestamp("2003-10-11T22:14:15.003Z").
						SetHostname("host.local"),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(3).
						SetVersion(1).
						SetHostname("host.local"),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(2).
						SetVersion(1).
						SetHostname("host.local").
						SetAppname("su").
						SetMessage("κόσμε"),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(1).
						SetVersion(1).
						SetTimestamp("2003-10-11T22:14:15.003Z").
						SetHostname("host.local"),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(3).
						SetVersion(1).
						SetHostname("host.local"),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(2).
						SetVersion(1).
						SetHostname("host.local").
						SetAppname("su").
						SetMessage("κόσμε"),
				},
			},
		},
		{
			"1st ok/2nd mf/3rd ok", // mf means malformed syslog message
			"16 <1>1 - - - - - -17 <2>12 A B C D E -16 <1>1 - - - - - -",
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					MessageError: getTimestampError(6),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Message:      (&rfc5424.SyslogMessage{}).SetPriority(2).SetVersion(12),
					MessageError: getTimestampError(6),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
			},
		},
		{
			"1st ok//max",
			fmt.Sprintf("8192 <%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681),
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
		},
		{
			"1st ok/2nd ok//max/max",
			fmt.Sprintf("8192 <%d>%d %s %s %s %s %s - %s8192 <%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681, maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681),
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
		},
		{
			"1st ok/2nd ok/3rd ok//max/no/max",
			fmt.Sprintf("8192 <%d>%d %s %s %s %s %s - %s16 <1>1 - - - - - -8192 <%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681, maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681),
			// results w/o best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(1),
				},
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
				},
			},
		},
		{
			"1st ml//maxlen gt 8192", // maxlength greather than the buffer size
			fmt.Sprintf("8193 <%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681),
			// results w/o best effort
			[]Result{
				Result{
					Error: fmt.Errorf("found %s after \"%s\", expecting a %s containing %d octets", EOF, fmt.Sprintf("<%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681), SYSLOGMSG, 8193),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message: (&rfc5424.SyslogMessage{}).
						SetPriority(maxPriority).
						SetVersion(maxVersion).
						SetTimestamp(maxTimestamp).
						SetHostname(maxHostname).
						SetAppname(maxAppname).
						SetProcID(maxProcID).
						SetMsgID(maxMsgID).
						SetMessage(message7681),
					Error: fmt.Errorf("found %s after \"%s\", expecting a %s containing %d octets", EOF, fmt.Sprintf("<%d>%d %s %s %s %s %s - %s", maxPriority, maxVersion, maxTimestamp, maxHostname, maxAppname, maxProcID, maxMsgID, message7681), SYSLOGMSG, 8193),
				},
			},
		},
		{
			"1st uf/2nd ok//incomplete SYSLOGMSG/notdetectable",
			"16 <1>217 <11>1 - - - - - -",
			// results w/o best effort
			[]Result{
				Result{
					MessageError: fmt.Errorf(`expecting a RFC3339MICRO timestamp or a nil value [col 7]`),
				},
			},
			// results with best effort
			[]Result{
				Result{
					Message:      (&rfc5424.SyslogMessage{}).SetPriority(1).SetVersion(217),
					MessageError: getTimestampError(7),
				},
				Result{
					Error: fmt.Errorf("found %s, expecting a %s", WS, MSGLEN),
				},
			},
		},
	}
}

func init() {
	testCases = getTestCases()
}

func TestParse(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.descr, func(t *testing.T) {
			t.Parallel()

			res1 := NewParser(strings.NewReader(tc.input)).Parse()
			res2 := NewParser(strings.NewReader(tc.input), WithBestEffort()).Parse()

			assert.Equal(t, tc.results, res1)
			assert.Equal(t, tc.pResults, res2)
		})
	}
}
