package snmp_trap

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestReceiveTrapV1(t *testing.T) {
	now := uint32(123123123)

	// If the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	// prepend one with time.Now()
	var tests = []struct {
		name string

		// send
		trap gosnmp.SnmpTrap // include pdus

		// receive
		entries  []entry
		expected []telegraf.Metric
	}{
		{
			name: "trap enterprise",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.2.3.4.5",
						Type:  gosnmp.OctetString,
						Value: "payload",
					},
				},
				Enterprise:   ".1.2.3",
				AgentAddress: "10.20.30.40",
				GenericTrap:  6, // enterpriseSpecific
				SpecificTrap: 55,
				Timestamp:    uint(now),
			},
			entries: []entry{
				{
					".1.2.3.4.5",
					snmp.MibEntry{
						MibName: "valueMIB",
						OidText: "valueOID",
					},
				},
				{
					".1.2.3.0.55",
					snmp.MibEntry{
						MibName: "enterpriseMIB",
						OidText: "enterpriseOID",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap",
					map[string]string{
						"oid":           ".1.2.3.0.55",
						"name":          "enterpriseOID",
						"mib":           "enterpriseMIB",
						"version":       "1",
						"source":        "127.0.0.1",
						"agent_address": "10.20.30.40",
						"community":     "public",
					},
					map[string]interface{}{
						"sysUpTimeInstance": uint(now),
						"valueOID":          "payload",
					},
					time.Unix(0, 0),
				),
			},
		},
		// v1 generic trap
		{
			name: "trap generic",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.2.3.4.5",
						Type:  gosnmp.OctetString,
						Value: "payload",
					},
					{
						Name:  ".1.2.3.4.6",
						Type:  gosnmp.OctetString,
						Value: []byte{0x7, 0xe8, 0x1, 0x4, 0xe, 0x2, 0x19, 0x0, 0x0, 0xe, 0x2},
					},
				},
				Enterprise:   ".1.2.3",
				AgentAddress: "10.20.30.40",
				GenericTrap:  0, // coldStart
				SpecificTrap: 0,
				Timestamp:    uint(now),
			},
			entries: []entry{
				{
					".1.2.3.4.5",
					snmp.MibEntry{
						MibName: "valueMIB",
						OidText: "valueOID",
					},
				},
				{
					".1.2.3.4.6",
					snmp.MibEntry{
						MibName: "valueMIB",
						OidText: "valueHexOID",
					},
				},
				{
					".1.3.6.1.6.3.1.1.5.1",
					snmp.MibEntry{
						MibName: "coldStartMIB",
						OidText: "coldStartOID",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap",
					map[string]string{
						"oid":           ".1.3.6.1.6.3.1.1.5.1",
						"name":          "coldStartOID",
						"mib":           "coldStartMIB",
						"version":       "1",
						"source":        "127.0.0.1",
						"agent_address": "10.20.30.40",
						"community":     "public",
					},
					map[string]interface{}{
						"sysUpTimeInstance": uint(now),
						"valueOID":          "payload",
						"valueHexOID":       "07e801040e021900000e02",
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We would prefer to specify port 0 and let the network
			// stack choose an unused port for us but TrapListener
			// doesn't have a way to return the autoselected port.
			// Instead, we'll use an unusual port and hope it's
			// unused.
			const port = 12399

			// Set up the service input plugin
			plugin := &SnmpTrap{
				ServiceAddress: "udp://:" + strconv.Itoa(port),
				Version:        "1",
				Translator:     "netsnmp",
				Log:            testutil.Logger{},
				timeFunc:       time.Now,
			}
			require.NoError(t, plugin.Init())

			// inject test translator
			plugin.transl = &testTranslator{entries: tt.entries}

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Create a v1 client and send the trap
			client := &gosnmp.GoSNMP{
				Port:      port,
				Version:   gosnmp.Version1,
				Timeout:   2 * time.Second,
				Retries:   1,
				MaxOids:   gosnmp.MaxOids,
				Target:    "127.0.0.1",
				Community: "public",
			}
			require.NoError(t, client.Connect(), "connecting failed")
			defer client.Conn.Close()
			_, err := client.SendTrap(tt.trap)
			require.NoError(t, err, "sending failed")
			require.NoError(t, client.Conn.Close(), "closing failed")

			// Wait for trap to be received
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond, "timed out waiting for trap to be received")

			// Verify plugin output
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestReceiveTrapV2c(t *testing.T) {
	now := uint32(123123123)

	// If the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	// prepend one with time.Now()
	var tests = []struct {
		name string

		// send
		trap gosnmp.SnmpTrap // include pdus

		// receive
		entries  []entry
		expected []telegraf.Metric
	}{
		// ordinary v2c coldStart trap
		{
			name: "v2c coldStart",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":       ".1.3.6.1.6.3.1.1.5.1",
						"name":      "coldStart",
						"mib":       "SNMPv2-MIB",
						"version":   "2c",
						"source":    "127.0.0.1",
						"community": "public",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We would prefer to specify port 0 and let the network
			// stack choose an unused port for us but TrapListener
			// doesn't have a way to return the autoselected port.
			// Instead, we'll use an unusual port and hope it's
			// unused.
			const port = 12399

			// Set up the service input plugin
			plugin := &SnmpTrap{
				ServiceAddress: "udp://:" + strconv.Itoa(port),
				Version:        "2c",
				Translator:     "netsnmp",
				Log:            testutil.Logger{},
				timeFunc:       time.Now,
			}
			require.NoError(t, plugin.Init())

			// inject test translator
			plugin.transl = &testTranslator{entries: tt.entries}

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Create a v1 client and send the trap
			client := &gosnmp.GoSNMP{
				Port:      port,
				Version:   gosnmp.Version2c,
				Timeout:   2 * time.Second,
				Retries:   1,
				MaxOids:   gosnmp.MaxOids,
				Target:    "127.0.0.1",
				Community: "public",
			}
			require.NoError(t, client.Connect(), "connecting failed")
			defer client.Conn.Close()
			_, err := client.SendTrap(tt.trap)
			require.NoError(t, err, "sending failed")
			require.NoError(t, client.Conn.Close(), "closing failed")

			// Wait for trap to be received
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond, "timed out waiting for trap to be received")

			// Verify plugin output
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestReceiveTrapV3(t *testing.T) {
	now := uint32(123123123)

	// If the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	// prepend one with time.Now()
	var tests = []struct {
		name string

		// send
		trap gosnmp.SnmpTrap // include pdus

		// auth and priv parameters
		secName   string // v3 username
		secLevel  string // v3 security level
		authProto string // Auth protocol: "", MD5 or SHA
		authPass  string // Auth passphrase
		privProto string // Priv protocol: "", DES or AES
		privPass  string // Priv passphrase

		// sender context
		contextName string
		engineID    string

		// receive
		entries  []entry
		expected []telegraf.Metric
	}{
		// ordinary v3 coldStart trap no auth and no priv
		{
			name:        "coldStart noAuthNoPriv",
			secName:     "noAuthNoPriv",
			secLevel:    "noAuthNoPriv",
			contextName: "foo_context_name",
			engineID:    "bar_engine_id",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":          ".1.3.6.1.6.3.1.1.5.1",
						"name":         "coldStart",
						"mib":          "SNMPv2-MIB",
						"version":      "3",
						"source":       "127.0.0.1",
						"context_name": "foo_context_name",
						"engine_id":    "6261725f656e67696e655f6964",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA auth and no priv
		{
			name:      "coldStart authShaNoPriv",
			secName:   "authShaNoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA224 auth and no priv
		{
			name:      "coldStart authShaNoPriv",
			secName:   "authSha224NoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA224",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA256 auth and no priv
		{
			name:      "coldStart authSha256NoPriv",
			secName:   "authSha256NoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA256",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA384 auth and no priv
		{
			name:      "coldStart authSha384NoPriv",
			secName:   "authSha384NoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA384",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA512 auth and no priv
		{
			name:      "coldStart authShaNoPriv",
			secName:   "authSha512NoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA512",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap SHA auth and no priv
		{
			name:      "coldStart authShaNoPriv",
			secName:   "authShaNoPriv",
			secLevel:  "authNoPriv",
			authProto: "SHA",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldstart trap MD5 auth and no priv
		{
			name:      "coldStart authMD5NoPriv",
			secName:   "authMD5NoPriv",
			secLevel:  "authNoPriv",
			authProto: "MD5",
			authPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and AES priv
		{
			name:      "coldStart authSHAPrivAES",
			secName:   "authSHAPrivAES",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "AES",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and DES priv
		{
			name:      "coldStart authSHAPrivDES",
			secName:   "authSHAPrivDES",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "DES",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and AES192 priv
		{
			name:      "coldStart authSHAPrivAES192",
			secName:   "authSHAPrivAES192",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "AES192",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and AES192C priv
		{
			name:      "coldStart authSHAPrivAES192C",
			secName:   "authSHAPrivAES192C",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "AES192C",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and AES256 priv
		{
			name:      "coldStart authSHAPrivAES256",
			secName:   "authSHAPrivAES256",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "AES256",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
		// ordinary v3 coldStart SHA trap auth and AES256C priv
		{
			name:      "coldStart authSHAPrivAES256C",
			secName:   "authSHAPrivAES256C",
			secLevel:  "authPriv",
			authProto: "SHA",
			authPass:  "passpass",
			privProto: "AES256C",
			privPass:  "passpass",
			trap: gosnmp.SnmpTrap{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  ".1.3.6.1.2.1.1.3.0",
						Type:  gosnmp.TimeTicks,
						Value: now,
					},
					{
						Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
						Type:  gosnmp.ObjectIdentifier,
						Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
					},
				},
			},
			entries: []entry{
				{
					oid: ".1.3.6.1.6.3.1.1.4.1.0",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: snmp.MibEntry{
						MibName: "SNMPv2-MIB",
						OidText: "coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: snmp.MibEntry{
						MibName: "UNUSED_MIB_NAME",
						OidText: "sysUpTimeInstance",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":     ".1.3.6.1.6.3.1.1.5.1",
						"name":    "coldStart",
						"mib":     "SNMPv2-MIB",
						"version": "3",
						"source":  "127.0.0.1",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": now,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We would prefer to specify port 0 and let the network
			// stack choose an unused port for us but TrapListener
			// doesn't have a way to return the autoselected port.
			// Instead, we'll use an unusual port and hope it's
			// unused.
			const port = 12399

			// Set up the service input plugin
			plugin := &SnmpTrap{
				ServiceAddress: "udp://:" + strconv.Itoa(port),
				Version:        "3",
				Translator:     "netsnmp",
				SecName:        config.NewSecret([]byte(tt.secName)),
				SecLevel:       tt.secLevel,
				AuthProtocol:   tt.authProto,
				AuthPassword:   config.NewSecret([]byte(tt.authPass)),
				PrivProtocol:   tt.privProto,
				PrivPassword:   config.NewSecret([]byte(tt.privPass)),
				Log:            testutil.Logger{},
				timeFunc:       time.Now,
			}
			require.NoError(t, plugin.Init())

			// inject test translator
			plugin.transl = &testTranslator{entries: tt.entries}

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Create a v3 client and send the trap
			var msgFlags gosnmp.SnmpV3MsgFlags
			switch strings.ToLower(tt.secLevel) {
			case "noauthnopriv", "":
				msgFlags = gosnmp.NoAuthNoPriv
			case "authnopriv":
				msgFlags = gosnmp.AuthNoPriv
			case "authpriv":
				msgFlags = gosnmp.AuthPriv
			default:
				require.FailNowf(t, "unknown security level %q", tt.secLevel)
			}
			security := createSecurityParameters(tt.authProto, tt.privProto, tt.secName, tt.privPass, tt.authPass)

			client := &gosnmp.GoSNMP{
				Port:               port,
				Version:            gosnmp.Version3,
				Timeout:            2 * time.Second,
				Retries:            1,
				MaxOids:            gosnmp.MaxOids,
				Target:             "127.0.0.1",
				SecurityParameters: security,
				SecurityModel:      gosnmp.UserSecurityModel,
				MsgFlags:           msgFlags,
				ContextName:        tt.contextName,
				ContextEngineID:    tt.engineID,
			}
			require.NoError(t, client.Connect(), "connecting failed")
			defer client.Conn.Close()
			_, err := client.SendTrap(tt.trap)
			require.NoError(t, err, "sending failed")
			require.NoError(t, client.Conn.Close(), "closing failed")

			// Wait for trap to be received
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond, "timed out waiting for trap to be received")

			// Verify plugin output
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestOidLookupFail(t *testing.T) {
	now := uint32(123123123)

	// Check that we're not running snmptranslate to look up oids
	// when we shouldn't. This sends and receives a valid trap
	// but metric production should fail because the oids aren't in
	// the cache and oid lookup is intentionally mocked to fail.
	trap := gosnmp.SnmpTrap{
		Variables: []gosnmp.SnmpPDU{
			{
				Name:  ".1.3.6.1.2.1.1.3.0",
				Type:  gosnmp.TimeTicks,
				Value: now,
			},
			{
				Name:  ".1.3.6.1.6.3.1.1.4.1.0", // SNMPv2-MIB::snmpTrapOID.0
				Type:  gosnmp.ObjectIdentifier,
				Value: ".1.3.6.1.6.3.1.1.5.1", // coldStart
			},
		},
	}

	// We would prefer to specify port 0 and let the network
	// stack choose an unused port for us but TrapListener
	// doesn't have a way to return the autoselected port.
	// Instead, we'll use an unusual port and hope it's
	// unused.
	const port = 12399

	// Set up the service input plugin
	logger := &testutil.CaptureLogger{}
	fail := make(chan bool, 1)
	plugin := &SnmpTrap{
		ServiceAddress: "udp://:" + strconv.Itoa(port),
		Version:        "2c",
		Translator:     "netsnmp",
		Log:            logger,
		timeFunc:       time.Now,
	}
	require.NoError(t, plugin.Init())

	// inject test translator
	plugin.transl = &testTranslator{fail: fail}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Create a v1 client and send the trap
	client := &gosnmp.GoSNMP{
		Port:      port,
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   1,
		MaxOids:   gosnmp.MaxOids,
		Target:    "127.0.0.1",
		Community: "public",
	}
	require.NoError(t, client.Connect(), "connecting failed")
	defer client.Conn.Close()
	_, err := client.SendTrap(trap)
	require.NoError(t, err, "sending failed")
	require.NoError(t, client.Conn.Close(), "closing failed")

	// Wait for lookup to fail
	select {
	case <-fail:
	case <-time.After(time.Second):
		t.Log("timeout waiting for failing OID lookup")
		t.Fail()
	}

	// Verify plugin output
	require.Empty(t, acc.GetTelegrafMetrics())
	var found bool
	for _, msg := range logger.Errors() {
		if found = strings.Contains(msg, "unexpected oid"); found {
			break
		}
	}
	require.True(t, found, "did not receive expected error message")
}

type entry struct {
	oid string
	e   snmp.MibEntry
}

type testTranslator struct {
	entries []entry
	fail    chan bool
}

func (t *testTranslator) lookup(input string) (snmp.MibEntry, error) {
	for _, entry := range t.entries {
		if input == entry.oid {
			return snmp.MibEntry{MibName: entry.e.MibName, OidText: entry.e.OidText}, nil
		}
	}
	if t.fail != nil {
		t.fail <- true
	}
	return snmp.MibEntry{}, errors.New("unexpected oid")
}

func createSecurityParameters(authProto, privProto, username, privPass, authPass string) *gosnmp.UsmSecurityParameters {
	var authenticationProtocol gosnmp.SnmpV3AuthProtocol
	switch strings.ToLower(authProto) {
	case "md5":
		authenticationProtocol = gosnmp.MD5
	case "sha":
		authenticationProtocol = gosnmp.SHA
	case "sha224":
		authenticationProtocol = gosnmp.SHA224
	case "sha256":
		authenticationProtocol = gosnmp.SHA256
	case "sha384":
		authenticationProtocol = gosnmp.SHA384
	case "sha512":
		authenticationProtocol = gosnmp.SHA512
	case "":
		authenticationProtocol = gosnmp.NoAuth
	default:
		authenticationProtocol = gosnmp.NoAuth
	}

	var privacyProtocol gosnmp.SnmpV3PrivProtocol
	switch strings.ToLower(privProto) {
	case "aes":
		privacyProtocol = gosnmp.AES
	case "des":
		privacyProtocol = gosnmp.DES
	case "aes192":
		privacyProtocol = gosnmp.AES192
	case "aes192c":
		privacyProtocol = gosnmp.AES192C
	case "aes256":
		privacyProtocol = gosnmp.AES256
	case "aes256c":
		privacyProtocol = gosnmp.AES256C
	case "":
		privacyProtocol = gosnmp.NoPriv
	default:
		privacyProtocol = gosnmp.NoPriv
	}

	return &gosnmp.UsmSecurityParameters{
		AuthoritativeEngineID:    "deadbeef", // has to be between 5 & 32 chars
		AuthoritativeEngineBoots: 1,
		AuthoritativeEngineTime:  1,
		UserName:                 username,
		PrivacyProtocol:          privacyProtocol,
		PrivacyPassphrase:        privPass,
		AuthenticationPassphrase: authPass,
		AuthenticationProtocol:   authenticationProtocol,
	}
}
