//go:generate go run -tags generate translator_netsnmp_mocks_generate.go
package snmp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestFieldInit(t *testing.T) {
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
		{".1.0.0.0.1.1.0", "", "", ".1.0.0.0.1.1.0", "server.0", ""},
		{".999", "", "", ".999", ".999", ""},
		{"TEST::server", "", "", ".1.0.0.0.1.1", "server", ""},
		{"TEST::server.0", "", "", ".1.0.0.0.1.1.0", "server.0", ""},
		{"TEST::server", "foo", "", ".1.0.0.0.1.1", "foo", ""},
		{"IF-MIB::ifPhysAddress.1", "", "", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "hwaddr"},
		{"IF-MIB::ifPhysAddress.1", "", "none", ".1.3.6.1.2.1.2.2.1.6.1", "ifPhysAddress.1", "none"},
		{"BRIDGE-MIB::dot1dTpFdbAddress.1", "", "", ".1.3.6.1.2.1.17.4.3.1.1.1", "dot1dTpFdbAddress.1", "hwaddr"},
		{"TCP-MIB::tcpConnectionLocalAddress.1", "", "", ".1.3.6.1.2.1.6.19.1.2.1", "tcpConnectionLocalAddress.1", "ipaddr"},
	}

	tr := NewNetsnmpTranslator(testutil.Logger{})
	for _, txl := range translations {
		f := Field{Oid: txl.inputOid, Name: txl.inputName, Conversion: txl.inputConversion}
		err := f.Init(tr)
		require.NoError(t, err, "inputOid=%q inputName=%q", txl.inputOid, txl.inputName)
		require.Equal(t, txl.expectedOid, f.Oid, "inputOid=%q inputName=%q inputConversion=%q", txl.inputOid, txl.inputName, txl.inputConversion)
		require.Equal(t, txl.expectedName, f.Name, "inputOid=%q inputName=%q inputConversion=%q", txl.inputOid, txl.inputName, txl.inputConversion)
	}
}

func TestTableInit(t *testing.T) {
	tbl := Table{
		Oid: ".1.0.0.0",
		Fields: []Field{
			{Oid: ".999", Name: "foo"},
			{Oid: "TEST::description", Name: "description", IsTag: true},
		},
	}
	err := tbl.Init(NewNetsnmpTranslator(testutil.Logger{}))
	require.NoError(t, err)

	require.Equal(t, "testTable", tbl.Name)

	require.Len(t, tbl.Fields, 5)

	require.Equal(t, ".999", tbl.Fields[0].Oid)
	require.Equal(t, "foo", tbl.Fields[0].Name)
	require.False(t, tbl.Fields[0].IsTag)
	require.Empty(t, tbl.Fields[0].Conversion)

	require.Equal(t, ".1.0.0.0.1.1", tbl.Fields[2].Oid)
	require.Equal(t, "server", tbl.Fields[2].Name)
	require.True(t, tbl.Fields[1].IsTag)
	require.Empty(t, tbl.Fields[1].Conversion)

	require.Equal(t, ".1.0.0.0.1.2", tbl.Fields[3].Oid)
	require.Equal(t, "connections", tbl.Fields[3].Name)
	require.False(t, tbl.Fields[3].IsTag)
	require.Empty(t, tbl.Fields[3].Conversion)

	require.Equal(t, ".1.0.0.0.1.3", tbl.Fields[4].Oid)
	require.Equal(t, "latency", tbl.Fields[4].Name)
	require.False(t, tbl.Fields[4].IsTag)
	require.Empty(t, tbl.Fields[4].Conversion)

	require.Equal(t, ".1.0.0.0.1.4", tbl.Fields[1].Oid)
	require.Equal(t, "description", tbl.Fields[1].Name)
	require.True(t, tbl.Fields[1].IsTag)
	require.Empty(t, tbl.Fields[1].Conversion)
}

func TestTableBuild_walk(t *testing.T) {
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
			{
				Name:      "myfield6",
				Oid:       ".1.0.0.0.1.6",
				Translate: true,
			},
			{
				Name:      "myfield7",
				Oid:       ".1.0.0.0.1.6",
				Translate: false,
			},
		},
	}

	require.NoError(t, tbl.Init(NewNetsnmpTranslator(testutil.Logger{})))

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
			"myfield6": "testTableEntry.7",
			"myfield7": ".1.0.0.0.1.7",
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

func TestTableBuild_noWalk(t *testing.T) {
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

func TestSnmpTranslateCache_miss(t *testing.T) {
	snmpTranslateCaches = nil
	oid := "IF-MIB::ifPhysAddress.1"
	mibName, oidNum, oidText, conversion, err := NewNetsnmpTranslator(testutil.Logger{}).SnmpTranslate(oid)
	require.Len(t, snmpTranslateCaches, 1)
	stc := snmpTranslateCaches[oid]
	require.NotNil(t, stc)
	require.Equal(t, mibName, stc.mibName)
	require.Equal(t, oidNum, stc.oidNum)
	require.Equal(t, oidText, stc.oidText)
	require.Equal(t, conversion, stc.conversion)
	require.Equal(t, err, stc.err)
}

func TestSnmpTranslateCache_hit(t *testing.T) {
	snmpTranslateCaches = map[string]snmpTranslateCache{
		"foo": {
			mibName:    "a",
			oidNum:     "b",
			oidText:    "c",
			conversion: "d",
		},
	}
	mibName, oidNum, oidText, conversion, err := NewNetsnmpTranslator(testutil.Logger{}).SnmpTranslate("foo")
	require.Equal(t, "a", mibName)
	require.Equal(t, "b", oidNum)
	require.Equal(t, "c", oidText)
	require.Equal(t, "d", conversion)
	require.NoError(t, err)
	snmpTranslateCaches = nil
}

func TestSnmpTableCache_miss(t *testing.T) {
	snmpTableCaches = nil
	oid := ".1.0.0.0"
	mibName, oidNum, oidText, fields, err := NewNetsnmpTranslator(testutil.Logger{}).SnmpTable(oid)
	require.Len(t, snmpTableCaches, 1)
	stc := snmpTableCaches[oid]
	require.NotNil(t, stc)
	require.Equal(t, mibName, stc.mibName)
	require.Equal(t, oidNum, stc.oidNum)
	require.Equal(t, oidText, stc.oidText)
	require.Equal(t, fields, stc.fields)
	require.Equal(t, err, stc.err)
}

func TestSnmpTableCache_hit(t *testing.T) {
	snmpTableCaches = map[string]snmpTableCache{
		"foo": {
			mibName: "a",
			oidNum:  "b",
			oidText: "c",
			fields:  []Field{{Name: "d"}},
		},
	}
	mibName, oidNum, oidText, fields, err := NewNetsnmpTranslator(testutil.Logger{}).SnmpTable("foo")
	require.Equal(t, "a", mibName)
	require.Equal(t, "b", oidNum)
	require.Equal(t, "c", oidText)
	require.Equal(t, []Field{{Name: "d"}}, fields)
	require.NoError(t, err)
}
