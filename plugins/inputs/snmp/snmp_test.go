package snmp

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSNMPErrorGet1(t *testing.T) {
	get1 := Data{
		Name: "oid1",
		Unit: "octets",
		Oid:  ".1.3.6.1.2.1.2.2.1.16.1",
	}
	h := Host{
		Collect: []string{"oid1"},
	}
	s := Snmp{
		SnmptranslateFile: "bad_oid.txt",
		Host:              []Host{h},
		Get:               []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.Error(t, err)
}

func TestSNMPErrorGet2(t *testing.T) {
	get1 := Data{
		Name: "oid1",
		Unit: "octets",
		Oid:  ".1.3.6.1.2.1.2.2.1.16.1",
	}
	h := Host{
		Collect: []string{"oid1"},
	}
	s := Snmp{
		Host: []Host{h},
		Get:  []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)
	assert.Equal(t, 0, len(acc.Metrics))
}

func TestSNMPErrorBulk(t *testing.T) {
	bulk1 := Data{
		Name: "oid1",
		Unit: "octets",
		Oid:  ".1.3.6.1.2.1.2.2.1.16",
	}
	h := Host{
		Address: testutil.GetLocalHost(),
		Collect: []string{"oid1"},
	}
	s := Snmp{
		Host: []Host{h},
		Bulk: []Data{bulk1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)
	assert.Equal(t, 0, len(acc.Metrics))
}

func TestSNMPGet1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	get1 := Data{
		Name: "oid1",
		Unit: "octets",
		Oid:  ".1.3.6.1.2.1.2.2.1.16.1",
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		Host: []Host{h},
		Get:  []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"oid1",
		map[string]interface{}{
			"oid1": uint(543846),
		},
		map[string]string{
			"unit":      "octets",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPGet2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	get1 := Data{
		Name: "oid1",
		Oid:  "ifNumber",
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Get:               []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifNumber",
		map[string]interface{}{
			"ifNumber": int(4),
		},
		map[string]string{
			"instance":  "0",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPGet3(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	get1 := Data{
		Name:     "oid1",
		Unit:     "octets",
		Oid:      "ifSpeed",
		Instance: "1",
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Get:               []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifSpeed",
		map[string]interface{}{
			"ifSpeed": uint(10000000),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "1",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPEasyGet4(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	get1 := Data{
		Name:     "oid1",
		Unit:     "octets",
		Oid:      "ifSpeed",
		Instance: "1",
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
		GetOids:   []string{"ifNumber"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Get:               []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifSpeed",
		map[string]interface{}{
			"ifSpeed": uint(10000000),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "1",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifNumber",
		map[string]interface{}{
			"ifNumber": int(4),
		},
		map[string]string{
			"instance":  "0",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPEasyGet5(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	get1 := Data{
		Name:     "oid1",
		Unit:     "octets",
		Oid:      "ifSpeed",
		Instance: "1",
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
		GetOids:   []string{".1.3.6.1.2.1.2.1.0"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Get:               []Data{get1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifSpeed",
		map[string]interface{}{
			"ifSpeed": uint(10000000),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "1",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifNumber",
		map[string]interface{}{
			"ifNumber": int(4),
		},
		map[string]string{
			"instance":  "0",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPEasyGet6(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		GetOids:   []string{"1.3.6.1.2.1.2.1.0"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifNumber",
		map[string]interface{}{
			"ifNumber": int(4),
		},
		map[string]string{
			"instance":  "0",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

func TestSNMPBulk1(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	bulk1 := Data{
		Name:          "oid1",
		Unit:          "octets",
		Oid:           ".1.3.6.1.2.1.2.2.1.16",
		MaxRepetition: 2,
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Bulk:              []Data{bulk1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(543846),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "1",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(26475179),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "2",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(108963968),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "3",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(12991453),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "36",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}

// TODO find why, if this test is active
// Circle CI stops with the following error...
// bash scripts/circle-test.sh died unexpectedly
// Maybe the test is too long ??
func dTestSNMPBulk2(t *testing.T) {
	bulk1 := Data{
		Name:          "oid1",
		Unit:          "octets",
		Oid:           "ifOutOctets",
		MaxRepetition: 2,
	}
	h := Host{
		Address:   testutil.GetLocalHost() + ":31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		SnmptranslateFile: "./testdata/oids.txt",
		Host:              []Host{h},
		Bulk:              []Data{bulk1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(543846),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "1",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(26475179),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "2",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(108963968),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "3",
			"snmp_host": testutil.GetLocalHost(),
		},
	)

	acc.AssertContainsTaggedFields(t,
		"ifOutOctets",
		map[string]interface{}{
			"ifOutOctets": uint(12991453),
		},
		map[string]string{
			"unit":      "octets",
			"instance":  "36",
			"snmp_host": testutil.GetLocalHost(),
		},
	)
}
