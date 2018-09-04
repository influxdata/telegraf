package rfc5425

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func output(out interface{}) {
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true
	spew.Dump(out)
}

func Example() {
	r := strings.NewReader("48 <1>1 2003-10-11T22:14:15.003Z host.local - - - -25 <3>1 - host.local - - - -38 <2>1 - host.local su - - - κόσμε")
	output(NewParser(r, WithBestEffort()).Parse())
	// Output:
	// ([]rfc5425.Result) (len=3) {
	//  (rfc5425.Result) {
	//   Message: (*rfc5424.SyslogMessage)({
	//    priority: (*uint8)(1),
	//    facility: (*uint8)(0),
	//    severity: (*uint8)(1),
	//    version: (uint16) 1,
	//    timestamp: (*time.Time)(2003-10-11 22:14:15.003 +0000 UTC),
	//    hostname: (*string)((len=10) "host.local"),
	//    appname: (*string)(<nil>),
	//    procID: (*string)(<nil>),
	//    msgID: (*string)(<nil>),
	//    structuredData: (*map[string]map[string]string)(<nil>),
	//    message: (*string)(<nil>)
	//   }),
	//   MessageError: (error) <nil>,
	//   Error: (error) <nil>
	//  },
	//  (rfc5425.Result) {
	//   Message: (*rfc5424.SyslogMessage)({
	//    priority: (*uint8)(3),
	//    facility: (*uint8)(0),
	//    severity: (*uint8)(3),
	//    version: (uint16) 1,
	//    timestamp: (*time.Time)(<nil>),
	//    hostname: (*string)((len=10) "host.local"),
	//    appname: (*string)(<nil>),
	//    procID: (*string)(<nil>),
	//    msgID: (*string)(<nil>),
	//    structuredData: (*map[string]map[string]string)(<nil>),
	//    message: (*string)(<nil>)
	//   }),
	//   MessageError: (error) <nil>,
	//   Error: (error) <nil>
	//  },
	//  (rfc5425.Result) {
	//   Message: (*rfc5424.SyslogMessage)({
	//    priority: (*uint8)(2),
	//    facility: (*uint8)(0),
	//    severity: (*uint8)(2),
	//    version: (uint16) 1,
	//    timestamp: (*time.Time)(<nil>),
	//    hostname: (*string)((len=10) "host.local"),
	//    appname: (*string)((len=2) "su"),
	//    procID: (*string)(<nil>),
	//    msgID: (*string)(<nil>),
	//    structuredData: (*map[string]map[string]string)(<nil>),
	//    message: (*string)((len=11) "κόσμε")
	//   }),
	//   MessageError: (error) <nil>,
	//   Error: (error) <nil>
	//  }
	// }
}

func Example_handler() {
	// Defining a custom handler to send the parsing results into a channel
	EmittingParsing := func(p *Parser) chan Result {
		c := make(chan Result)

		toChannel := func(r *Result) {
			c <- *r
		}

		go func() {
			defer close(c)
			p.ParseExecuting(toChannel)
		}()

		return c
	}

	// Use it
	r := strings.NewReader("16 <1>1 - - - - - -17 <2>12 A B C D E -16 <1>1")
	results := EmittingParsing(NewParser(r, WithBestEffort()))

	for r := range results {
		output(r)
	}
	// Output:
	// (rfc5425.Result) {
	//  Message: (*rfc5424.SyslogMessage)({
	//   priority: (*uint8)(1),
	//   facility: (*uint8)(0),
	//   severity: (*uint8)(1),
	//   version: (uint16) 1,
	//   timestamp: (*time.Time)(<nil>),
	//   hostname: (*string)(<nil>),
	//   appname: (*string)(<nil>),
	//   procID: (*string)(<nil>),
	//   msgID: (*string)(<nil>),
	//   structuredData: (*map[string]map[string]string)(<nil>),
	//   message: (*string)(<nil>)
	//  }),
	//  MessageError: (error) <nil>,
	//  Error: (error) <nil>
	// }
	// (rfc5425.Result) {
	//  Message: (*rfc5424.SyslogMessage)({
	//   priority: (*uint8)(2),
	//   facility: (*uint8)(0),
	//   severity: (*uint8)(2),
	//   version: (uint16) 12,
	//   timestamp: (*time.Time)(<nil>),
	//   hostname: (*string)(<nil>),
	//   appname: (*string)(<nil>),
	//   procID: (*string)(<nil>),
	//   msgID: (*string)(<nil>),
	//   structuredData: (*map[string]map[string]string)(<nil>),
	//   message: (*string)(<nil>)
	//  }),
	//  MessageError: (*errors.errorString)(expecting a RFC3339MICRO timestamp or a nil value [col 6]),
	//  Error: (error) <nil>
	// }
	// (rfc5425.Result) {
	//  Message: (*rfc5424.SyslogMessage)({
	//   priority: (*uint8)(1),
	//   facility: (*uint8)(0),
	//   severity: (*uint8)(1),
	//   version: (uint16) 1,
	//   timestamp: (*time.Time)(<nil>),
	//   hostname: (*string)(<nil>),
	//   appname: (*string)(<nil>),
	//   procID: (*string)(<nil>),
	//   msgID: (*string)(<nil>),
	//   structuredData: (*map[string]map[string]string)(<nil>),
	//   message: (*string)(<nil>)
	//  }),
	//  MessageError: (*errors.errorString)(parsing error [col 4]),
	//  Error: (*errors.errorString)(found EOF after "<1>1", expecting a SYSLOGMSG containing 16 octets)
	// }
}
