package snmp_trap

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/soniah/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	defer snmp.SnmpTranslateClear()
	snmp.SnmpTranslateForce(
		".1.3.6.1.6.3.1.1.5.1",
		"SNMPv2-MIB",
		".1.3.6.1.6.3.1.1.5.1",
		"coldStart",
		"",
	)

	mibName, oidNum, oidText, conversion, err := snmp.SnmpTranslate(".1.3.6.1.6.3.1.1.5.1")
	require.NoError(t, err)
	require.Equal(t, "SNMPv2-MIB", mibName)
	require.Equal(t, ".1.3.6.1.6.3.1.1.5.1", oidNum)
	require.Equal(t, "coldStart", oidText)
	require.Equal(t, "", conversion)
}

func sendTrap(t *testing.T, port uint16) (sentTimestamp uint32) {
	s := &gosnmp.GoSNMP{
		Port:      port,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(2) * time.Second,
		Retries:   3,
		MaxOids:   gosnmp.MaxOids,
		Target:    "127.0.0.1",
	}

	err := s.Connect()
	if err != nil {
		t.Errorf("Connect() err: %v", err)
	}
	defer s.Conn.Close()

	// If the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	// prepend one with time.Now().  The time value is part of the
	// plugin output so we need to keep track of it and verify it
	// later.
	now := uint32(time.Now().Unix())
	timePdu := gosnmp.SnmpPDU{
		Name:  ".1.3.6.1.2.1.1.3.0",
		Type:  gosnmp.TimeTicks,
		Value: now,
	}

	pdu := gosnmp.SnmpPDU{
		Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
		Type:  gosnmp.ObjectIdentifier,
		Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
	}

	trap := gosnmp.SnmpTrap{
		Variables: []gosnmp.SnmpPDU{
			timePdu,
			pdu,
		},
	}

	_, err = s.SendTrap(trap)
	if err != nil {
		t.Errorf("SendTrap() err: %v", err)
	}

	return now
}

func TestReceiveTrap(t *testing.T) {
	// Preload the cache with the oids we'll use in this test so
	// snmptranslate and mibs don't need to be installed.
	defer snmp.SnmpTranslateClear()
	snmp.SnmpTranslateForce(
		".1.3.6.1.6.3.1.1.4.1.0",
		"SNMPv2-MIB",
		".1.3.6.1.6.3.1.1.4.1.0",
		"snmpTrapOID.0",
		"",
	)
	snmp.SnmpTranslateForce(
		".1.3.6.1.6.3.1.1.5.1",
		"SNMPv2-MIB",
		".1.3.6.1.6.3.1.1.5.1",
		"coldStart",
		"",
	)
	snmp.SnmpTranslateForce(
		".1.3.6.1.2.1.1.3.0",
		"UNUSED_MIB_NAME",
		".1.3.6.1.2.1.1.3.0",
		"sysUpTimeInstance",
		"",
	)

	// We would prefer to specify port 0 and let the network stack
	// choose an unused port for us but TrapListener doesn't have a
	// way to return the autoselected port.  Instead, we'll use an
	// unusual port and hope it's unused.
	const port = 12399
	var fakeTime = time.Now()

	// hook into the trap handler so the test knows when the trap has
	// been received
	received := make(chan int)
	wrap := func(f handler) handler {
		return func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
			f(p, a)
			received <- 0
		}
	}

	// set up the service input plugin
	n := &SnmpTrap{
		ServiceAddress:     "udp://:" + strconv.Itoa(port),
		makeHandlerWrapper: wrap,
		timeFunc: func() time.Time {
			return fakeTime
		},
		Log: testutil.Logger{},
	}
	require.Nil(t, n.Init())
	var acc testutil.Accumulator
	require.Nil(t, n.Start(&acc))
	defer n.Stop()

	// send the trap
	sentTimestamp := sendTrap(t, port)

	// wait for trap to be received
	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for trap to be received")
	}

	// verify plugin output
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"snmp_trap", // name
			map[string]string{ // tags
				"trap_oid":     ".1.3.6.1.6.3.1.1.5.1",
				"trap_name":    "coldStart",
				"trap_mib":     "SNMPv2-MIB",
				"trap_version": "2c",
				"source":       "127.0.0.1",
			},
			map[string]interface{}{ // fields
				"sysUpTimeInstance":      sentTimestamp,
				"sysUpTimeInstance_type": "TimeTicks",
			},
			fakeTime,
		),
	}

	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())

}
