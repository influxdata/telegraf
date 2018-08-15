package bond

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var sampleTest802 = `
Ethernet Channel Bonding Driver: v3.5.0 (November 4, 2008)

Bonding Mode: IEEE 802.3ad Dynamic link aggregation
Transmit Hash Policy: layer2 (0)
MII Status: up
MII Polling Interval (ms): 100
Up Delay (ms): 0
Down Delay (ms): 0

802.3ad info
LACP rate: fast
Aggregator selection policy (ad_select): stable
bond bond0 has no active aggregator

Slave Interface: eth1
MII Status: up
Link Failure Count: 0
Permanent HW addr: 00:0c:29:f5:b7:11
Aggregator ID: N/A

Slave Interface: eth2
MII Status: up
Link Failure Count: 3
Permanent HW addr: 00:0c:29:f5:b7:1b
Aggregator ID: N/A
`

var sampleTestAB = `
Ethernet Channel Bonding Driver: v3.6.0 (September 26, 2009)

Bonding Mode: fault-tolerance (active-backup)
Primary Slave: eth2 (primary_reselect always)
Currently Active Slave: eth2
MII Status: up
MII Polling Interval (ms): 100
Up Delay (ms): 0
Down Delay (ms): 0

Slave Interface: eth3
MII Status: down
Speed: 1000 Mbps
Duplex: full
Link Failure Count: 2
Permanent HW addr:
Slave queue ID: 0

Slave Interface: eth2
MII Status: up
Speed: 100 Mbps
Duplex: full
Link Failure Count: 0
Permanent HW addr:
`

func TestGatherBondInterface(t *testing.T) {
	var acc testutil.Accumulator
	bond := &Bond{}

	bond.gatherBondInterface("bond802", sampleTest802, &acc)
	acc.AssertContainsTaggedFields(t, "bond", map[string]interface{}{"status": 1}, map[string]string{"bond": "bond802"})
	acc.AssertContainsTaggedFields(t, "bond_slave", map[string]interface{}{"failures": 0, "status": 1}, map[string]string{"bond": "bond802", "interface": "eth1"})
	acc.AssertContainsTaggedFields(t, "bond_slave", map[string]interface{}{"failures": 3, "status": 1}, map[string]string{"bond": "bond802", "interface": "eth2"})

	bond.gatherBondInterface("bondAB", sampleTestAB, &acc)
	acc.AssertContainsTaggedFields(t, "bond", map[string]interface{}{"active_slave": "eth2", "status": 1}, map[string]string{"bond": "bondAB"})
	acc.AssertContainsTaggedFields(t, "bond_slave", map[string]interface{}{"failures": 2, "status": 0}, map[string]string{"bond": "bondAB", "interface": "eth3"})
	acc.AssertContainsTaggedFields(t, "bond_slave", map[string]interface{}{"failures": 0, "status": 1}, map[string]string{"bond": "bondAB", "interface": "eth2"})
}
