package snmp_trap

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/soniah/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	s := &SnmpTrap{}
	require.Nil(t, s.Init())

	defer s.clear()
	s.load(
		".1.3.6.1.6.3.1.1.5.1",
		mibEntry{
			"SNMPv2-MIB",
			"coldStart",
		},
	)

	e, err := s.lookup(".1.3.6.1.6.3.1.1.5.1")
	require.NoError(t, err)
	require.Equal(t, "SNMPv2-MIB", e.mibName)
	require.Equal(t, "coldStart", e.oidText)
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
	s := &SnmpTrap{
		ServiceAddress:     "udp://:" + strconv.Itoa(port),
		makeHandlerWrapper: wrap,
		timeFunc: func() time.Time {
			return fakeTime
		},
		Log: testutil.Logger{},
	}
	require.Nil(t, s.Init())
	var acc testutil.Accumulator
	require.Nil(t, s.Start(&acc))
	defer s.Stop()

	// Preload the cache with the oids we'll use in this test so
	// snmptranslate and mibs don't need to be installed.
	defer s.clear()
	s.load(".1.3.6.1.6.3.1.1.4.1.0",
		mibEntry{
			"SNMPv2-MIB",
			"snmpTrapOID.0",
		})
	s.load(".1.3.6.1.6.3.1.1.5.1",
		mibEntry{
			"SNMPv2-MIB",
			"coldStart",
		})
	s.load(".1.3.6.1.2.1.1.3.0",
		mibEntry{
			"UNUSED_MIB_NAME",
			"sysUpTimeInstance",
		})

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
				"oid":     ".1.3.6.1.6.3.1.1.5.1",
				"name":    "coldStart",
				"mib":     "SNMPv2-MIB",
				"version": "2c",
				"source":  "127.0.0.1",
			},
			map[string]interface{}{ // fields
				"sysUpTimeInstance": sentTimestamp,
			},
			fakeTime,
		),
	}

	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())

}

func fakeExecCmd(_ internal.Duration, _ string, _ ...string) ([]byte, error) {
	return nil, fmt.Errorf("intentional failure")
}

func TestMissingOid(t *testing.T) {
	// should fail even if snmptranslate is installed
	const port = 12399
	var fakeTime = time.Now()

	received := make(chan int)
	wrap := func(f handler) handler {
		return func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
			f(p, a)
			received <- 0
		}
	}

	s := &SnmpTrap{
		ServiceAddress:     "udp://:" + strconv.Itoa(port),
		makeHandlerWrapper: wrap,
		timeFunc: func() time.Time {
			return fakeTime
		},
		Log: testutil.Logger{},
	}
	require.Nil(t, s.Init())
	var acc testutil.Accumulator
	require.Nil(t, s.Start(&acc))
	defer s.Stop()

	// make sure the cache is empty
	s.clear()

	// don't call the real snmptranslate
	s.execCmd = fakeExecCmd

	_ = sendTrap(t, port)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for trap to be received")
	}

	// oid lookup should fail so we shouldn't get a metric
	expected := []telegraf.Metric{}

	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())
}
