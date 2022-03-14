package snmp

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/testutil"
)

func getGosmiTr(t *testing.T) Translator {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)
	return tr
}

func TestGosmiTranslator(t *testing.T) {
	var tr Translator
	var err error

	tr, err = NewGosmiTranslator([]string{"testdata"}, testutil.Logger{})
	require.NoError(t, err)
	require.NotNil(t, tr)
}

//gosmi uses the same connection struct as netsnmp but has a few
//different test cases, so it has its own copy
var gosmiTsc = &testSNMPConnection{
	host: "tsc",
	values: map[string]interface{}{
		".1.3.6.1.2.1.3.1.1.1.0": "foo",
		".1.3.6.1.2.1.3.1.1.1.1": []byte("bar"),
		".1.3.6.1.2.1.3.1.1.1.2": []byte(""),
		".1.3.6.1.2.1.3.1.1.102": "bad",
		".1.3.6.1.2.1.3.1.1.2.0": 1,
		".1.3.6.1.2.1.3.1.1.2.1": 2,
		".1.3.6.1.2.1.3.1.1.2.2": 0,
		".1.3.6.1.2.1.3.1.1.3.0": "1.3.6.1.2.1.3.1.1.3",
		".1.3.6.1.2.1.3.1.1.5.0": 123456,
		".1.0.0.0.1.1.0":         "foo",
		".1.0.0.0.1.1.1":         []byte("bar"),
		".1.0.0.0.1.1.2":         []byte(""),
		".1.0.0.0.1.102":         "bad",
		".1.0.0.0.1.2.0":         1,
		".1.0.0.0.1.2.1":         2,
		".1.0.0.0.1.2.2":         0,
		".1.0.0.0.1.3.0":         "0.123",
		".1.0.0.0.1.3.1":         "0.456",
		".1.0.0.0.1.3.2":         "0.000",
		".1.0.0.0.1.3.3":         "9.999",
		".1.0.0.0.1.5.0":         123456,
		".1.0.0.1.1":             "baz",
		".1.0.0.1.2":             234,
		".1.0.0.1.3":             []byte("byte slice"),
		".1.0.0.2.1.5.0.9.9":     11,
		".1.0.0.2.1.5.1.9.9":     22,
		".1.0.0.0.1.6.0":         ".1.0.0.0.1.7",
		".1.0.0.3.1.1.10":        "instance",
		".1.0.0.3.1.1.11":        "instance2",
		".1.0.0.3.1.1.12":        "instance3",
		".1.0.0.3.1.2.10":        10,
		".1.0.0.3.1.2.11":        20,
		".1.0.0.3.1.2.12":        20,
		".1.0.0.3.1.3.10":        1,
		".1.0.0.3.1.3.11":        2,
		".1.0.0.3.1.3.12":        3,
	},
}

func TestFieldInitGosmi(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	translations := []struct {
		inputOid           string
		inputName          string
		inputConversion    string
		expectedOid        string
		expectedName       string
		expectedConversion string
	}{
		{".1.2.3", "foo", "", ".1.2.3", "foo", ""},
		{".iso.2.3", "foo", "", ".1.2.3", "foo", ""},
		{".1.0.0.0.1.1", "", "", ".1.0.0.0.1.1", "server", ""},
		{"IF-MIB::ifPhysAddress.1", "", "", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "hwaddr"},
		{"IF-MIB::ifPhysAddress.1", "", "none", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "none"},
		{"BRIDGE-MIB::dot1dTpFdbAddress.1", "", "", ".1.3.6.1.2.1.17.4.3.1.1.1", "dot1dTpFdbAddress.1", "hwaddr"},
		{"TCP-MIB::tcpConnectionLocalAddress.1", "", "", ".1.3.6.1.2.1.6.19.1.2.1", "tcpConnectionLocalAddress.1", "ipaddr"},
		{".999", "", "", ".999", ".999", ""},
	}

	for _, txl := range translations {
		f := Field{Oid: txl.inputOid, Name: txl.inputName, Conversion: txl.inputConversion}
		err := f.init(tr)
		require.NoError(t, err, "inputOid='%s' inputName='%s'", txl.inputOid, txl.inputName)

		assert.Equal(t, txl.expectedOid, f.Oid, "inputOid='%s' inputName='%s' inputConversion='%s'", txl.inputOid, txl.inputName, txl.inputConversion)
		assert.Equal(t, txl.expectedName, f.Name, "inputOid='%s' inputName='%s' inputConversion='%s'", txl.inputOid, txl.inputName, txl.inputConversion)
	}
}

func TestTableInitGosmi(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	s := &Snmp{
		ClientConfig: snmp.ClientConfig{
			Path:       []string{testDataPath},
			Translator: "gosmi",
		},
		Tables: []Table{
			{Oid: ".1.3.6.1.2.1.3.1",
				Fields: []Field{
					{Oid: ".999", Name: "foo"},
					{Oid: ".1.3.6.1.2.1.3.1.1.1", Name: "atIfIndex", IsTag: true},
					{Oid: "RFC1213-MIB::atPhysAddress", Name: "atPhysAddress"},
				}},
		},
	}
	err = s.Init()
	require.NoError(t, err)

	assert.Equal(t, "atTable", s.Tables[0].Name)

	assert.Len(t, s.Tables[0].Fields, 5)
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".999", Name: "foo", initialized: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.1", Name: "atIfIndex", initialized: true, IsTag: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.2", Name: "atPhysAddress", initialized: true, Conversion: "hwaddr"})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.3", Name: "atNetAddress", initialized: true, IsTag: true})
}

func TestSnmpInitGosmi(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	s := &Snmp{
		Tables: []Table{
			{Oid: "RFC1213-MIB::atTable"},
		},
		Fields: []Field{
			{Oid: "RFC1213-MIB::atPhysAddress"},
		},
		ClientConfig: snmp.ClientConfig{
			Path:       []string{testDataPath},
			Translator: "gosmi",
		},
	}

	err = s.Init()
	require.NoError(t, err)

	assert.Len(t, s.Tables[0].Fields, 3)
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.1", Name: "atIfIndex", IsTag: true, initialized: true})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.2", Name: "atPhysAddress", initialized: true, Conversion: "hwaddr"})
	assert.Contains(t, s.Tables[0].Fields, Field{Oid: ".1.3.6.1.2.1.3.1.1.3", Name: "atNetAddress", IsTag: true, initialized: true})

	assert.Equal(t, Field{
		Oid:         ".1.3.6.1.2.1.3.1.1.2",
		Name:        "atPhysAddress",
		Conversion:  "hwaddr",
		initialized: true,
	}, s.Fields[0])
}

func TestSnmpInit_noTranslateGosmi(t *testing.T) {
	s := &Snmp{
		Fields: []Field{
			{Oid: ".9.1.1.1.1", Name: "one", IsTag: true},
			{Oid: ".9.1.1.1.2", Name: "two"},
			{Oid: ".9.1.1.1.3"},
		},
		Tables: []Table{
			{Name: "testing",
				Fields: []Field{
					{Oid: ".9.1.1.1.4", Name: "four", IsTag: true},
					{Oid: ".9.1.1.1.5", Name: "five"},
					{Oid: ".9.1.1.1.6"},
				}},
		},
		ClientConfig: snmp.ClientConfig{
			Path:       []string{},
			Translator: "gosmi",
		},
	}

	err := s.Init()
	require.NoError(t, err)

	assert.Equal(t, ".9.1.1.1.1", s.Fields[0].Oid)
	assert.Equal(t, "one", s.Fields[0].Name)
	assert.Equal(t, true, s.Fields[0].IsTag)

	assert.Equal(t, ".9.1.1.1.2", s.Fields[1].Oid)
	assert.Equal(t, "two", s.Fields[1].Name)
	assert.Equal(t, false, s.Fields[1].IsTag)

	assert.Equal(t, ".9.1.1.1.3", s.Fields[2].Oid)
	assert.Equal(t, ".9.1.1.1.3", s.Fields[2].Name)
	assert.Equal(t, false, s.Fields[2].IsTag)

	assert.Equal(t, ".9.1.1.1.4", s.Tables[0].Fields[0].Oid)
	assert.Equal(t, "four", s.Tables[0].Fields[0].Name)
	assert.Equal(t, true, s.Tables[0].Fields[0].IsTag)

	assert.Equal(t, ".9.1.1.1.5", s.Tables[0].Fields[1].Oid)
	assert.Equal(t, "five", s.Tables[0].Fields[1].Name)
	assert.Equal(t, false, s.Tables[0].Fields[1].IsTag)

	assert.Equal(t, ".9.1.1.1.6", s.Tables[0].Fields[2].Oid)
	assert.Equal(t, ".9.1.1.1.6", s.Tables[0].Fields[2].Name)
	assert.Equal(t, false, s.Tables[0].Fields[2].IsTag)
}

//TestTableBuild_walk in snmp_test.go is split into two tests here,
//noTranslate and Translate.
//
//This is only running with gosmi translator but should be valid with
//netsnmp too.
func TestTableBuild_walk_noTranslate(t *testing.T) {
	tbl := Table{
		Name:       "mytable",
		IndexAsTag: true,
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.0.1.2",
			},
			{
				Name:       "myfield3",
				Oid:        ".1.0.0.0.1.3",
				Conversion: "float",
			},
			{
				Name:           "myfield4",
				Oid:            ".1.0.0.2.1.5",
				OidIndexSuffix: ".9.9",
			},
			{
				Name:           "myfield5",
				Oid:            ".1.0.0.2.1.5",
				OidIndexLength: 1,
			},
		},
	}

	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tb, err := tbl.Build(gosmiTsc, true, tr)
	require.NoError(t, err)
	assert.Equal(t, tb.Name, "mytable")
	rtr1 := RTableRow{
		Tags: map[string]string{
			"myfield1": "foo",
			"index":    "0",
		},
		Fields: map[string]interface{}{
			"myfield2": 1,
			"myfield3": float64(0.123),
			"myfield4": 11,
			"myfield5": 11,
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"myfield1": "bar",
			"index":    "1",
		},
		Fields: map[string]interface{}{
			"myfield2": 2,
			"myfield3": float64(0.456),
			"myfield4": 22,
			"myfield5": 22,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"index": "2",
		},
		Fields: map[string]interface{}{
			"myfield2": 0,
			"myfield3": float64(0.0),
		},
	}
	rtr4 := RTableRow{
		Tags: map[string]string{
			"index": "3",
		},
		Fields: map[string]interface{}{
			"myfield3": float64(9.999),
		},
	}
	assert.Len(t, tb.Rows, 4)
	assert.Contains(t, tb.Rows, rtr1)
	assert.Contains(t, tb.Rows, rtr2)
	assert.Contains(t, tb.Rows, rtr3)
	assert.Contains(t, tb.Rows, rtr4)
}

func TestTableBuild_walk_Translate(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tbl := Table{
		Name:       "atTable",
		IndexAsTag: true,
		Fields: []Field{
			{
				Name:  "ifIndex",
				Oid:   "1.3.6.1.2.1.3.1.1.1",
				IsTag: true,
			},
			{
				Name:      "atPhysAddress",
				Oid:       "1.3.6.1.2.1.3.1.1.2",
				Translate: false,
			},
			{
				Name:      "atNetAddress",
				Oid:       "1.3.6.1.2.1.3.1.1.3",
				Translate: true,
			},
		},
	}

	err = tbl.Init(tr)
	require.NoError(t, err)
	tb, err := tbl.Build(gosmiTsc, true, tr)
	require.NoError(t, err)

	require.Equal(t, tb.Name, "atTable")

	rtr1 := RTableRow{
		Tags: map[string]string{
			"ifIndex": "foo",
			"index":   "0",
		},
		Fields: map[string]interface{}{
			"atPhysAddress": 1,
			"atNetAddress":  "atNetAddress",
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"ifIndex": "bar",
			"index":   "1",
		},
		Fields: map[string]interface{}{
			"atPhysAddress": 2,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"index": "2",
		},
		Fields: map[string]interface{}{
			"atPhysAddress": 0,
		},
	}

	require.Len(t, tb.Rows, 3)
	require.Contains(t, tb.Rows, rtr1)
	require.Contains(t, tb.Rows, rtr2)
	require.Contains(t, tb.Rows, rtr3)
}

func TestTableBuild_noWalkGosmi(t *testing.T) {
	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tbl := Table{
		Name: "mytable",
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
			{
				Name:  "myfield3",
				Oid:   ".1.0.0.1.2",
				IsTag: true,
			},
			{
				Name: "empty",
				Oid:  ".1.0.0.0.1.1.2",
			},
			{
				Name: "noexist",
				Oid:  ".1.2.3.4.5",
			},
		},
	}

	tb, err := tbl.Build(gosmiTsc, false, tr)
	require.NoError(t, err)

	rtr := RTableRow{
		Tags:   map[string]string{"myfield1": "baz", "myfield3": "234"},
		Fields: map[string]interface{}{"myfield2": 234},
	}
	assert.Len(t, tb.Rows, 1)
	assert.Contains(t, tb.Rows, rtr)
}

func TestGatherGosmi(t *testing.T) {
	s := &Snmp{
		Agents: []string{"TestGather"},
		Name:   "mytable",
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
			{
				Name: "myfield3",
				Oid:  "1.0.0.1.1",
			},
		},
		Tables: []Table{
			{
				Name:        "myOtherTable",
				InheritTags: []string{"myfield1"},
				Fields: []Field{
					{
						Name: "myOtherField",
						Oid:  ".1.0.0.0.1.5",
					},
				},
			},
		},

		connectionCache: []snmpConnection{
			gosmiTsc,
		},

		ClientConfig: snmp.ClientConfig{
			Path:       []string{"testdata"},
			Translator: "gosmi",
		},
	}
	acc := &testutil.Accumulator{}

	tstart := time.Now()
	require.NoError(t, s.Gather(acc))
	tstop := time.Now()

	require.Len(t, acc.Metrics, 2)

	m := acc.Metrics[0]
	assert.Equal(t, "mytable", m.Measurement)
	assert.Equal(t, "tsc", m.Tags[s.AgentHostTag])
	assert.Equal(t, "baz", m.Tags["myfield1"])
	assert.Len(t, m.Fields, 2)
	assert.Equal(t, 234, m.Fields["myfield2"])
	assert.Equal(t, "baz", m.Fields["myfield3"])
	assert.False(t, tstart.After(m.Time))
	assert.False(t, tstop.Before(m.Time))

	m2 := acc.Metrics[1]
	assert.Equal(t, "myOtherTable", m2.Measurement)
	assert.Equal(t, "tsc", m2.Tags[s.AgentHostTag])
	assert.Equal(t, "baz", m2.Tags["myfield1"])
	assert.Len(t, m2.Fields, 1)
	assert.Equal(t, 123456, m2.Fields["myOtherField"])
}

func TestGather_hostGosmi(t *testing.T) {
	s := &Snmp{
		Agents: []string{"TestGather"},
		Name:   "mytable",
		Fields: []Field{
			{
				Name:  "host",
				Oid:   ".1.0.0.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.1.2",
			},
		},

		connectionCache: []snmpConnection{
			gosmiTsc,
		},
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, s.Gather(acc))

	require.Len(t, acc.Metrics, 1)
	m := acc.Metrics[0]
	assert.Equal(t, "baz", m.Tags["host"])
}

func TestFieldConvertGosmi(t *testing.T) {
	testTable := []struct {
		input    interface{}
		conv     string
		expected interface{}
	}{
		{[]byte("foo"), "", "foo"},
		{"0.123", "float", float64(0.123)},
		{[]byte("0.123"), "float", float64(0.123)},
		{float32(0.123), "float", float64(float32(0.123))},
		{float64(0.123), "float", float64(0.123)},
		{float64(0.123123123123), "float", float64(0.123123123123)},
		{123, "float", float64(123)},
		{123, "float(0)", float64(123)},
		{123, "float(4)", float64(0.0123)},
		{int8(123), "float(3)", float64(0.123)},
		{int16(123), "float(3)", float64(0.123)},
		{int32(123), "float(3)", float64(0.123)},
		{int64(123), "float(3)", float64(0.123)},
		{uint(123), "float(3)", float64(0.123)},
		{uint8(123), "float(3)", float64(0.123)},
		{uint16(123), "float(3)", float64(0.123)},
		{uint32(123), "float(3)", float64(0.123)},
		{uint64(123), "float(3)", float64(0.123)},
		{"123", "int", int64(123)},
		{[]byte("123"), "int", int64(123)},
		{"123123123123", "int", int64(123123123123)},
		{[]byte("123123123123"), "int", int64(123123123123)},
		{float32(12.3), "int", int64(12)},
		{float64(12.3), "int", int64(12)},
		{123, "int", int64(123)},
		{int8(123), "int", int64(123)},
		{int16(123), "int", int64(123)},
		{int32(123), "int", int64(123)},
		{int64(123), "int", int64(123)},
		{uint(123), "int", int64(123)},
		{uint8(123), "int", int64(123)},
		{uint16(123), "int", int64(123)},
		{uint32(123), "int", int64(123)},
		{uint64(123), "int", int64(123)},
		{[]byte("abcdef"), "hwaddr", "61:62:63:64:65:66"},
		{"abcdef", "hwaddr", "61:62:63:64:65:66"},
		{[]byte("abcd"), "ipaddr", "97.98.99.100"},
		{"abcd", "ipaddr", "97.98.99.100"},
		{[]byte("abcdefghijklmnop"), "ipaddr", "6162:6364:6566:6768:696a:6b6c:6d6e:6f70"},
		{[]byte{0x00, 0x09, 0x3E, 0xE3, 0xF6, 0xD5, 0x3B, 0x60}, "hextoint:BigEndian:uint64", uint64(2602423610063712)},
		{[]byte{0x00, 0x09, 0x3E, 0xE3}, "hextoint:BigEndian:uint32", uint32(605923)},
		{[]byte{0x00, 0x09}, "hextoint:BigEndian:uint16", uint16(9)},
		{[]byte{0x00, 0x09, 0x3E, 0xE3, 0xF6, 0xD5, 0x3B, 0x60}, "hextoint:LittleEndian:uint64", uint64(6934371307618175232)},
		{[]byte{0x00, 0x09, 0x3E, 0xE3}, "hextoint:LittleEndian:uint32", uint32(3812493568)},
		{[]byte{0x00, 0x09}, "hextoint:LittleEndian:uint16", uint16(2304)},
	}

	for _, tc := range testTable {
		act, err := fieldConvert(tc.conv, tc.input)
		assert.NoError(t, err, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected)
		assert.EqualValues(t, tc.expected, act, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected)
	}
}

func TestSnmpTranslateCache_missGosmi(t *testing.T) {
	gosmiSnmpTranslateCaches = nil
	oid := "IF-MIB::ifPhysAddress.1"
	mibName, oidNum, oidText, conversion, err := getGosmiTr(t).SnmpTranslate(oid)
	assert.Len(t, gosmiSnmpTranslateCaches, 1)
	stc := gosmiSnmpTranslateCaches[oid]
	assert.NotNil(t, stc)
	assert.Equal(t, mibName, stc.mibName)
	assert.Equal(t, oidNum, stc.oidNum)
	assert.Equal(t, oidText, stc.oidText)
	assert.Equal(t, conversion, stc.conversion)
	assert.Equal(t, err, stc.err)
}

func TestSnmpTranslateCache_hitGosmi(t *testing.T) {
	gosmiSnmpTranslateCaches = map[string]gosmiSnmpTranslateCache{
		"foo": {
			mibName:    "a",
			oidNum:     "b",
			oidText:    "c",
			conversion: "d",
			err:        fmt.Errorf("e"),
		},
	}
	mibName, oidNum, oidText, conversion, err := getGosmiTr(t).SnmpTranslate("foo")
	assert.Equal(t, "a", mibName)
	assert.Equal(t, "b", oidNum)
	assert.Equal(t, "c", oidText)
	assert.Equal(t, "d", conversion)
	assert.Equal(t, fmt.Errorf("e"), err)
	gosmiSnmpTranslateCaches = nil
}

func TestSnmpTableCache_missGosmi(t *testing.T) {
	gosmiSnmpTableCaches = nil
	oid := ".1.0.0.0"
	mibName, oidNum, oidText, fields, err := getGosmiTr(t).SnmpTable(oid)
	assert.Len(t, gosmiSnmpTableCaches, 1)
	stc := gosmiSnmpTableCaches[oid]
	assert.NotNil(t, stc)
	assert.Equal(t, mibName, stc.mibName)
	assert.Equal(t, oidNum, stc.oidNum)
	assert.Equal(t, oidText, stc.oidText)
	assert.Equal(t, fields, stc.fields)
	assert.Equal(t, err, stc.err)
}

func TestSnmpTableCache_hitGosmi(t *testing.T) {
	gosmiSnmpTableCaches = map[string]gosmiSnmpTableCache{
		"foo": {
			mibName: "a",
			oidNum:  "b",
			oidText: "c",
			fields:  []Field{{Name: "d"}},
			err:     fmt.Errorf("e"),
		},
	}
	mibName, oidNum, oidText, fields, err := getGosmiTr(t).SnmpTable("foo")
	assert.Equal(t, "a", mibName)
	assert.Equal(t, "b", oidNum)
	assert.Equal(t, "c", oidText)
	assert.Equal(t, []Field{{Name: "d"}}, fields)
	assert.Equal(t, fmt.Errorf("e"), err)
}

func TestTableJoin_walkGosmi(t *testing.T) {
	tbl := Table{
		Name:       "mytable",
		IndexAsTag: true,
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.3.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.3.1.2",
			},
			{
				Name:                "myfield3",
				Oid:                 ".1.0.0.3.1.3",
				SecondaryIndexTable: true,
			},
			{
				Name:              "myfield4",
				Oid:               ".1.0.0.0.1.1",
				SecondaryIndexUse: true,
				IsTag:             true,
			},
			{
				Name:              "myfield5",
				Oid:               ".1.0.0.0.1.2",
				SecondaryIndexUse: true,
			},
		},
	}

	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tb, err := tbl.Build(gosmiTsc, true, tr)
	require.NoError(t, err)

	assert.Equal(t, tb.Name, "mytable")
	rtr1 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance",
			"myfield4": "bar",
			"index":    "10",
		},
		Fields: map[string]interface{}{
			"myfield2": 10,
			"myfield3": 1,
			"myfield5": 2,
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance2",
			"index":    "11",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 2,
			"myfield5": 0,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance3",
			"index":    "12",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 3,
		},
	}
	assert.Len(t, tb.Rows, 3)
	assert.Contains(t, tb.Rows, rtr1)
	assert.Contains(t, tb.Rows, rtr2)
	assert.Contains(t, tb.Rows, rtr3)
}

func TestTableOuterJoin_walkGosmi(t *testing.T) {
	tbl := Table{
		Name:       "mytable",
		IndexAsTag: true,
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.3.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.3.1.2",
			},
			{
				Name:                "myfield3",
				Oid:                 ".1.0.0.3.1.3",
				SecondaryIndexTable: true,
				SecondaryOuterJoin:  true,
			},
			{
				Name:              "myfield4",
				Oid:               ".1.0.0.0.1.1",
				SecondaryIndexUse: true,
				IsTag:             true,
			},
			{
				Name:              "myfield5",
				Oid:               ".1.0.0.0.1.2",
				SecondaryIndexUse: true,
			},
		},
	}

	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tb, err := tbl.Build(gosmiTsc, true, tr)
	require.NoError(t, err)

	assert.Equal(t, tb.Name, "mytable")
	rtr1 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance",
			"myfield4": "bar",
			"index":    "10",
		},
		Fields: map[string]interface{}{
			"myfield2": 10,
			"myfield3": 1,
			"myfield5": 2,
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance2",
			"index":    "11",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 2,
			"myfield5": 0,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance3",
			"index":    "12",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 3,
		},
	}
	rtr4 := RTableRow{
		Tags: map[string]string{
			"index":    "Secondary.0",
			"myfield4": "foo",
		},
		Fields: map[string]interface{}{
			"myfield5": 1,
		},
	}
	assert.Len(t, tb.Rows, 4)
	assert.Contains(t, tb.Rows, rtr1)
	assert.Contains(t, tb.Rows, rtr2)
	assert.Contains(t, tb.Rows, rtr3)
	assert.Contains(t, tb.Rows, rtr4)
}

func TestTableJoinNoIndexAsTag_walkGosmi(t *testing.T) {
	tbl := Table{
		Name:       "mytable",
		IndexAsTag: false,
		Fields: []Field{
			{
				Name:  "myfield1",
				Oid:   ".1.0.0.3.1.1",
				IsTag: true,
			},
			{
				Name: "myfield2",
				Oid:  ".1.0.0.3.1.2",
			},
			{
				Name:                "myfield3",
				Oid:                 ".1.0.0.3.1.3",
				SecondaryIndexTable: true,
			},
			{
				Name:              "myfield4",
				Oid:               ".1.0.0.0.1.1",
				SecondaryIndexUse: true,
				IsTag:             true,
			},
			{
				Name:              "myfield5",
				Oid:               ".1.0.0.0.1.2",
				SecondaryIndexUse: true,
			},
		},
	}

	testDataPath, err := filepath.Abs("./testdata")
	require.NoError(t, err)

	tr, err := NewGosmiTranslator([]string{testDataPath}, testutil.Logger{})
	require.NoError(t, err)

	tb, err := tbl.Build(gosmiTsc, true, tr)
	require.NoError(t, err)

	assert.Equal(t, tb.Name, "mytable")
	rtr1 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance",
			"myfield4": "bar",
			//"index":    "10",
		},
		Fields: map[string]interface{}{
			"myfield2": 10,
			"myfield3": 1,
			"myfield5": 2,
		},
	}
	rtr2 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance2",
			//"index":    "11",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 2,
			"myfield5": 0,
		},
	}
	rtr3 := RTableRow{
		Tags: map[string]string{
			"myfield1": "instance3",
			//"index":    "12",
		},
		Fields: map[string]interface{}{
			"myfield2": 20,
			"myfield3": 3,
		},
	}
	assert.Len(t, tb.Rows, 3)
	assert.Contains(t, tb.Rows, rtr1)
	assert.Contains(t, tb.Rows, rtr2)
	assert.Contains(t, tb.Rows, rtr3)
}

func BenchmarkMibLoading(b *testing.B) {
	log := testutil.Logger{}
	path := []string{"testdata"}
	for i := 0; i < b.N; i++ {
		err := snmp.LoadMibsFromPath(path, log, &snmp.GosmiMibLoader{})
		require.NoError(b, err)
	}
}

func TestCanNotParse(t *testing.T) {
	s := &Snmp{
		Fields: []Field{
			{Oid: "RFC1213-MIB::"},
		},
		ClientConfig: snmp.ClientConfig{
			Path:       []string{"testdata"},
			Translator: "gosmi",
		},
	}

	err := s.Init()
	require.Error(t, err)
}

func TestMissingMibPath(t *testing.T) {
	log := testutil.Logger{}
	path := []string{"non-existing-directory"}
	err := snmp.LoadMibsFromPath(path, log, &snmp.GosmiMibLoader{})
	require.NoError(t, err)
}
