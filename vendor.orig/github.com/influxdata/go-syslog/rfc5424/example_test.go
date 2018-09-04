package rfc5424

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

func output(out interface{}) {
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true
	spew.Dump(out)
}

func Example() {
	i := []byte(`<165>4 2018-10-11T22:14:15.003Z mymach.it e - 1 [ex@32473 iut="3"] An application event log entry...`)
	p := NewParser()
	m, _ := p.Parse(i, nil)
	output(m)
	// Output:
	// (*rfc5424.SyslogMessage)({
	//  priority: (*uint8)(165),
	//  facility: (*uint8)(20),
	//  severity: (*uint8)(5),
	//  version: (uint16) 4,
	//  timestamp: (*time.Time)(2018-10-11 22:14:15.003 +0000 UTC),
	//  hostname: (*string)((len=9) "mymach.it"),
	//  appname: (*string)((len=1) "e"),
	//  procID: (*string)(<nil>),
	//  msgID: (*string)((len=1) "1"),
	//  structuredData: (*map[string]map[string]string)((len=1) {
	//   (string) (len=8) "ex@32473": (map[string]string) (len=1) {
	//    (string) (len=3) "iut": (string) (len=1) "3"
	//   }
	//  }),
	//  message: (*string)((len=33) "An application event log entry...")
	// })
}

func Example_besteffort() {
	bestEffortOn := true
	i := []byte(`<1>1 A - - - - - -`)
	p := NewParser()
	m, e := p.Parse(i, &bestEffortOn)
	output(m)
	fmt.Println(e)
	// Output:
	// (*rfc5424.SyslogMessage)({
	//  priority: (*uint8)(1),
	//  facility: (*uint8)(0),
	//  severity: (*uint8)(1),
	//  version: (uint16) 1,
	//  timestamp: (*time.Time)(<nil>),
	//  hostname: (*string)(<nil>),
	//  appname: (*string)(<nil>),
	//  procID: (*string)(<nil>),
	//  msgID: (*string)(<nil>),
	//  structuredData: (*map[string]map[string]string)(<nil>),
	//  message: (*string)(<nil>)
	// })
	// expecting a RFC3339MICRO timestamp or a nil value [col 5]
}

func Example_builder() {
	msg := &SyslogMessage{}
	msg.SetTimestamp("not a RFC3339MICRO timestamp")
	fmt.Println("Valid?", msg.Valid())
	msg.SetPriority(191)
	msg.SetVersion(1)
	fmt.Println("Valid?", msg.Valid())
	output(msg)
	str, _ := msg.String()
	fmt.Println(str)
	// Output:
	// Valid? false
	// Valid? true
	// (*rfc5424.SyslogMessage)({
	//  priority: (*uint8)(191),
	//  facility: (*uint8)(23),
	//  severity: (*uint8)(7),
	//  version: (uint16) 1,
	//  timestamp: (*time.Time)(<nil>),
	//  hostname: (*string)(<nil>),
	//  appname: (*string)(<nil>),
	//  procID: (*string)(<nil>),
	//  msgID: (*string)(<nil>),
	//  structuredData: (*map[string]map[string]string)(<nil>),
	//  message: (*string)(<nil>)
	// })
	// <191>1 - - - - - -
}
