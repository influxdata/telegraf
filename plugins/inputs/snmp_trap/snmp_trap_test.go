package snmp_trap

import (
	"fmt"
	"net"
	"strconv"
	"strings"
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

func fakeExecCmd(_ internal.Duration, x string, y ...string) ([]byte, error) {
	return nil, fmt.Errorf("mock " + x + " " + strings.Join(y, " "))
}

func sendTrap(t *testing.T, port uint16, now uint32, trap gosnmp.SnmpTrap, version gosnmp.SnmpVersion, seclevel string, username string, authproto string, authpass string, privproto string, privpass string) {
	s := &gosnmp.GoSNMP{
		Port:        port,
		Community:   "public",
		Version:     version,
		Timeout:     time.Duration(2) * time.Second,
		Retries:     1,
		MaxOids:     gosnmp.MaxOids,
		Target:      "127.0.0.1",
		ContextName: "public",
	}
	if s.Version == gosnmp.Version3 {
		sp := &gosnmp.UsmSecurityParameters{}
		s.SecurityParameters = sp
		s.SecurityModel = gosnmp.UserSecurityModel

		// TODO: Find out why sending traps with auth and/or priv yields invalied packets
		switch strings.ToLower(seclevel) {
		case "noauthnopriv", "":
			s.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			s.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			s.MsgFlags = gosnmp.AuthPriv
		default:
			s.MsgFlags = gosnmp.NoAuthNoPriv
		}

		switch strings.ToLower(authproto) {
		case "md5":
			sp.AuthenticationProtocol = gosnmp.MD5
		case "sha":
			sp.AuthenticationProtocol = gosnmp.SHA
		case "":
			sp.AuthenticationProtocol = gosnmp.NoAuth
		default:
			sp.AuthenticationProtocol = gosnmp.NoAuth
		}

		switch strings.ToLower(privproto) {
		case "aes":
			sp.PrivacyProtocol = gosnmp.AES
		case "des":
			sp.PrivacyProtocol = gosnmp.DES
		case "":
			sp.PrivacyProtocol = gosnmp.NoPriv
		default:
			sp.PrivacyProtocol = gosnmp.NoPriv
		}

		if privpass != "" {
			sp.PrivacyPassphrase = privpass
		}
		if authpass != "" {
			sp.AuthenticationPassphrase = authpass
		}

		sp.UserName = username
		//sp.AuthoritativeEngineID = "80001f8880031dd407f608905e00000000"
		//sp.AuthoritativeEngineBoots = 1
		//sp.AuthoritativeEngineTime = 1

	}

	err := s.Connect()
	if err != nil {
		t.Errorf("Connect() err: %v", err)
	}
	defer s.Conn.Close()

	_, err = s.SendTrap(trap)
	if err != nil {
		t.Errorf("SendTrap() err: %v", err)
	}
}

func TestReceiveTrap(t *testing.T) {
	var now uint32
	now = 123123123

	var fakeTime time.Time
	fakeTime = time.Unix(456456456, 456)

	type entry struct {
		oid string
		e   mibEntry
	}

	// If the first pdu isn't type TimeTicks, gosnmp.SendTrap() will
	// prepend one with time.Now()
	var tests = []struct {
		name string

		// send
		version gosnmp.SnmpVersion
		trap    gosnmp.SnmpTrap // include pdus
		// V3 auth and priv parameters
		secname   string // v3 username
		seclevel  string // v3 security level
		authproto string // Auth protocol: "", MD5 or SHA
		authpass  string // Auth passphrase
		privproto string // Priv protocol: "", DES or AES
		privpass  string // Priv passphrase

		// recieve
		entries []entry
		metrics []telegraf.Metric
	}{
		//ordinary v2c coldStart trap
		{
			name:    "v2c coldStart",
			version: gosnmp.Version2c,
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
					e: mibEntry{
						"SNMPv2-MIB",
						"snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: mibEntry{
						"SNMPv2-MIB",
						"coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: mibEntry{
						"UNUSED_MIB_NAME",
						"sysUpTimeInstance",
					},
				},
			},
			metrics: []telegraf.Metric{
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
						"sysUpTimeInstance": now,
					},
					fakeTime,
				),
			},
		},
		//Check that we're not running snmptranslate to look up oids
		//when we shouldn't be.  This sends and receives a valid trap
		//but metric production should fail because the oids aren't in
		//the cache and oid lookup is intentionally mocked to fail.
		{
			name:    "missing oid",
			version: gosnmp.Version2c,
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
			entries: []entry{}, //nothing in cache
			metrics: []telegraf.Metric{},
		},
		//v1 enterprise specific trap
		{
			name:    "v1 trap enterprise",
			version: gosnmp.Version1,
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
					mibEntry{
						"valueMIB",
						"valueOID",
					},
				},
				{
					".1.2.3.0.55",
					mibEntry{
						"enterpriseMIB",
						"enterpriseOID",
					},
				},
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":           ".1.2.3.0.55",
						"name":          "enterpriseOID",
						"mib":           "enterpriseMIB",
						"version":       "1",
						"source":        "127.0.0.1",
						"agent_address": "10.20.30.40",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": uint(now),
						"valueOID":          "payload",
					},
					fakeTime,
				),
			},
		},
		//v1 generic trap
		{
			name:    "v1 trap generic",
			version: gosnmp.Version1,
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
				GenericTrap:  0, //coldStart
				SpecificTrap: 0,
				Timestamp:    uint(now),
			},
			entries: []entry{
				{
					".1.2.3.4.5",
					mibEntry{
						"valueMIB",
						"valueOID",
					},
				},
				{
					".1.3.6.1.6.3.1.1.5.1",
					mibEntry{
						"coldStartMIB",
						"coldStartOID",
					},
				},
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"snmp_trap", // name
					map[string]string{ // tags
						"oid":           ".1.3.6.1.6.3.1.1.5.1",
						"name":          "coldStartOID",
						"mib":           "coldStartMIB",
						"version":       "1",
						"source":        "127.0.0.1",
						"agent_address": "10.20.30.40",
					},
					map[string]interface{}{ // fields
						"sysUpTimeInstance": uint(now),
						"valueOID":          "payload",
					},
					fakeTime,
				),
			},
		},
		//ordinary v3 coldStart trap no auth and no priv
		{
			name:     "v3 coldStart noAuthNoPriv",
			version:  gosnmp.Version3,
			secname:  "test-user",
			seclevel: "noAuthNoPriv",
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
					e: mibEntry{
						"SNMPv2-MIB",
						"snmpTrapOID.0",
					},
				},
				{
					oid: ".1.3.6.1.6.3.1.1.5.1",
					e: mibEntry{
						"SNMPv2-MIB",
						"coldStart",
					},
				},
				{
					oid: ".1.3.6.1.2.1.1.3.0",
					e: mibEntry{
						"UNUSED_MIB_NAME",
						"sysUpTimeInstance",
					},
				},
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
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
					fakeTime,
				),
			},
		}, /*
			// TODO: Find out why sending traps with auth and/or priv yields invalied packets

			//ordinary v3 coldstart trap auth and no priv
			{
				name:      "v3 coldStart authNoPriv",
				version:   gosnmp.Version3,
				secname:   "test-user",
				seclevel:  "authNoPriv",
				authproto: "SHA",
				authpass:  "passpass",
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
						e: mibEntry{
							"SNMPv2-MIB",
							"snmpTrapOID.0",
						},
					},
					{
						oid: ".1.3.6.1.6.3.1.1.5.1",
						e: mibEntry{
							"SNMPv2-MIB",
							"coldStart",
						},
					},
					{
						oid: ".1.3.6.1.2.1.1.3.0",
						e: mibEntry{
							"UNUSED_MIB_NAME",
							"sysUpTimeInstance",
						},
					},
				},
				metrics: []telegraf.Metric{
					testutil.MustMetric(
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
						fakeTime,
					),
				},
			},
				//ordinary v3 coldStart trap auth and priv
				{
					name:      "v3 coldStart authPriv",
					version:   gosnmp.Version3,
					secname:   "test-user",
					seclevel:  "authPriv",
					authproto: "SHA",
					authpass:  "passpass",
					privproto: "DES",
					privpass:  "passpasspasspass",
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
							e: mibEntry{
								"SNMPv2-MIB",
								"snmpTrapOID.0",
							},
						},
						{
							oid: ".1.3.6.1.6.3.1.1.5.1",
							e: mibEntry{
								"SNMPv2-MIB",
								"coldStart",
							},
						},
						{
							oid: ".1.3.6.1.2.1.1.3.0",
							e: mibEntry{
								"UNUSED_MIB_NAME",
								"sysUpTimeInstance",
							},
						},
					},
					metrics: []telegraf.Metric{
						testutil.MustMetric(
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
							fakeTime,
						),
					},
				},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We would prefer to specify port 0 and let the network
			// stack choose an unused port for us but TrapListener
			// doesn't have a way to return the autoselected port.
			// Instead, we'll use an unusual port and hope it's
			// unused.
			const port = 12399

			// Hook into the trap handler so the test knows when the
			// trap has been received
			received := make(chan int)
			wrap := func(f handler) handler {
				return func(p *gosnmp.SnmpPacket, a *net.UDPAddr) {
					f(p, a)
					received <- 0
				}
			}

			// Set up the service input plugin
			s := &SnmpTrap{
				ServiceAddress:     "udp://:" + strconv.Itoa(port),
				makeHandlerWrapper: wrap,
				timeFunc: func() time.Time {
					return fakeTime
				},
				Log:          testutil.Logger{},
				Version:      tt.version.String(),
				SecName:      tt.secname,
				SecLevel:     tt.seclevel,
				AuthProtocol: tt.authproto,
				AuthPassword: tt.authpass,
				PrivProtocol: tt.privproto,
				PrivPassword: tt.privpass,
				EngineID:     "80001f8880031dd407f608905e00000000",
				EngineBoots:  1,
				EngineTime:   1,
				ContextName:  "public",
			}
			require.Nil(t, s.Init())
			var acc testutil.Accumulator
			require.Nil(t, s.Start(&acc))
			defer s.Stop()

			// Preload the cache with the oids we'll use in this test
			// so snmptranslate and mibs don't need to be installed.
			for _, entry := range tt.entries {
				s.load(entry.oid, entry.e)
			}

			// Don't look up oid with snmptranslate.
			s.execCmd = fakeExecCmd

			// Send the trap
			sendTrap(t, port, now, tt.trap, tt.version, tt.seclevel, tt.secname, tt.authproto, tt.authpass, tt.privproto, tt.privpass)

			// Wait for trap to be received
			select {
			case <-received:
			case <-time.After(2 * time.Second):
				t.Fatal("timed out waiting for trap to be received")
			}

			// Verify plugin output
			testutil.RequireMetricsEqual(t,
				tt.metrics, acc.GetTelegrafMetrics(),
				testutil.SortMetrics())
		})
	}

}
