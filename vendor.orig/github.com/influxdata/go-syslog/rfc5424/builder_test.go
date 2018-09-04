package rfc5424

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetTimestamp(t *testing.T) {
	m := &SyslogMessage{}

	assert.Equal(t, time.Date(2003, 10, 11, 22, 14, 15, 0, time.UTC), *m.SetTimestamp("2003-10-11T22:14:15Z").Timestamp())
	assert.Equal(t, time.Date(2003, 10, 11, 22, 14, 15, 3000, time.UTC), *m.SetTimestamp("2003-10-11T22:14:15.000003Z").Timestamp())

	// (note) > timestamp is invalid but it accepts until valid - ie., Z char
	// (note) > this dependes on the builder internal parser which does not have a final state, nor we check for any error or final state
	// (todo) > decide wheter to be more strict or not
	assert.Equal(t, time.Date(2003, 10, 11, 22, 14, 15, 3000, time.UTC), *m.SetTimestamp("2003-10-11T22:14:15.000003Z+02:00").Timestamp())
}

func TestFacilityAndSeverity(t *testing.T) {
	m := &SyslogMessage{}

	assert.Nil(t, m.Facility())
	assert.Nil(t, m.FacilityMessage())
	assert.Nil(t, m.FacilityLevel())
	assert.Nil(t, m.Severity())
	assert.Nil(t, m.SeverityMessage())
	assert.Nil(t, m.SeverityLevel())
	assert.Nil(t, m.SeverityShortLevel())

	m.SetPriority(1)

	assert.Equal(t, uint8(0), *m.Facility())
	assert.Equal(t, "kernel messages", *m.FacilityMessage())
	assert.Equal(t, "kern", *m.FacilityLevel())
	assert.Equal(t, uint8(1), *m.Severity())
	assert.Equal(t, "action must be taken immediately", *m.SeverityMessage())
	assert.Equal(t, "alert", *m.SeverityLevel())
	assert.Equal(t, "alert", *m.SeverityShortLevel())

	m.SetPriority(120)

	assert.Equal(t, uint8(15), *m.Facility())
	assert.Equal(t, "clock daemon (note 2)", *m.FacilityMessage())
	assert.Equal(t, "cron", *m.FacilityLevel())
	assert.Equal(t, uint8(0), *m.Severity())
	assert.Equal(t, "system is unusable", *m.SeverityMessage())
	assert.Equal(t, "emergency", *m.SeverityLevel())
	assert.Equal(t, "emerg", *m.SeverityShortLevel())

	m.SetPriority(99)

	assert.Equal(t, uint8(12), *m.Facility())
	assert.Equal(t, "NTP subsystem", *m.FacilityMessage())
	assert.Equal(t, "NTP subsystem", *m.FacilityLevel()) // MUST fallback to message
	assert.Equal(t, uint8(3), *m.Severity())
	assert.Equal(t, "error conditions", *m.SeverityMessage())
	assert.Equal(t, "error", *m.SeverityLevel())
	assert.Equal(t, "err", *m.SeverityShortLevel())
}

func TestSetNilTimestamp(t *testing.T) {
	m := &SyslogMessage{}
	assert.Nil(t, m.SetTimestamp("-").Timestamp())
}

func TestSetIncompleteTimestamp(t *testing.T) {
	m := &SyslogMessage{}
	date := []byte("2003-11-02T23:12:46.012345")
	prev := make([]byte, 0, len(date))
	for _, d := range date {
		prev = append(prev, d)
		assert.Nil(t, m.SetTimestamp(string(prev)).Timestamp())
	}
}

func TestSetSyntacticallyCompleteButIncorrectTimestamp(t *testing.T) {
	m := &SyslogMessage{}
	assert.Nil(t, m.SetTimestamp("2003-42-42T22:14:15Z").Timestamp())
}

func TestSetImpossibleButSyntacticallyCorrectTimestamp(t *testing.T) {
	m := &SyslogMessage{}
	assert.Nil(t, m.SetTimestamp("2003-09-31T22:14:15Z").Timestamp())
}

func TestSetTooLongHostname(t *testing.T) {
	m := &SyslogMessage{}
	m.SetHostname("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcX")
	assert.Nil(t, m.Hostname())
}

func TestSetNilOrEmptyHostname(t *testing.T) {
	m := &SyslogMessage{}
	assert.Nil(t, m.SetHostname("-").Hostname())
	assert.Nil(t, m.SetHostname("").Hostname())
}

func TestSetHostname(t *testing.T) {
	m := &SyslogMessage{}

	maxlen := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc")

	prev := make([]byte, 0, len(maxlen))
	for _, input := range maxlen {
		prev = append(prev, input)
		str := string(prev)
		assert.Equal(t, str, *m.SetHostname(str).Hostname())
	}
}

func TestSetAppname(t *testing.T) {
	m := &SyslogMessage{}

	maxlen := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef")

	prev := make([]byte, 0, len(maxlen))
	for _, input := range maxlen {
		prev = append(prev, input)
		str := string(prev)
		assert.Equal(t, str, *m.SetAppname(str).Appname())
	}
}

func TestSetProcID(t *testing.T) {
	m := &SyslogMessage{}

	maxlen := []byte("abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab")

	prev := make([]byte, 0, len(maxlen))
	for _, input := range maxlen {
		prev = append(prev, input)
		str := string(prev)
		assert.Equal(t, str, *m.SetProcID(str).ProcID())
	}
}

func TestSetMsgID(t *testing.T) {
	m := &SyslogMessage{}

	maxlen := []byte("abcdefghilmnopqrstuvzabcdefghilm")

	prev := make([]byte, 0, len(maxlen))
	for _, input := range maxlen {
		prev = append(prev, input)
		str := string(prev)
		assert.Equal(t, str, *m.SetMsgID(str).MsgID())
	}
}

func TestSetSyntacticallyWrongHostnameAppnameProcIDMsgID(t *testing.T) {
	m := &SyslogMessage{}
	assert.Nil(t, m.SetHostname("white space not possible").Hostname())
	assert.Nil(t, m.SetHostname(string([]byte{0x0})).Hostname())
	assert.Nil(t, m.SetAppname("white space not possible").Appname())
	assert.Nil(t, m.SetAppname(string([]byte{0x0})).Appname())
	assert.Nil(t, m.SetProcID("white space not possible").ProcID())
	assert.Nil(t, m.SetProcID(string([]byte{0x0})).ProcID())
	assert.Nil(t, m.SetMsgID("white space not possible").MsgID())
	assert.Nil(t, m.SetMsgID(string([]byte{0x0})).MsgID())
}

func TestSetMessage(t *testing.T) {
	m := &SyslogMessage{}
	greek := "κόσμε"
	assert.Equal(t, greek, *m.SetMessage(greek).Message())
}

func TestSetEmptyMessage(t *testing.T) {
	m := &SyslogMessage{}
	m.SetMessage("")
	assert.Nil(t, m.Message())
}

func TestSetWrongUTF8Message(t *testing.T) {}

func TestSetMessageWithBOM(t *testing.T) {}

func TestSetMessageWithNewline(t *testing.T) {}

func TestSetOutOfRangeVersion(t *testing.T) {
	m := &SyslogMessage{}
	m.SetVersion(1000)
	assert.Equal(t, m.Version(), uint16(0)) // 0 signals nil for version
	m.SetVersion(0)
	assert.Equal(t, m.Version(), uint16(0)) // 0 signals nil for version
}

func TestSetOutOfRangePriority(t *testing.T) {
	m := &SyslogMessage{}
	m.SetPriority(192)
	assert.Nil(t, m.Priority())
}

func TestSetVersion(t *testing.T) {
	m := &SyslogMessage{}
	m.SetVersion(1)
	assert.Equal(t, m.Version(), uint16(1))
	m.SetVersion(999)
	assert.Equal(t, m.Version(), uint16(999))
}

func TestSetPriority(t *testing.T) {
	m := &SyslogMessage{}
	m.SetPriority(0)
	assert.Equal(t, *m.Priority(), uint8(0))
	m.SetPriority(1)
	assert.Equal(t, *m.Priority(), uint8(1))
	m.SetPriority(191)
	assert.Equal(t, *m.Priority(), uint8(191))
}

func TestSetSDID(t *testing.T) {
	identifier := "one"
	m := &SyslogMessage{}
	assert.Nil(t, m.StructuredData())
	m.SetElementID(identifier)
	sd := m.StructuredData()
	assert.NotNil(t, sd)
	assert.IsType(t, (*map[string]map[string]string)(nil), sd)
	assert.NotNil(t, (*sd)[identifier])
	assert.IsType(t, map[string]string{}, (*sd)[identifier])
	m.SetElementID(identifier)
	assert.Len(t, *sd, 1)
}

func TestSetAllLenghtsSDID(t *testing.T) {
	m := &SyslogMessage{}

	maxlen := []byte("abcdefghilmnopqrstuvzabcdefghilm")

	prev := make([]byte, 0, len(maxlen))
	for i, input := range maxlen {
		prev = append(prev, input)
		id := string(prev)
		m.SetElementID(id)
		assert.Len(t, *m.StructuredData(), i+1)
		assert.IsType(t, map[string]string{}, (*m.StructuredData())[id])
	}
}

func TestSetTooLongSDID(t *testing.T) {
	m := &SyslogMessage{}
	m.SetElementID("abcdefghilmnopqrstuvzabcdefghilmX")
	assert.Nil(t, m.StructuredData())
}

func TestSetSyntacticallyWrongSDID(t *testing.T) {
	m := &SyslogMessage{}
	m.SetElementID("no whitespaces")
	assert.Nil(t, m.StructuredData())
	m.SetElementID(" ")
	assert.Nil(t, m.StructuredData())
	m.SetElementID(`"`)
	assert.Nil(t, m.StructuredData())
	m.SetElementID(`no"`)
	assert.Nil(t, m.StructuredData())
	m.SetElementID(`"no`)
	assert.Nil(t, m.StructuredData())
	m.SetElementID("]")
	assert.Nil(t, m.StructuredData())
	m.SetElementID("no]")
	assert.Nil(t, m.StructuredData())
	m.SetElementID("]no")
	assert.Nil(t, m.StructuredData())
}

func TestSetEmptySDID(t *testing.T) {
	m := &SyslogMessage{}
	m.SetElementID("")
	assert.Nil(t, m.StructuredData())
}

func TestSetSDParam(t *testing.T) {
	id := "one"
	pn := "pname"
	pv := "pvalue"
	m := &SyslogMessage{}
	m.SetParameter(id, pn, pv)
	sd := m.StructuredData()
	assert.NotNil(t, sd)
	assert.IsType(t, (*map[string]map[string]string)(nil), sd)
	assert.NotNil(t, (*sd)[id])
	assert.IsType(t, map[string]string{}, (*sd)[id])
	assert.Len(t, *sd, 1)
	assert.Len(t, (*sd)[id], 1)
	assert.Equal(t, pv, (*sd)[id][pn])

	pn1 := "pname1"
	pv1 := "κόσμε"
	m.SetParameter(id, pn1, pv1)
	assert.Len(t, (*sd)[id], 2)
	assert.Equal(t, pv1, (*sd)[id][pn1])

	id1 := "another"
	m.SetParameter(id1, pn1, pv1).SetParameter(id1, pn, pv)
	assert.Len(t, *sd, 2)
	assert.Len(t, (*sd)[id1], 2)
	assert.Equal(t, pv1, (*sd)[id1][pn1])
	assert.Equal(t, pv, (*sd)[id1][pn])

	id2 := "tre"
	pn2 := "meta"
	m.SetParameter(id2, pn, `valid\\`).SetParameter(id2, pn1, `\]valid`).SetParameter(id2, pn2, `is\"valid`)
	assert.Len(t, *sd, 3)
	assert.Len(t, (*sd)[id2], 3)
	assert.Equal(t, `valid\`, (*sd)[id2][pn])
	assert.Equal(t, `]valid`, (*sd)[id2][pn1])
	assert.Equal(t, `is"valid`, (*sd)[id2][pn2])
	// Cannot contain \, ], " unless escaped
	m.SetParameter(id2, pn, `is\valid`).SetParameter(id2, pn1, `is]valid`).SetParameter(id2, pn2, `is"valid`)
	assert.Len(t, (*sd)[id2], 3)
}

func TestSetEmptySDParam(t *testing.T) {
	id := "id"
	pn := "pn"
	m := &SyslogMessage{}
	m.SetParameter(id, pn, "")
	sd := m.StructuredData()
	assert.Len(t, *sd, 1)
	assert.Len(t, (*sd)[id], 1)
	assert.Equal(t, "", (*sd)[id][pn])
}

func TestSerialization(t *testing.T) {
	var res string
	var err error
	var pout *SyslogMessage
	var perr error

	p := NewParser()

	// Valid syslog message
	m := &SyslogMessage{}
	m.SetPriority(1)
	m.SetVersion(1)
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, "<1>1 - - - - - -", res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetMessage("-") // does not means nil in this case, remember
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, "<1>1 - - - - - - -", res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.
		SetParameter("mega", "x", "a").
		SetParameter("mega", "y", "b").
		SetParameter("mega", "z", `\" \] \\`).
		SetParameter("peta", "a", "name").
		SetParameter("giga", "1", "").
		SetParameter("peta", "c", "nomen")

	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 - - - - - [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetHostname("host1")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 - host1 - - - [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetAppname("su")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 - host1 su - - [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetProcID("22")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 - host1 su 22 - [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetMsgID("#1")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 - host1 su 22 #1 [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetTimestamp("2002-10-22T16:33:15.000087+01:00")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 2002-10-22T16:33:15.000087+01:00 host1 su 22 #1 [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] -`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	m.SetMessage("κόσμε")
	res, err = m.String()
	assert.Nil(t, err)
	assert.Equal(t, `<1>1 2002-10-22T16:33:15.000087+01:00 host1 su 22 #1 [giga 1=""][mega x="a" y="b" z="\" \] \\"][peta a="name" c="nomen"] κόσμε`, res)

	pout, perr = p.Parse([]byte(res), nil)
	assert.Equal(t, m, pout)
	assert.Nil(t, perr)

	// Invalid syslog message
	m2 := &SyslogMessage{}
	m2.SetPriority(192)
	m2.SetVersion(9999)
	res, err = m2.String()
	assert.Empty(t, res)
	assert.Error(t, err)
}
