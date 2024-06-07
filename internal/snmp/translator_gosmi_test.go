package snmp

import (
	"path/filepath"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func getGosmiTr(t *testing.T) Translator {
	testDataPath, err := filepath.Abs("./testdata/gosmi")
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

func TestFieldInitGosmi(t *testing.T) {
	tr := getGosmiTr(t)

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
		require.NoError(t, f.Init(tr), "inputOid=%q inputName=%q", txl.inputOid, txl.inputName)

		require.Equal(t, txl.expectedOid, f.Oid, "inputOid=%q inputName=%q inputConversion=%q", txl.inputOid, txl.inputName, txl.inputConversion)
		require.Equal(t, txl.expectedName, f.Name, "inputOid=%q inputName=%q inputConversion=%q", txl.inputOid, txl.inputName, txl.inputConversion)
		require.Equal(t, txl.expectedConversion, f.Conversion, "inputOid=%q inputName=%q inputConversion=%q", txl.inputOid, txl.inputName, txl.inputConversion)
	}
}

func TestTableInitGosmi(t *testing.T) {
	tbl := Table{
		Oid: ".1.3.6.1.2.1.3.1",
		Fields: []Field{
			{Oid: ".999", Name: "foo"},
			{Oid: ".1.3.6.1.2.1.3.1.1.1", Name: "atIfIndex", IsTag: true},
			{Oid: "RFC1213-MIB::atPhysAddress", Name: "atPhysAddress"},
		},
	}

	tr := getGosmiTr(t)
	require.NoError(t, tbl.Init(tr))

	require.Equal(t, "atTable", tbl.Name)

	require.Len(t, tbl.Fields, 5)

	require.Equal(t, ".999", tbl.Fields[0].Oid)
	require.Equal(t, "foo", tbl.Fields[0].Name)
	require.False(t, tbl.Fields[0].IsTag)
	require.Empty(t, tbl.Fields[0].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.1", tbl.Fields[1].Oid)
	require.Equal(t, "atIfIndex", tbl.Fields[1].Name)
	require.True(t, tbl.Fields[1].IsTag)
	require.Empty(t, tbl.Fields[1].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.2", tbl.Fields[2].Oid)
	require.Equal(t, "atPhysAddress", tbl.Fields[2].Name)
	require.False(t, tbl.Fields[2].IsTag)
	require.Equal(t, "hwaddr", tbl.Fields[2].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.3", tbl.Fields[4].Oid)
	require.Equal(t, "atNetAddress", tbl.Fields[4].Name)
	require.True(t, tbl.Fields[4].IsTag)
	require.Empty(t, tbl.Fields[4].Conversion)
}

// TestTableBuild_walk in snmp_test.go is split into two tests here,
// noTranslate and Translate.
//
// This is only running with gosmi translator but should be valid with
// netsnmp too.
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

	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)
	require.Equal(t, "mytable", tb.Name)
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
	require.Len(t, tb.Rows, 4)
	require.Contains(t, tb.Rows, rtr1)
	require.Contains(t, tb.Rows, rtr2)
	require.Contains(t, tb.Rows, rtr3)
	require.Contains(t, tb.Rows, rtr4)
}

func TestTableBuild_walk_Translate(t *testing.T) {
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

	require.NoError(t, tbl.Init(getGosmiTr(t)))
	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)

	require.Equal(t, "atTable", tb.Name)

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

	tb, err := tbl.Build(tsc, false)
	require.NoError(t, err)

	rtr := RTableRow{
		Tags:   map[string]string{"myfield1": "baz", "myfield3": "234"},
		Fields: map[string]interface{}{"myfield2": 234},
	}
	require.Len(t, tb.Rows, 1)
	require.Contains(t, tb.Rows, rtr)
}

func TestFieldConvertGosmi(t *testing.T) {
	testTable := []struct {
		input    interface{}
		conv     string
		expected interface{}
	}{
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
		{3, "enum", "testing"},
		{3, "enum(1)", "testing(3)"},
	}

	for _, tc := range testTable {
		f := Field{
			Name:       "test",
			Conversion: tc.conv,
		}
		require.NoError(t, f.Init(getGosmiTr(t)))

		act, err := f.Convert(gosnmp.SnmpPDU{Name: ".1.3.6.1.2.1.2.2.1.8", Value: tc.input})
		require.NoError(t, err, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected)
		require.EqualValues(t, tc.expected, act, "input=%T(%v) conv=%s expected=%T(%v)", tc.input, tc.input, tc.conv, tc.expected, tc.expected)
	}
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

	require.NoError(t, tbl.Init(getGosmiTr(t)))
	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)

	require.Equal(t, "mytable", tb.Name)
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
	require.Len(t, tb.Rows, 3)
	require.Contains(t, tb.Rows, rtr1)
	require.Contains(t, tb.Rows, rtr2)
	require.Contains(t, tb.Rows, rtr3)
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

	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)

	require.Equal(t, "mytable", tb.Name)
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
	require.Len(t, tb.Rows, 4)
	require.Contains(t, tb.Rows, rtr1)
	require.Contains(t, tb.Rows, rtr2)
	require.Contains(t, tb.Rows, rtr3)
	require.Contains(t, tb.Rows, rtr4)
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

	tb, err := tbl.Build(tsc, true)
	require.NoError(t, err)

	require.Equal(t, "mytable", tb.Name)
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
	require.Len(t, tb.Rows, 3)
	require.Contains(t, tb.Rows, rtr1)
	require.Contains(t, tb.Rows, rtr2)
	require.Contains(t, tb.Rows, rtr3)
}

func TestCanNotParse(t *testing.T) {
	tr := getGosmiTr(t)
	f := Field{
		Oid: "RFC1213-MIB::",
	}

	require.Error(t, f.Init(tr))
}

func TestTrapLookup(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected MibEntry
	}{
		{
			name: "Known trap OID",
			oid:  ".1.3.6.1.6.3.1.1.5.1",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "coldStart",
			},
		},
		{
			name: "Known trap value OID",
			oid:  ".1.3.6.1.2.1.1.3.0",
			expected: MibEntry{
				MibName: "TGTEST-MIB",
				OidText: "sysUpTimeInstance",
			},
		},
		{
			name: "Unknown enterprise sub-OID",
			oid:  ".1.3.6.1.4.1.0.1.2.3",
			expected: MibEntry{
				MibName: "SNMPv2-SMI",
				OidText: "enterprises.0.1.2.3",
			},
		},
		{
			name:     "Unknown MIB",
			oid:      ".1.999",
			expected: MibEntry{OidText: "iso.999"},
		},
	}

	// Load the MIBs
	getGosmiTr(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			actual, err := TrapLookup(tt.oid)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestTrapLookupFail(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected string
	}{
		{
			name:     "New top level OID",
			oid:      ".3.6.1.3.0",
			expected: "Could not find node for OID 3.6.1.3.0",
		},
		{
			name:     "Malformed OID",
			oid:      ".1.3.dod.1.3.0",
			expected: "could not convert OID .1.3.dod.1.3.0: strconv.ParseUint: parsing \"dod\": invalid syntax",
		},
	}

	// Load the MIBs
	getGosmiTr(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the actual test
			_, err := TrapLookup(tt.oid)
			require.EqualError(t, err, tt.expected)
		})
	}
}
