package snmp_trap

// todo: tests that look up oids will pass only if snmptranslate (part
// of net-snmp) is installed and working.  We need to mock name lookup
// or add a way to disable it so tests will pass when snmptranslate
// isn't available.

import (
	// "log"
	// "os"
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
		// Logger:    log.New(os.Stdout, "", 0),
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

//  TestReceiveTrap
func TestReceiveTrap(t *testing.T) {
	const port = 12399 // todo: find unused port
	var fakeTime = time.Now()

	// hook into the trap handler so the test knows when the trap has been received
	received := make(chan int)
	wrap := func(f func(*gosnmp.SnmpPacket, *net.UDPAddr)) func(*gosnmp.SnmpPacket, *net.UDPAddr) {
		return func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
			f(p, a)
			received <- 0
		}
	}

	// set up the service input plugin
	n := &SnmpTrap{
		ServiceAddress:     "localhost:" + strconv.Itoa(port),
		makeHandlerWrapper: wrap,
		timeFunc: func() time.Time {
			return fakeTime
		},
	}
	n.Init()
	var acc testutil.Accumulator
	n.Start(&acc)
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
