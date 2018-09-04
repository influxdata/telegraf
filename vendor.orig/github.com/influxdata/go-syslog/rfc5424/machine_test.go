package rfc5424

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func timeParse(layout, value string) *time.Time {
	t, _ := time.Parse(layout, value)
	return &t
}

func getStringAddress(str string) *string {
	return &str
}

func getUint8Address(x uint8) *uint8 {
	return &x
}

func rxpad(str string, lim int) string {
	str = str + strings.Repeat(" ", lim)
	return str[:lim]
}

type testCase struct {
	input        []byte
	valid        bool
	value        *SyslogMessage
	errorString  string
	partialValue *SyslogMessage
}

var testCases = []testCase{
	// Invalid, empty input
	{
		[]byte(""),
		false,
		nil,
		"expecting a priority value within angle brackets [col 0]",
		nil,
	},
	// Invalid, multiple syslog messages on multiple lines
	{
		[]byte(`<1>1 - - - - - -
		<2>1 - - - - - -`),
		false,
		nil,
		"parsing error [col 16]",
		&SyslogMessage{
			priority: getUint8Address(1),
			severity: getUint8Address(1),
			facility: getUint8Address(0),
			version:  1,
		},
	},
	// Invalid, new lines allowed only within message part
	{
		[]byte("<1>1 - \nhostname - - - -"),
		false,
		nil,
		"expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col 7]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - host\x0Aname - - - -"),
		false,
		nil,
		"expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col 11]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - \nan - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 9]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - a\x0An - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 10]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - \npid - -"),
		false,
		nil,
		"expecting a procid (from 1 to max 128 US-ASCII characters) or a nil value [col 11]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - p\x0Aid - -"),
		false,
		nil,
		"expecting a procid (from 1 to max 128 US-ASCII characters) or a nil value [col 12]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - \nmid -"),
		false,
		nil,
		"expecting a msgid (from 1 to max 32 US-ASCII characters) or a nil value [col 13]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - m\x0Aid -"),
		false,
		nil,
		"expecting a msgid (from 1 to max 32 US-ASCII characters) or a nil value [col 14]",
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, malformed pri
	{
		[]byte("(190>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value within angle brackets [col 0]",
		nil,
	},
	// Malformed pri outputs wrong error
	{
		[]byte("<87]123 -"),
		false,
		nil,
		// (note) > machine can only understand that the ] char is not in the reachable states (just as any number would be in this situation), so it gives the error about the priority val submachine, not about the pri submachine (ie., <prival>)
		"expecting a priority value in the range 1-191 or equal to 0 [col 3]",
		nil, // nil since cannot reach version
	},
	// Invalid, missing pri
	{
		[]byte("122 - - - - - -"),
		false,
		nil,
		"expecting a priority value within angle brackets [col 0]",
		nil,
	},
	// Invalid, missing prival
	{
		[]byte("<>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value in the range 1-191 or equal to 0 [col 1]",
		nil,
	},
	// Invalid, prival with too much digits
	{
		[]byte("<19000021>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value in the range 1-191 or equal to 0 [col 4]",
		nil, // no valid partial message since was not able to reach and extract version (which is mandatory for a valid message)
	},
	// Invalid, prival too high
	{
		[]byte("<192>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value in the range 1-191 or equal to 0 [col 3]",
		nil,
	},
	// Invalid, 0 starting prival
	{
		[]byte("<002>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value in the range 1-191 or equal to 0 [col 2]",
		nil,
	},
	// Invalid, non numeric prival
	{
		[]byte("<aaa>122 2018-11-22"),
		false,
		nil,
		"expecting a priority value in the range 1-191 or equal to 0 [col 1]",
		nil,
	},
	// Invalid, missing version
	{
		[]byte("<100> 2018-11-22"),
		false,
		nil,
		"expecting a version value in the range 1-999 [col 5]",
		nil,
	},
	// Invalid, 0 version
	{
		[]byte("<103>0 2018-11-22"),
		false,
		nil,
		"expecting a version value in the range 1-999 [col 5]",
		nil,
	},
	// Invalid, out of range version
	{
		[]byte("<101>1000 2018-11-22"),
		false,
		nil,
		"expecting a version value in the range 1-999 [col 8]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  100,
		},
	},
	// Invalid, truncated after version whitespace
	{
		[]byte("<1>2 "),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 5]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  2,
		},
	},
	// Invalid, truncated after version
	{
		[]byte("<1>1"),
		false,
		nil,
		"parsing error [col 4]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, non numeric (also partially) version
	{
		[]byte("<1>3a"),
		false,
		nil,
		"parsing error [col 4]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  3,
		},
	},
	{
		[]byte("<1>4a "),
		false,
		nil,
		"parsing error [col 4]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  4,
		},
	},
	{
		[]byte("<102>abc 2018-11-22"),
		false,
		nil,
		"expecting a version value in the range 1-999 [col 5]",
		nil,
	},
	// Invalid, letter rather than timestamp
	{
		[]byte("<1>5 A"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 5]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  5,
		},
	},
	// Invalid, timestamp T and Z must be uppercase
	{
		[]byte(`<29>1 2006-01-02t15:04:05Z - - - - -`),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 16]",
		&SyslogMessage{
			facility: getUint8Address(3),
			severity: getUint8Address(5),
			priority: getUint8Address(29),
			version:  1,
		},
	},
	{
		[]byte(`<29>2 2006-01-02T15:04:05z - - - - -`),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 25]",
		&SyslogMessage{
			facility: getUint8Address(3),
			severity: getUint8Address(5),
			priority: getUint8Address(29),
			version:  2,
		},
	},
	// Invalid, wrong year
	{
		[]byte("<101>123 2"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 10]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  123,
		},
	},
	{
		[]byte("<101>124 20"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 11]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  124,
		},
	},
	{
		[]byte("<101>125 201"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 12]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  125,
		},
	},
	{
		[]byte("<101>125 2013"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 13]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  125,
		},
	},
	{
		[]byte("<101>126 2013-"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 14]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  126,
		},
	},
	{
		[]byte("<101>122 201-11-22"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 12]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  122,
		},
	},
	{
		[]byte("<101>189 0-11-22"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 10]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  189,
		},
	},
	// Invalid, wrong month
	{
		[]byte("<101>122 2018-112-22"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 16]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  122,
		},
	},
	// Invalid, wrong day
	{
		[]byte("<101>123 2018-02-32"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 18]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  123,
		},
	},
	// Invalid, wrong hour
	{
		[]byte("<101>124 2018-02-01:25:15Z"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 19]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  124,
		},
	},
	// Invalid, wrong minutes
	{
		[]byte("<101>125 2003-09-29T22:99:16Z"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 23]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  125,
		},
	},
	// Invalid, wrong seconds
	{
		[]byte("<101>126 2003-09-29T22:09:99Z"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 26]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  126,
		},
	},
	// Invalid, wrong sec fraction
	{
		[]byte("<101>127 2003-09-29T22:09:01.000000000009Z"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 35]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  127,
		},
	},
	{
		[]byte("<101>128 2003-09-29T22:09:01.Z"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 29]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  128,
		},
	},
	{
		[]byte("<101>28 2003-09-29T22:09:01."),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 28]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  28,
		},
	},
	// Invalid, wrong time offset
	{
		[]byte("<101>129 2003-09-29T22:09:01A"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 28]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  129,
		},
	},
	{
		[]byte("<101>130 2003-08-24T05:14:15.000003-24:00"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 37]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  130,
		},
	},
	{
		[]byte("<101>131 2003-08-24T05:14:15.000003-60:00"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 36]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  131,
		},
	},
	{
		[]byte("<101>132 2003-08-24T05:14:15.000003-07:61"),
		false,
		nil,
		"expecting a RFC3339MICRO timestamp or a nil value [col 39]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  132,
		},
	},
	{
		[]byte(`<29>1 2006-01-02T15:04:05Z+07:00 - - - - -`),
		false,
		nil,
		"parsing error [col 26]", // after the Z (valid and complete timestamp) it searches for a whitespace
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2006-01-02T15:04:05Z"),
		},
	},
	// Invalid, non existing dates
	{
		[]byte("<101>11 2003-09-31T22:14:15.003Z"),
		false,
		nil,
		"parsing time \"2003-09-31T22:14:15.003Z\": day out of range [col 32]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  11,
		},
	},
	{
		[]byte("<101>12 2003-09-31T22:14:16Z"),
		false,
		nil,
		"parsing time \"2003-09-31T22:14:16Z\": day out of range [col 28]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  12,
		},
	},
	{
		[]byte("<101>12 2018-02-29T22:14:16+01:00"),
		false,
		nil,
		"parsing time \"2018-02-29T22:14:16+01:00\": day out of range [col 33]",
		&SyslogMessage{
			priority: getUint8Address(101),
			facility: getUint8Address(12),
			severity: getUint8Address(5),
			version:  12,
		},
	},
	// Invalid, hostname too long
	{
		[]byte("<1>1 - abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcX - - - -"),
		false,
		nil,
		"expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col 262]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 2003-09-29T22:14:16Z abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcX - - - -"),
		false,
		nil,
		"expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col 281]",
		&SyslogMessage{
			priority:  getUint8Address(1),
			facility:  getUint8Address(0),
			severity:  getUint8Address(1),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2003-09-29T22:14:16Z"),
		},
	},
	// Invalid, appname too long
	{
		[]byte("<1>1 - - abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefX - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 57]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 2003-09-29T22:14:16Z - abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefX - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 76]",
		&SyslogMessage{
			priority:  getUint8Address(1),
			facility:  getUint8Address(0),
			severity:  getUint8Address(1),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2003-09-29T22:14:16Z"),
		},
	},
	{
		[]byte("<1>1 - host abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefX - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 60]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			hostname: getStringAddress("host"),
		},
	},
	{
		[]byte("<1>1 2003-09-29T22:14:16Z host abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefX - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 79]",
		&SyslogMessage{
			priority:  getUint8Address(1),
			facility:  getUint8Address(0),
			severity:  getUint8Address(1),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2003-09-29T22:14:16Z"),
			hostname:  getStringAddress("host"),
		},
	},
	// Invalid, procid too long
	{
		[]byte("<1>1 - - - abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabX - -"),
		false,
		nil,
		"expecting a procid (from 1 to max 128 US-ASCII characters) or a nil value [col 139]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, msgid too long
	{
		[]byte("<1>1 - - - - abcdefghilmnopqrstuvzabcdefghilmX -"),
		false,
		nil,
		"expecting a msgid (from 1 to max 32 US-ASCII characters) or a nil value [col 45]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Not print US-ASCII chars for hostname, appname, procid, and msgid
	{
		[]byte("<1>1 -   - - - -"),
		false,
		nil,
		"expecting an hostname (from 1 to max 255 US-ASCII characters) or a nil value [col 7]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - -   - - -"),
		false,
		nil,
		"expecting an app-name (from 1 to max 48 US-ASCII characters) or a nil value [col 9]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - -   - -"),
		false,
		nil,
		"expecting a procid (from 1 to max 128 US-ASCII characters) or a nil value [col 11]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - -   -"),
		false,
		nil,
		"expecting a msgid (from 1 to max 32 US-ASCII characters) or a nil value [col 13]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, with malformed structured data
	{
		[]byte("<1>1 - - - - - X"),
		false,
		nil,
		"expecting a structured data section containing one or more elements (`[id( key=\"value\")*]+`) or a nil value [col 15]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, with empty structured data
	{
		[]byte("<1>1 - - - - - []"),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 16]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, with structured data id containing space
	{
		[]byte("<1>1 - - - - - [ ]"),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 16]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, with structured data id containing =
	{
		[]byte("<1>1 - - - - - [=]"),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 16]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, with structured data id containing ]
	{
		[]byte("<1>1 - - - - - []]"),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 16]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, with structured data id containing "
	{
		[]byte(`<1>1 - - - - - ["]`),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 16]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, too long structured data id
	{
		[]byte(`<1>1 - - - - - [abcdefghilmnopqrstuvzabcdefghilmX]`),
		false,
		nil,
		"expecting a structured data element id (from 1 to max 32 US-ASCII characters; except `=`, ` `, `]`, and `\"` [col 48]",
		&SyslogMessage{
			priority:       getUint8Address(1),
			facility:       getUint8Address(0),
			severity:       getUint8Address(1),
			version:        1,
			structuredData: nil,
		},
	},
	// Invalid, too long structured data param key
	{
		[]byte(`<1>1 - - - - - [id abcdefghilmnopqrstuvzabcdefghilmX="val"]`),
		false,
		nil,
		"expecting a structured data parameter (`key=\"value\"`, both part from 1 to max 32 US-ASCII characters; key cannot contain `=`, ` `, `]`, and `\"`, while value cannot contain `]`, backslash, and `\"` unless escaped) [col 51]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			structuredData: &map[string]map[string]string{
				"id": map[string]string{},
			},
		},
	},
	// Valid, minimal
	{
		[]byte("<1>1 - - - - - -"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
		},
		"",
		nil,
	},
	{
		[]byte("<0>1 - - - - - -"),
		true,
		&SyslogMessage{
			priority: getUint8Address(0),
			facility: getUint8Address(0),
			severity: getUint8Address(0),
			version:  1,
		},
		"",
		nil,
	},
	// Valid, average message
	{
		[]byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice - - [origin x-service="someservice"][meta sequenceId="14125553"] 127.0.0.1 - - 1456029177 "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2016-02-21T04:32:57+00:00"),
			hostname:  getStringAddress("web1"),
			appname:   getStringAddress("someservice"),
			structuredData: &map[string]map[string]string{
				"origin": map[string]string{
					"x-service": "someservice",
				},
				"meta": map[string]string{
					"sequenceId": "14125553",
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1456029177 "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
		},
		"",
		nil,
	},
	// Valid, hostname, appname, procid, msgid can contain dashes
	{
		[]byte("<1>100 - host-name - - - -"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  100,
			hostname: getStringAddress("host-name"),
		},
		"",
		nil,
	},
	{
		[]byte("<1>101 - host-name app-name - - -"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  101,
			hostname: getStringAddress("host-name"),
			appname:  getStringAddress("app-name"),
		},
		"",
		nil,
	},
	{
		[]byte("<1>102 - host-name app-name proc-id - -"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  102,
			hostname: getStringAddress("host-name"),
			appname:  getStringAddress("app-name"),
			procID:   getStringAddress("proc-id"),
		},
		"",
		nil,
	},
	{
		[]byte("<1>103 - host-name app-name proc-id msg-id -"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  103,
			hostname: getStringAddress("host-name"),
			appname:  getStringAddress("app-name"),
			procID:   getStringAddress("proc-id"),
			msgID:    getStringAddress("msg-id"),
		},
		"",
		nil,
	},
	// Valid, w/0 structured data and w/o message, with other fields all max length
	{
		[]byte("<191>999 2018-12-31T23:59:59.999999-23:59 abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab abcdefghilmnopqrstuvzabcdefghilm -"),
		true,
		&SyslogMessage{
			priority:  getUint8Address(191),
			facility:  getUint8Address(23),
			severity:  getUint8Address(7),
			version:   999,
			timestamp: timeParse(RFC3339MICRO, "2018-12-31T23:59:59.999999-23:59"),
			hostname:  getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"),
			appname:   getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"),
			procID:    getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"),
			msgID:     getStringAddress("abcdefghilmnopqrstuvzabcdefghilm"),
		},
		"",
		nil,
	},
	// Valid, all fields max length, with structured data and message
	{
		[]byte(`<191>999 2018-12-31T23:59:59.999999-23:59 abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab abcdefghilmnopqrstuvzabcdefghilm [an@id key1="val1" key2="val2"][another@id key1="val1"] Some message "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(191),
			facility:  getUint8Address(23),
			severity:  getUint8Address(7),
			version:   999,
			timestamp: timeParse(RFC3339MICRO, "2018-12-31T23:59:59.999999-23:59"),
			hostname:  getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"),
			appname:   getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"),
			procID:    getStringAddress("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"),
			msgID:     getStringAddress("abcdefghilmnopqrstuvzabcdefghilm"),
			structuredData: &map[string]map[string]string{
				"an@id": map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				"another@id": map[string]string{
					"key1": "val1",
				},
			},
			message: getStringAddress(`Some message "GET"`),
		},
		"",
		nil,
	},
	// Valid, w/o structure data, w/0 procid
	{
		[]byte("<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8"),
		true,
		&SyslogMessage{
			facility:       getUint8Address(4),
			severity:       getUint8Address(2),
			priority:       getUint8Address(34),
			version:        1,
			timestamp:      timeParse(RFC3339MICRO, "2003-10-11T22:14:15.003Z"),
			hostname:       getStringAddress("mymachine.example.com"),
			appname:        getStringAddress("su"),
			procID:         nil,
			msgID:          getStringAddress("ID47"),
			structuredData: nil,
			message:        getStringAddress("BOM'su root' failed for lonvick on /dev/pts/8"),
		},
		"",
		nil,
	},
	// Valid, w/o structure data, w/o timestamp
	{
		[]byte("<187>222 - mymachine.example.com su - ID47 - 'su root' failed for lonvick on /dev/pts/8"),
		true,
		&SyslogMessage{
			facility:       getUint8Address(23),
			severity:       getUint8Address(3),
			priority:       getUint8Address(187),
			version:        222,
			timestamp:      nil,
			hostname:       getStringAddress("mymachine.example.com"),
			appname:        getStringAddress("su"),
			procID:         nil,
			msgID:          getStringAddress("ID47"),
			structuredData: nil,
			message:        getStringAddress("'su root' failed for lonvick on /dev/pts/8"),
		},
		"",
		nil,
	},
	// Valid, w/o structure data, w/o msgid
	{
		[]byte("<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% Time to make the do-nuts."),
		true,
		&SyslogMessage{
			facility:       getUint8Address(20),
			severity:       getUint8Address(5),
			priority:       getUint8Address(165),
			version:        1,
			timestamp:      timeParse(RFC3339MICRO, "2003-08-24T05:14:15.000003-07:00"),
			hostname:       getStringAddress("192.0.2.1"),
			appname:        getStringAddress("myproc"),
			procID:         getStringAddress("8710"),
			msgID:          nil,
			structuredData: nil,
			message:        getStringAddress("%% Time to make the do-nuts."),
		},
		"",
		nil,
	},
	// Valid, w/o structure data, w/o hostname, w/o appname, w/o procid, w/o msgid, w/o msg
	{
		[]byte("<165>2 2003-08-24T05:14:15.000003-07:00 - - - - -"),
		true,
		&SyslogMessage{
			facility:       getUint8Address(20),
			severity:       getUint8Address(5),
			priority:       getUint8Address(165),
			version:        2,
			timestamp:      timeParse(RFC3339MICRO, "2003-08-24T05:14:15.000003-07:00"),
			hostname:       nil,
			appname:        nil,
			procID:         nil,
			msgID:          nil,
			structuredData: nil,
			message:        nil,
		},
		"",
		nil,
	},
	// Valid, w/o structure data, w/o hostname, w/o appname, w/o procid, w/o msgid, empty msg
	{
		[]byte("<165>222 2003-08-24T05:14:15.000003-07:00 - - - - - "),
		true,
		&SyslogMessage{
			facility:       getUint8Address(20),
			severity:       getUint8Address(5),
			priority:       getUint8Address(165),
			version:        222,
			timestamp:      timeParse(RFC3339MICRO, "2003-08-24T05:14:15.000003-07:00"),
			hostname:       nil,
			appname:        nil,
			procID:         nil,
			msgID:          nil,
			structuredData: nil,
			message:        nil,
		},
		"",
		nil,
	},
	// Valid, with structured data is, w/o structured data params
	{
		[]byte("<78>1 2016-01-15T00:04:01+00:00 host1 CROND 10391 - [sdid] some_message"),
		true,
		&SyslogMessage{
			facility:  getUint8Address(9),
			severity:  getUint8Address(6),
			priority:  getUint8Address(78),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T00:04:01+00:00"),
			hostname:  getStringAddress("host1"),
			appname:   getStringAddress("CROND"),
			procID:    getStringAddress("10391"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"sdid": map[string]string{},
			},
			message: getStringAddress("some_message"),
		},
		"",
		nil,
	},
	// Valid, with structured data id, with structured data params
	{
		[]byte(`<78>1 2016-01-15T00:04:01+00:00 host1 CROND 10391 - [sdid x="⌘"] some_message`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(9),
			severity:  getUint8Address(6),
			priority:  getUint8Address(78),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T00:04:01+00:00"),
			hostname:  getStringAddress("host1"),
			appname:   getStringAddress("CROND"),
			procID:    getStringAddress("10391"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"sdid": map[string]string{
					"x": "⌘",
				},
			},
			message: getStringAddress("some_message"),
		},
		"",
		nil,
	},
	// Valid, with structured data is, with structured data params
	{
		[]byte(`<78>2 2016-01-15T00:04:01+00:00 host1 CROND 10391 - [sdid x="hey \\u2318 hey"] some_message`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(9),
			severity:  getUint8Address(6),
			priority:  getUint8Address(78),
			version:   2,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T00:04:01+00:00"),
			hostname:  getStringAddress("host1"),
			appname:   getStringAddress("CROND"),
			procID:    getStringAddress("10391"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"sdid": map[string]string{
					"x": `hey \u2318 hey`,
				},
			},
			message: getStringAddress("some_message"),
		},
		"",
		nil,
	},
	// Valid, with (escaped) backslash within structured data param value
	{
		[]byte(`<29>50 2016-01-15T01:00:43Z hn S - - [meta es="\\valid"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   50,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"es": `\valid`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>52 2016-01-15T01:00:43Z hn S - - [meta one="\\one" two="\\two"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   52,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"one": `\one`,
					"two": `\two`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>53 2016-01-15T01:00:43Z hn S - - [meta one="\\one"][other two="\\two" double="\\a\\b"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   53,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"one": `\one`,
				},
				"other": map[string]string{
					"two":    `\two`,
					"double": `\a\b`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>51 2016-01-15T01:00:43Z hn S - - [meta es="\\double\\slash"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   51,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"es": `\double\slash`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>54 2016-01-15T01:00:43Z hn S - - [meta es="in \\middle of the string"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   54,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"es": `in \middle of the string`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>55 2016-01-15T01:00:43Z hn S - - [meta es="at the \\end"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   55,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"es": `at the \end`,
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	// Valid, with control characters within structured data param value
	{
		[]byte("<29>50 2016-01-15T01:00:43Z hn S - - [meta es=\"\t5Ὂg̀9!℃ᾭGa b\"] 127.0.0.1 - - 1452819643 \"GET\""),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   50,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"es": "\t5Ὂg̀9!℃ᾭGa b",
				},
			},
			message: getStringAddress(`127.0.0.1 - - 1452819643 "GET"`),
		},
		"",
		nil,
	},
	// Valid, with utf8 within structured data param value
	{
		[]byte(`<29>50 2016-01-15T01:00:43Z hn S - - [meta gr="κόσμε" es="ñ"][beta pr="₡"] 𐌼 "GET"`),
		true,
		&SyslogMessage{
			priority:  getUint8Address(29),
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			version:   50,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"gr": "κόσμε",
					"es": "ñ",
				},
				"beta": map[string]string{
					"pr": "₡",
				},
			},
			message: getStringAddress(`𐌼 "GET"`),
		},
		"",
		nil,
	},
	// Valid, with structured data, w/o msg
	{
		[]byte("<165>3 2003-10-11T22:14:15.003Z example.com evnts - ID27 [exampleSDID@32473 iut=\"3\" eventSource=\"Application\" eventID=\"1011\"][examplePriority@32473 class=\"high\"]"),
		true,
		&SyslogMessage{
			facility:  getUint8Address(20),
			severity:  getUint8Address(5),
			priority:  getUint8Address(165),
			version:   3,
			timestamp: timeParse(RFC3339MICRO, "2003-10-11T22:14:15.003Z"),
			hostname:  getStringAddress("example.com"),
			appname:   getStringAddress("evnts"),
			procID:    nil,
			msgID:     getStringAddress("ID27"),
			structuredData: &map[string]map[string]string{
				"exampleSDID@32473": map[string]string{
					"iut":         "3",
					"eventSource": "Application",
					"eventID":     "1011",
				},
				"examplePriority@32473": map[string]string{
					"class": "high",
				},
			},
			message: nil,
		},
		"",
		nil,
	},
	// Invalid, with duplicated structured data id
	{
		[]byte("<165>3 2003-10-11T22:14:15.003Z example.com evnts - ID27 [id1][id1]"),
		false,
		nil,
		"duplicate structured data element id [col 66]",
		&SyslogMessage{
			priority:  getUint8Address(165),
			facility:  getUint8Address(20),
			severity:  getUint8Address(5),
			version:   3,
			timestamp: timeParse(RFC3339MICRO, "2003-10-11T22:14:15.003Z"),
			hostname:  getStringAddress("example.com"),
			appname:   getStringAddress("evnts"),
			msgID:     getStringAddress("ID27"),
			structuredData: &map[string]map[string]string{
				"id1": map[string]string{},
			},
		},
	},
	// Invalid, with duplicated structured data id
	{
		[]byte("<165>3 2003-10-11T22:14:15.003Z example.com evnts - ID27 [dupe e=\"1\"][id1][dupe class=\"l\"]"),
		false,
		nil,
		"duplicate structured data element id [col 79]",
		&SyslogMessage{
			priority:  getUint8Address(165),
			facility:  getUint8Address(20),
			severity:  getUint8Address(5),
			version:   3,
			timestamp: timeParse(RFC3339MICRO, "2003-10-11T22:14:15.003Z"),
			hostname:  getStringAddress("example.com"),
			appname:   getStringAddress("evnts"),
			msgID:     getStringAddress("ID27"),
			structuredData: &map[string]map[string]string{
				"id1": map[string]string{},
				"dupe": map[string]string{
					"e": "1",
				},
			},
		},
	},
	// Valid, with structured data w/o msg
	{
		[]byte(`<165>4 2003-10-11T22:14:15.003Z mymachine.it e - 1 [ex@32473 iut="3" eventSource="A"] An application event log entry...`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(20),
			severity:  getUint8Address(5),
			priority:  getUint8Address(165),
			version:   4,
			timestamp: timeParse(RFC3339MICRO, "2003-10-11T22:14:15.003Z"),
			hostname:  getStringAddress("mymachine.it"),
			appname:   getStringAddress("e"),
			procID:    nil,
			msgID:     getStringAddress("1"),
			structuredData: &map[string]map[string]string{
				"ex@32473": map[string]string{
					"iut":         "3",
					"eventSource": "A",
				},
			},
			message: getStringAddress("An application event log entry..."),
		},
		"",
		nil,
	},
	// Valid, with double quotes in the message
	{
		[]byte(`<29>1 2016-01-15T01:00:43Z some-host-name SEKRETPROGRAM prg - [origin x-service="svcname"][meta sequenceId="1"] 127.0.0.1 - - 1452819643 "GET"`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   1,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("some-host-name"),
			appname:   getStringAddress("SEKRETPROGRAM"),
			procID:    getStringAddress("prg"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"origin": map[string]string{
					"x-service": "svcname",
				},
				"meta": map[string]string{
					"sequenceId": "1",
				},
			},
			message: getStringAddress("127.0.0.1 - - 1452819643 \"GET\""),
		},
		"",
		nil,
	},
	// Valid, with empty structured data param value
	{
		[]byte(`<1>1 - - - - - [id pk=""]`),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			structuredData: &map[string]map[string]string{
				"id": map[string]string{
					"pk": "",
				},
			},
		},
		"",
		nil,
	},
	// Valid, with escaped character within param value
	{
		[]byte(`<29>2 2016-01-15T01:00:43Z some-host-name SEKRETPROGRAM prg - [meta escape="\]"] some "mex"`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   2,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("some-host-name"),
			appname:   getStringAddress("SEKRETPROGRAM"),
			procID:    getStringAddress("prg"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"escape": "]",
				},
			},
			message: getStringAddress(`some "mex"`),
		},
		"",
		nil,
	},
	{
		[]byte(`<29>2 2016-01-15T01:00:43Z some-host-name SEKRETPROGRAM prg - [meta escape="\\"]`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   2,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("some-host-name"),
			appname:   getStringAddress("SEKRETPROGRAM"),
			procID:    getStringAddress("prg"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"escape": `\`,
				},
			},
		},
		"",
		nil,
	},
	{
		[]byte(`<29>2 2016-01-15T01:00:43Z some-host-name SEKRETPROGRAM prg - [meta escape="\""]`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   2,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("some-host-name"),
			appname:   getStringAddress("SEKRETPROGRAM"),
			procID:    getStringAddress("prg"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"escape": `"`,
				},
			},
		},
		"",
		nil,
	},
	{
		[]byte(`<29>2 2016-01-15T01:00:43Z some-host-name SEKRETPROGRAM prg - [meta escape="\]\"\\\\\]\""]`),
		true,
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   2,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("some-host-name"),
			appname:   getStringAddress("SEKRETPROGRAM"),
			procID:    getStringAddress("prg"),
			msgID:     nil,
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{
					"escape": `]"\\]"`,
				},
			},
		},
		"",
		nil,
	},
	// Invalid, param value can not contain closing square bracket - ie., ]
	{
		[]byte(`<29>3 2016-01-15T01:00:43Z hn S - - [meta escape="]"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 50]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   3,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>5 2016-01-15T01:00:43Z hn S - - [meta escape="]q"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 50]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   5,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>4 2016-01-15T01:00:43Z hn S - - [meta escape="p]"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 51]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   4,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	// Invalid, param value can not contain doublequote char - ie., ""
	{
		[]byte(`<29>4 2017-01-15T01:00:43Z hn S - - [meta escape="""] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting a structured data parameter (`key=\"value\"`, both part from 1 to max 32 US-ASCII characters; key cannot contain `=`, ` `, `]`, and `\"`, while value cannot contain `]`, backslash, and `\"` unless escaped) [col 51]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   4,
			timestamp: timeParse(RFC3339MICRO, "2017-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>6 2016-01-15T01:00:43Z hn S - - [meta escape="a""] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting a structured data parameter (`key=\"value\"`, both part from 1 to max 32 US-ASCII characters; key cannot contain `=`, ` `, `]`, and `\"`, while value cannot contain `]`, backslash, and `\"` unless escaped) [col 52]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   6,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>4 2018-01-15T01:00:43Z hn S - - [meta escape=""b"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting a structured data parameter (`key=\"value\"`, both part from 1 to max 32 US-ASCII characters; key cannot contain `=`, ` `, `]`, and `\"`, while value cannot contain `]`, backslash, and `\"` unless escaped) [col 51]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   4,
			timestamp: timeParse(RFC3339MICRO, "2018-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	// Invalid, param value can not contain backslash - ie., \
	{
		[]byte(`<29>5 2019-01-15T01:00:43Z hn S - - [meta escape="\"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 52]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   5,
			timestamp: timeParse(RFC3339MICRO, "2019-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>7 2019-01-15T01:00:43Z hn S - - [meta escape="a\"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 53]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   7,
			timestamp: timeParse(RFC3339MICRO, "2019-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	{
		[]byte(`<29>8 2016-01-15T01:00:43Z hn S - - [meta escape="\n"] 127.0.0.1 - - 1452819643 "GET"`),
		false,
		nil,
		"expecting chars `]`, `\"`, and `\\` to be escaped within param value [col 51]",
		&SyslogMessage{
			facility:  getUint8Address(3),
			severity:  getUint8Address(5),
			priority:  getUint8Address(29),
			version:   8,
			timestamp: timeParse(RFC3339MICRO, "2016-01-15T01:00:43Z"),
			hostname:  getStringAddress("hn"),
			appname:   getStringAddress("S"),
			structuredData: &map[string]map[string]string{
				"meta": map[string]string{},
			},
		},
	},
	// Valid, message starting with byte order mark (BOM, \uFEFF)
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\ufeff"),
		},
		"",
		nil,
	},
	// Valid, greek
	{
		[]byte("<1>1 - - - - - - κόσμε"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("κόσμε"),
		},
		"",
		nil,
	},
	// Valid, 2 octet sequence
	{
		[]byte("<1>1 - - - - - - "),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress(""),
		},
		"",
		nil,
	},
	// Valid, spanish (2 octet sequence)
	{
		[]byte("<1>1 - - - - - - \xc3\xb1"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("ñ"),
		},
		"",
		nil,
	},
	// Valid, colon currency sign (3 octet sequence)
	{
		[]byte("<1>1 - - - - - - \xe2\x82\xa1"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("₡"),
		},
		"",
		nil,
	},
	// Valid, gothic letter (4 octet sequence)
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF \xf0\x90\x8c\xbc"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\ufeff 𐌼"),
		},
		"",
		nil,
	},
	// Valid, 5 octet sequence
	{
		[]byte("<1>1 - - - - - - \xC8\x80\x30\x30\x30"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("Ȁ000"),
		},
		"",
		nil,
	},
	// Valid, 6 octet sequence
	{
		[]byte("<1>1 - - - - - - \xE4\x80\x80\x30\x30\x30"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("䀀000"),
		},
		"",
		nil,
	},
	// Valid, UTF-8 boundary conditions
	{
		[]byte("<1>1 - - - - - - \xC4\x90\x30\x30\x30"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("Đ000"),
		},
		"",
		nil,
	},
	{
		[]byte("<1>1 - - - - - - \x0D\x37\x46\x46"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\r7FF"),
		},
		"",
		nil,
	},
	// Valid, Tamil poetry of Subramaniya Bharathiyar
	{
		[]byte("<1>1 - - - - - - யாமறிந்த மொழிகளிலே தமிழ்மொழி போல் இனிதாவது எங்கும் காணோம், பாமரராய் விலங்குகளாய், உலகனைத்தும் இகழ்ச்சிசொலப் பான்மை கெட்டு, நாமமது தமிழரெனக் கொண்டு இங்கு வாழ்ந்திடுதல் நன்றோ? சொல்லீர்! தேமதுரத் தமிழோசை உலகமெலாம் பரவும்வகை செய்தல் வேண்டும்."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("யாமறிந்த மொழிகளிலே தமிழ்மொழி போல் இனிதாவது எங்கும் காணோம், பாமரராய் விலங்குகளாய், உலகனைத்தும் இகழ்ச்சிசொலப் பான்மை கெட்டு, நாமமது தமிழரெனக் கொண்டு இங்கு வாழ்ந்திடுதல் நன்றோ? சொல்லீர்! தேமதுரத் தமிழோசை உலகமெலாம் பரவும்வகை செய்தல் வேண்டும்."),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Milanese)
	{
		[]byte("<1>1 - - - - - - Sôn bôn de magnà el véder, el me fa minga mal."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("Sôn bôn de magnà el véder, el me fa minga mal."),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Romano)
	{
		[]byte("<1>1 - - - - - - Me posso magna' er vetro, e nun me fa male."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("Me posso magna' er vetro, e nun me fa male."),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Braille)
	{
		[]byte("<1>1 - - - - - - ⠊⠀⠉⠁⠝⠀⠑⠁⠞⠀⠛⠇⠁⠎⠎⠀⠁⠝⠙⠀⠊⠞⠀⠙⠕⠑⠎⠝⠞⠀⠓⠥⠗⠞⠀⠍⠑"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("⠊⠀⠉⠁⠝⠀⠑⠁⠞⠀⠛⠇⠁⠎⠎⠀⠁⠝⠙⠀⠊⠞⠀⠙⠕⠑⠎⠝⠞⠀⠓⠥⠗⠞⠀⠍⠑"),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Sanskrit)
	{
		[]byte("<1>1 - - - - - - काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("काचं शक्नोम्यत्तुम् । नोपहिनस्ति माम् ॥"),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Urdu)
	{
		[]byte("<1>1 - - - - - - میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("میں کانچ کھا سکتا ہوں اور مجھے تکلیف نہیں ہوتی ۔"),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Yiddish)
	{
		[]byte("<1>1 - - - - - - איך קען עסן גלאָז און עס טוט מיר נישט װײ."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("איך קען עסן גלאָז און עס טוט מיר נישט װײ."),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Polish)
	{
		[]byte("<1>1 - - - - - - Mogę jeść szkło, i mi nie szkodzi."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("Mogę jeść szkło, i mi nie szkodzi."),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Japanese)
	{
		[]byte("<1>1 - - - - - - 私はガラスを食べられます。それは私を傷つけません。"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("私はガラスを食べられます。それは私を傷つけません。"),
		},
		"",
		nil,
	},
	// Valid, I Can Eat Glass (Arabic)
	{
		[]byte("<1>1 - - - - - - أنا قادر على أكل الزجاج و هذا لا يؤلمني."),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("أنا قادر على أكل الزجاج و هذا لا يؤلمني."),
		},
		"",
		nil,
	},
	// Valid, russian alphabet
	{
		[]byte("<1>1 - - - - - - абвгдеёжзийклмнопрстуфхцчшщъыьэюя"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("абвгдеёжзийклмнопрстуфхцчшщъыьэюя"),
		},
		"",
		nil,
	},
	// Valid, armenian letters
	{
		[]byte("<1>1 - - - - - - ԰ԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖ՗՘ՙ՚՛՜՝՞՟աբգդեզէըթիլխծկհձղճմյնշոչպջռսվտրցւփքօֆևֈ։֊֋֌֍֎֏"),
		true,
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\u0530ԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖ\u0557\u0558ՙ՚՛՜՝՞՟աբգդեզէըթիլխծկհձղճմյնշոչպջռսվտրցւփքօֆև\u0588։֊\u058b\u058c֍֎֏"),
		},
		"",
		nil,
	},
	// Valid, new line within message
	{
		[]byte("<1>1 - - - - - - x\x0Ay"),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  1,
			message:  getStringAddress("x\ny"),
		},
		"",
		nil,
	},
	{
		[]byte(`<1>2 - - - - - - x
y`),
		true,
		&SyslogMessage{
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			priority: getUint8Address(1),
			version:  2,
			message:  getStringAddress("x\ny"),
		},
		"",
		nil,
	},
	// Invalid, out of range code within message
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xC1"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 20]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF"),
		},
	},
	{
		[]byte("<1>2 - - - - - - \xC1"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  2,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xc3\x28"), // invalid 2 octet sequence
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 21]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xc3"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xc3\x28"), // invalid 2 octet sequence
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xc3"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xa0\xa1"), // invalid sequence identifier
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 20]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xa0\xa1"), // invalid sequence identifier
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xe2\x28\xa1"), // invalid 3 octet sequence (2nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 21]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xe2"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xe2\x28\xa1"), // invalid 3 octet sequence (2nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xe2"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xe2\x82\x28"), // invalid 3 octet sequence (3nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 22]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xe2\x82"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xe2\x82\x28"), // invalid 3 octet sequence (3nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 19]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xe2\x82"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xf0\x28\x8c\xbc"), // invalid 4 octet sequence (2nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 21]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xf0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf0\x28\x8c\xbc"), // invalid 4 octet sequence (2nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xf0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xf0\x90\x28\xbc"), // invalid 4 octet sequence (3nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 22]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xf0\x90"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf0\x90\x28\xbc"), // invalid 4 octet sequence (3nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 19]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xf0\x90"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xEF\xBB\xBF\xf0\x28\x8c\x28"), // invalid 4 octet sequence (4nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 21]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xEF\xBB\xBF\xf0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf0\x28\x8c\x28"), // invalid 4 octet sequence (4nd octet)
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xf0"),
		},
	},
	// Invalid, impossible bytes
	{
		[]byte("<1>1 - - - - - - \xfe\xfe\xff\xff"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xfe"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xff"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, overlong sequences
	{
		[]byte("<1>1 - - - - - - \xfc\x80\x80\x80\x80\xaf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf8\x80\x80\x80\xaf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf0\x80\x80\xaf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xf0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xe0\x80\xaf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xe0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xc0\xaf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	// Invalid, maximum overlong sequences
	{
		[]byte("<1>1 - - - - - - \xfc\x83\xbf\xbf\xbf\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf8\x87\xbf\xbf\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
		},
	},
	{
		[]byte("<1>1 - - - - - - \xf0\x8f\xbf\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xf0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xe0\x9f\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xe0"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xc1\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 17]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  nil,
		},
	},
	// Invalid, illegal code positions, single utf-16 surrogates
	{
		[]byte("<1>1 - - - - - - \xed\xa0\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xa0\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xad\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xae\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xaf\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xb0\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xbe\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	{
		[]byte("<1>1 - - - - - - \xed\xbf\xbf"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	// Invalid, illegal code positions, paired utf-16 surrogates
	{
		[]byte("<1>1 - - - - - - \xed\xa0\x80\xed\xb0\x80"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 18]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("\xed"),
		},
	},
	// Invalid, out of range code within message after valid string
	{
		[]byte("<1>1 - - - - - - valid\xEF\xBB\xBF\xC1"),
		false,
		nil,
		"expecting a free-form optional message in UTF-8 (starting with or without BOM) [col 25]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  1,
			message:  getStringAddress("valid\ufeff"),
		},
	},
	// Invalid, missing whitespace after nil timestamp
	{
		[]byte("<1>10 -- - - - -"),
		false,
		nil,
		"parsing error [col 7]",
		&SyslogMessage{
			priority: getUint8Address(1),
			facility: getUint8Address(0),
			severity: getUint8Address(1),
			version:  10,
		},
	},

	// (fixme) > evaluate non characters for UTF-8 security concerns, eg. \xef\xbf\xbe
}

func generateIncompleteTimestampTestCases() []testCase {
	incompleteTimestamp := []byte("2003-11-02T23:12:46.012345")
	prefix := []byte("<1>1 ")
	mex := &SyslogMessage{
		priority: getUint8Address(1),
		severity: getUint8Address(1),
		facility: getUint8Address(0),
		version:  1,
	}
	tCases := make([]testCase, 0, len(incompleteTimestamp))
	prev := make([]byte, 0, len(incompleteTimestamp))
	for i, d := range incompleteTimestamp {
		prev = append(prev, d)
		tc := testCase{
			input:        append(prefix, prev...),
			valid:        false,
			value:        nil,
			errorString:  fmt.Sprintf("expecting a RFC3339MICRO timestamp or a nil value [col %d]", len(prefix)+i+1),
			partialValue: mex,
		}
		tCases = append(tCases, tc)
	}
	return tCases
}

func generateUntilMaxLengthStringTestCases(max []byte, pos int) []testCase {
	if pos < 0 || pos > 3 {
		panic("position not available")
	}
	templ := "<%d>%d - - - - - -"
	where := 9 + (pos * 2)
	templ = templ[:where] + "%s" + templ[where+1:]

	tCases := []testCase{}
	prev := ""
	for _, c := range max {
		prev += string(c)
		randp := random(0, 9)
		randv := random(1, 9)

		input := []byte(fmt.Sprintf(templ, randp, randv, prev))

		mex := &SyslogMessage{
			priority: getUint8Address(uint8(randp)),
			severity: getUint8Address(uint8(randp % 8)),
			facility: getUint8Address(uint8(randp / 8)),
			version:  uint16(randv),
		}
		switch pos {
		case 0:
			mex.hostname = getStringAddress(string(prev))
		case 1:
			mex.appname = getStringAddress(string(prev))
		case 2:
			mex.procID = getStringAddress(string(prev))
		case 3:
			mex.msgID = getStringAddress(string(prev))
		}

		t := testCase{
			input,
			true,
			mex,
			"",
			nil,
		}

		tCases = append(tCases, t)
	}
	return tCases
}

func init() {
	rand.Seed(time.Now().Unix())

	testCases = append(testCases, generateIncompleteTimestampTestCases()...)

	hostnameMaxStr := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc")
	testCases = append(testCases, generateUntilMaxLengthStringTestCases(hostnameMaxStr, 0)...)

	appnameMaxStr := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef")
	testCases = append(testCases, generateUntilMaxLengthStringTestCases(appnameMaxStr, 1)...)

	procidMaxStr := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab")
	testCases = append(testCases, generateUntilMaxLengthStringTestCases(procidMaxStr, 2)...)

	msgidMaxStr := []byte("abcdefghilmnopqrstuvzabcdefghilm")
	testCases = append(testCases, generateUntilMaxLengthStringTestCases(msgidMaxStr, 3)...)
}

func TestMachineParse(t *testing.T) {
	for _, tc := range testCases {
		tc := tc
		t.Run(rxpad(string(tc.input), 50), func(t *testing.T) {
			t.Parallel()

			bestEffort := true
			fsm := NewMachine()
			message, merr := fsm.Parse(tc.input, nil)
			partial, perr := fsm.Parse(tc.input, &bestEffort)

			if !tc.valid {
				assert.Nil(t, message)
				assert.Error(t, merr)
				assert.EqualError(t, merr, tc.errorString)

				assert.Equal(t, tc.partialValue, partial)
				assert.EqualError(t, perr, tc.errorString)
			}
			if tc.valid {
				assert.Nil(t, merr)
				assert.NotEmpty(t, message)
				assert.Equal(t, message, partial)
				assert.Equal(t, merr, perr)
			}

			assert.Equal(t, tc.value, message)
		})
	}
}
