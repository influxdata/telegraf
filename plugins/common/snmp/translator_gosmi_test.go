package snmp

import (
	"fmt"
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

func TestNewGosmiTranslator(t *testing.T) {
	var tr Translator
	var err error

	tr, err = NewGosmiTranslator([]string{"testdata"}, testutil.Logger{})
	require.NoError(t, err)
	require.NotNil(t, tr)
}

func TestFieldInitGosmi(t *testing.T) {
	tests := []struct {
		name     string
		input    Field
		expected Field
	}{
		{
			name:     "no change",
			input:    Field{Oid: ".1.2.3", Name: "foo"},
			expected: Field{Oid: ".1.2.3", Name: "foo"},
		},
		{
			name:     "OID translation",
			input:    Field{Oid: ".iso.2.3", Name: "foo"},
			expected: Field{Oid: ".1.2.3", Name: "foo"},
		},
		{
			name:     "numerical OID to name",
			input:    Field{Oid: ".1.0.0.0.1.1"},
			expected: Field{Oid: ".1.0.0.0.1.1", Name: "server"},
		},
		{
			name:     "numerical OID to name and conversion",
			input:    Field{Oid: ".1.0.0.0.1.5"},
			expected: Field{Oid: ".1.0.0.0.1.5", Name: "dateAndTime", Conversion: "displayhint"},
		},
		{
			name:     "textual OID",
			input:    Field{Oid: "IF-MIB::ifPhysAddress.1"},
			expected: Field{Oid: ".1.3.6.1.2.1.2.2.1.6.1", Name: "ifPhysAddress.1", Conversion: "displayhint"},
		},
		{
			name:     "textual OID no conversion",
			input:    Field{Oid: "IF-MIB::ifPhysAddress.1", Conversion: "none"},
			expected: Field{Oid: ".1.3.6.1.2.1.2.2.1.6.1", Name: "ifPhysAddress.1", Conversion: "none"},
		},
		{
			name:     "ipaddr conversion",
			input:    Field{Oid: "TCP-MIB::tcpConnectionLocalAddress.1"},
			expected: Field{Oid: ".1.3.6.1.2.1.6.19.1.2.1", Name: "tcpConnectionLocalAddress.1", Conversion: "ipaddr"},
		},
		{
			name:     "unknown OID",
			input:    Field{Oid: ".999"},
			expected: Field{Oid: ".999", Name: ".999"},
		},
	}

	tr := getGosmiTr(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.input
			require.NoError(t, f.Init(tr))

			require.EqualExportedValues(t, tt.expected, f)
			require.True(t, f.initialized)
		})
	}
}

func TestFieldInitFailGosmi(t *testing.T) {
	f := Field{
		Oid: "RFC1213-MIB::",
	}

	require.Error(t, f.Init(getGosmiTr(t)))
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

	require.NoError(t, tbl.Init(getGosmiTr(t)))

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
	require.Equal(t, "displayhint", tbl.Fields[2].Conversion)

	require.Equal(t, ".1.3.6.1.2.1.3.1.1.3", tbl.Fields[4].Oid)
	require.Equal(t, "atNetAddress", tbl.Fields[4].Name)
	require.True(t, tbl.Fields[4].IsTag)
	require.Empty(t, tbl.Fields[4].Conversion)
}

func TestTableBuildWalkGosmi(t *testing.T) {
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

func TestTableBuildNoWalkGosmi(t *testing.T) {
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
			{
				Name:      "myfield4",
				Oid:       ".1.3.6.1.2.1.3.1.1.3.0",
				Translate: true,
			},
		},
	}

	require.NoError(t, tbl.Init(getGosmiTr(t)))
	tb, err := tbl.Build(tsc, false)
	require.NoError(t, err)

	rtr := RTableRow{
		Tags:   map[string]string{"myfield1": "baz", "myfield3": "234"},
		Fields: map[string]interface{}{"myfield2": 234, "myfield4": "atNetAddress"},
	}
	require.Len(t, tb.Rows, 1)
	require.Contains(t, tb.Rows, rtr)
}

func TestFieldConvertGosmi(t *testing.T) {
	tests := []struct {
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
		{3, "displayhint", "testing(3)"},
	}

	tr := getGosmiTr(t)
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %[2]T %[2]v", tt.conv, tt.input), func(t *testing.T) {
			f := Field{
				Name:       "test",
				Conversion: tt.conv,
			}
			require.NoError(t, f.Init(tr))

			actual, err := f.Convert(gosnmp.SnmpPDU{Name: ".1.3.6.1.2.1.2.2.1.8", Value: tt.input})
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestSnmpFormatDisplayHintGosmi(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		input    interface{}
		expected string
	}{
		{
			name:     "ifOperStatus",
			oid:      ".1.3.6.1.2.1.2.2.1.8",
			input:    3,
			expected: "testing(3)",
		}, {
			name:     "ifPhysAddress",
			oid:      ".1.3.6.1.2.1.2.2.1.6",
			input:    []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
			expected: "01:23:45:67:89:ab:cd:ef",
		}, {
			name:     "DateAndTime short",
			oid:      ".1.0.0.0.1.5",
			input:    []byte{0x07, 0xe8, 0x09, 0x18, 0x10, 0x24, 0x27, 0x05},
			expected: "2024-9-24,16:36:39.5",
		}, {
			name:     "DateAndTime long",
			oid:      ".1.0.0.0.1.5",
			input:    []byte{0x07, 0xe8, 0x09, 0x18, 0x10, 0x24, 0x27, 0x05, 0x2b, 0x02, 0x00},
			expected: "2024-9-24,16:36:39.5,+2:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := getGosmiTr(t)

			actual, err := tr.SnmpFormatDisplayHint(tt.oid, tt.input)
			require.NoError(t, err)

			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestTableJoinWalkGosmi(t *testing.T) {
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

func TestTableOuterJoinWalkGosmi(t *testing.T) {
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

func TestTableJoinNoIndexAsTagWalkGosmi(t *testing.T) {
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
			// "index":    "10",
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
			// "index":    "11",
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
			// "index":    "12",
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

func TestTrapLookupGosmi(t *testing.T) {
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

func TestTrapLookupFailGosmi(t *testing.T) {
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
