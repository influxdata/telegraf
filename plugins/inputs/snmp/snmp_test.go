package snmp

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSNMPErrorGet(t *testing.T) {
	get1 := Data{
		Name: "oid1",
		Unit: "second",
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
	require.Error(t, err)
}

func TestSNMPErrorBulk(t *testing.T) {
	bulk1 := Data{
		Name:          "oid1",
		Unit:          "second",
		Oid:           ".1.3.6.1.2.1.2.2.1.16",
		Snmptranslate: true,
	}
	h := Host{
		Collect: []string{"oid1"},
	}
	s := Snmp{
		Host: []Host{h},
		Bulk: []Data{bulk1},
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.Error(t, err)
}

func TestSNMPBulk(t *testing.T) {
	bulk1 := Data{
		Name:          "oid1",
		Unit:          "second",
		Oid:           ".1.3.6.1.2.1.2.2.1.16",
		MaxRepetition: 2,
		Snmptranslate: true,
	}
	h := Host{
		Address:   "127.0.0.1:31161",
		Community: "telegraf",
		Version:   2,
		Timeout:   2.0,
		Retries:   2,
		Collect:   []string{"oid1"},
	}
	s := Snmp{
		Host: []Host{h},
		Bulk: []Data{bulk1},
	}

	expected := map[string]uint{".1.3.6.1.2.1.2.2.1.16.1": 543846,
		".1.3.6.1.2.1.2.2.1.16.2":  26475179,
		".1.3.6.1.2.1.2.2.1.16.3":  108963968,
		".1.3.6.1.2.1.2.2.1.16.36": 12991453,
	}

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	for _, p := range acc.Points {
		assert.Equal(t, expected[p.Measurement], p.Fields[p.Measurement])
	}
}

func TestSNMPGet(t *testing.T) {
	get1 := Data{
		Name: "oid1",
		Unit: "second",
		Oid:  ".1.3.6.1.2.1.2.2.1.16.1",
	}
	h := Host{
		Address:   "127.0.0.1:31161",
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

	for _, p := range acc.Points {
		value := p.Fields["oid1"].(uint)
		assert.Equal(t, get1.Name, p.Measurement)
		assert.Equal(t, uint(543846), value)
	}
}
