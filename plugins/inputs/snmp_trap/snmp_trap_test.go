package snmp_trap

//todo: look up smi

import (
	//"log"
	//"os"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/soniah/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func sendTrap(t *testing.T, port uint16) (sentTimestamp uint32) {
	s := &gosnmp.GoSNMP{
		Port:      port,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(2) * time.Second,
		Retries:   3,
		MaxOids:   gosnmp.MaxOids,
		Target:    "127.0.0.1",
		//Logger:    log.New(os.Stdout, "", 0),
	}

	err := s.Connect()
	if err != nil {
		t.Errorf("Connect() err: %v", err)
	}
	defer s.Conn.Close()

	//if the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	//prepend one with time.Now().  We need to check the time later on
	//so we have to do add it here.
	now := uint32(time.Now().Unix())
	timePdu := gosnmp.SnmpPDU{
		Name:  ".1.3.6.1.2.1.1.3.0",
		Type:  gosnmp.TimeTicks,
		Value: now,
	}

	pdu := gosnmp.SnmpPDU{
		Name:  ".1.3.6.1.6.3.1.1.4.1.0",
		Type:  gosnmp.ObjectIdentifier,
		Value: ".1.3.6.1.6.3.1.1.5.1",
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

// TestReceiveTrap
func TestReceiveTrap(t *testing.T) {
	const port = 12399 //todo: find unused port
	var fakeTime = time.Now()

	//hook into the trap handler so the test knows when the trap has been received
	received := make(chan int)
	wrap := func(f func(*gosnmp.SnmpPacket, *net.UDPAddr)) func(*gosnmp.SnmpPacket, *net.UDPAddr) {
		return func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
			f(p, a)
			received <- 0
		}
	}

	//set up the service input plugin
	n := &SnmpTrap{
		Port:               port,
		makeHandlerWrapper: wrap,
		timeFunc: func() time.Time {
			return fakeTime
		},
	}
	n.Init()
	var acc testutil.Accumulator
	n.Start(&acc)
	defer n.Stop()

	//wait until input plugin is listening
	select {
	case <-n.Listening():
	case err := <-n.Errch:
		t.Fatalf("error in listen: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting to listen")
	}

	//send the trap
	sentTimestamp := sendTrap(t, port)

	//wait for trap to be received
	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for trap to be received")
	}

	//validate plugin output
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"snmp_trap", //name
			map[string]string{ //tags
				".1.3.6.1.2.1.1.3.0":          fmt.Sprintf("%v", sentTimestamp),
				".1.3.6.1.2.1.1.3.0_type":     "67",
				".1.3.6.1.6.3.1.1.4.1.0":      ".1.3.6.1.6.3.1.1.5.1",
				".1.3.6.1.6.3.1.1.4.1.0_type": "6",
			},
			map[string]interface{}{ //fields
				"foo": "bar",
			},
			fakeTime,
		),
	}

	testutil.RequireMetricsEqual(t,
		expected, acc.GetTelegrafMetrics(),
		testutil.SortMetrics())

}
