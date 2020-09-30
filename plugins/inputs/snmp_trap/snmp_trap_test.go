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

func sendTrap(t *testing.T, port uint16, now uint32, trap gosnmp.SnmpTrap, version gosnmp.SnmpVersion, secLevel string, username string, authProto string, authPass string, privProto string, privPass string, contextName string, engineID string) {
	var s gosnmp.GoSNMP

	if version == gosnmp.Version3 {
		var msgFlags gosnmp.SnmpV3MsgFlags
		switch strings.ToLower(secLevel) {
		case "noauthnopriv", "":
			msgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			msgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			msgFlags = gosnmp.AuthPriv
		default:
			msgFlags = gosnmp.NoAuthNoPriv
		}

		var authenticationProtocol gosnmp.SnmpV3AuthProtocol
		switch strings.ToLower(authProto) {
		case "md5":
			authenticationProtocol = gosnmp.MD5
		case "sha":
			authenticationProtocol = gosnmp.SHA
		//case "sha224":
		//	authenticationProtocol = gosnmp.SHA224
		//case "sha256":
		//	authenticationProtocol = gosnmp.SHA256
		//case "sha384":
		//	authenticationProtocol = gosnmp.SHA384
		//case "sha512":
		//	authenticationProtocol = gosnmp.SHA512
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

		sp := &gosnmp.UsmSecurityParameters{
			AuthoritativeEngineID:    "1",
			AuthoritativeEngineBoots: 1,
			AuthoritativeEngineTime:  1,
			UserName:                 username,
			PrivacyProtocol:          privacyProtocol,
			PrivacyPassphrase:        privPass,
			AuthenticationPassphrase: authPass,
			AuthenticationProtocol:   authenticationProtocol,
		}
		s = gosnmp.GoSNMP{
			Port:               port,
			Version:            version,
			Timeout:            time.Duration(2) * time.Second,
			Retries:            1,
			MaxOids:            gosnmp.MaxOids,
			Target:             "127.0.0.1",
			SecurityParameters: sp,
			SecurityModel:      gosnmp.UserSecurityModel,
			MsgFlags:           msgFlags,
			ContextName:        contextName,
			ContextEngineID:    engineID,
		}
	} else {
		s = gosnmp.GoSNMP{
			Port:      port,
			Version:   version,
			Timeout:   time.Duration(2) * time.Second,
			Retries:   1,
			MaxOids:   gosnmp.MaxOids,
			Target:    "127.0.0.1",
			Community: "public",
		}
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
		secName   string // v3 username
		secLevel  string // v3 security level
		authProto string // Auth protocol: "", MD5 or SHA
		authPass  string // Auth passphrase
		privProto string // Priv protocol: "", DES or AES
		privPass  string // Priv passphrase

		// V3 sender context
		contextName string
		engineID    string

		// receive
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
						"community":     "public",
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
						"community":     "public",
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
			name:        "v3 coldStart noAuthNoPriv",
			version:     gosnmp.Version3,
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
					fakeTime,
				),
			},
		},
		//ordinary v3 coldstart trap SHA auth and no priv
		{
			name:      "v3 coldStart authShaNoPriv",
			version:   gosnmp.Version3,
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
		/*
			//ordinary v3 coldstart trap SHA224 auth and no priv
			{
				name:      "v3 coldStart authShaNoPriv",
				version:   gosnmp.Version3,
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
			//ordinary v3 coldstart trap SHA256 auth and no priv
			{
				name:      "v3 coldStart authSha256NoPriv",
				version:   gosnmp.Version3,
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
			//ordinary v3 coldstart trap SHA384 auth and no priv
			{
				name:      "v3 coldStart authSha384NoPriv",
				version:   gosnmp.Version3,
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
			//ordinary v3 coldstart trap SHA512 auth and no priv
			{
				name:      "v3 coldStart authShaNoPriv",
				version:   gosnmp.Version3,
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
		//ordinary v3 coldstart trap SHA auth and no priv
		{
			name:      "v3 coldStart authShaNoPriv",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldstart trap MD5 auth and no priv
		{
			name:      "v3 coldStart authMD5NoPriv",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and AES priv
		{
			name:      "v3 coldStart authSHAPrivAES",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and DES priv
		{
			name:      "v3 coldStart authSHAPrivDES",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and AES192 priv
		{
			name:      "v3 coldStart authSHAPrivAES192",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and AES192C priv
		{
			name:      "v3 coldStart authSHAPrivAES192C",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and AES256 priv
		{
			name:      "v3 coldStart authSHAPrivAES256",
			version:   gosnmp.Version3,
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
		//ordinary v3 coldStart SHA trap auth and AES256C priv
		{
			name:      "v3 coldStart authSHAPrivAES256C",
			version:   gosnmp.Version3,
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
				SecName:      tt.secName,
				SecLevel:     tt.secLevel,
				AuthProtocol: tt.authProto,
				AuthPassword: tt.authPass,
				PrivProtocol: tt.privProto,
				PrivPassword: tt.privPass,
			}
			require.Nil(t, s.Init())
			// Don't look up oid with snmptranslate.
			s.execCmd = fakeExecCmd
			var acc testutil.Accumulator
			require.Nil(t, s.Start(&acc))
			defer s.Stop()

			// Preload the cache with the oids we'll use in this test
			// so snmptranslate and mibs don't need to be installed.
			for _, entry := range tt.entries {
				s.load(entry.oid, entry.e)
			}

			// Send the trap
			sendTrap(t, port, now, tt.trap, tt.version, tt.secLevel, tt.secName, tt.authProto, tt.authPass, tt.privProto, tt.privPass, tt.contextName, tt.engineID)

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
