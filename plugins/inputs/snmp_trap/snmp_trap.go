//go:generate ../../../tools/readme_config_includer/generator
package snmp_trap

import (
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gosnmp/gosnmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/snmp"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var defaultTimeout = config.Duration(time.Second * 5)

//go:embed sample.conf
var sampleConfig string

type SnmpTrap struct {
	ServiceAddress string          `toml:"service_address"`
	Timeout        config.Duration `toml:"timeout"`
	Version        string          `toml:"version"`
	Path           []string        `toml:"path"`

	// Settings for version 3 security
	SecLevel     string        `toml:"sec_level"`
	SecName      config.Secret `toml:"sec_name"`
	AuthProtocol string        `toml:"auth_protocol"`
	AuthPassword config.Secret `toml:"auth_password"`
	PrivProtocol string        `toml:"priv_protocol"`
	PrivPassword config.Secret `toml:"priv_password"`

	Translator string          `toml:"-"`
	Log        telegraf.Logger `toml:"-"`

	acc      telegraf.Accumulator
	listener *gosnmp.TrapListener

	transl translator
}

type translator interface {
	lookup(oid string) (snmp.MibEntry, error)
}

func (*SnmpTrap) SampleConfig() string {
	return sampleConfig
}

func (s *SnmpTrap) SetTranslator(name string) {
	s.Translator = name
}

func (s *SnmpTrap) Init() error {
	// Set defaults
	if s.ServiceAddress == "" {
		s.ServiceAddress = "udp://:162"
	}

	if len(s.Path) == 0 {
		s.Path = []string{"/usr/share/snmp/mibs"}
	}

	// Check input parameters
	switch s.Translator {
	case "gosmi":
		t, err := newGosmiTranslator(s.Path, s.Log)
		if err != nil {
			return err
		}
		s.transl = t
	case "netsnmp":
		s.transl = newNetsnmpTranslator(s.Timeout)
	default:
		// Ignore the translator for testing if an instance was set
		if s.transl == nil {
			return errors.New("invalid translator value")
		}
	}

	// Setup the SNMP parameters
	params := gosnmp.GoSNMP{
		Port:               gosnmp.Default.Port,
		Transport:          gosnmp.Default.Transport,
		Community:          gosnmp.Default.Community,
		Timeout:            gosnmp.Default.Timeout,
		Retries:            gosnmp.Default.Retries,
		ExponentialTimeout: gosnmp.Default.ExponentialTimeout,
		MaxOids:            gosnmp.Default.MaxOids,
		Logger:             gosnmp.NewLogger(&snmp.Logger{Logger: s.Log}),
	}

	switch s.Version {
	case "1":
		params.Version = gosnmp.Version1
	case "", "2c":
		params.Version = gosnmp.Version2c
	case "3":
		params.Version = gosnmp.Version3

		// Setup the security for v3
		params.SecurityModel = gosnmp.UserSecurityModel

		// Set security mechanisms
		switch strings.ToLower(s.SecLevel) {
		case "noauthnopriv", "":
			params.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			params.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			params.MsgFlags = gosnmp.AuthPriv
		default:
			return fmt.Errorf("unknown security level %q", s.SecLevel)
		}

		// Set authentication
		var security gosnmp.UsmSecurityParameters
		switch strings.ToLower(s.AuthProtocol) {
		case "":
			security.AuthenticationProtocol = gosnmp.NoAuth
		case "md5":
			security.AuthenticationProtocol = gosnmp.MD5
		case "sha":
			security.AuthenticationProtocol = gosnmp.SHA
		case "sha224":
			security.AuthenticationProtocol = gosnmp.SHA224
		case "sha256":
			security.AuthenticationProtocol = gosnmp.SHA256
		case "sha384":
			security.AuthenticationProtocol = gosnmp.SHA384
		case "sha512":
			security.AuthenticationProtocol = gosnmp.SHA512
		default:
			return fmt.Errorf("unknown authentication protocol %q", s.AuthProtocol)
		}

		// Set privacy
		switch strings.ToLower(s.PrivProtocol) {
		case "":
			security.PrivacyProtocol = gosnmp.NoPriv
		case "aes":
			security.PrivacyProtocol = gosnmp.AES
		case "des":
			security.PrivacyProtocol = gosnmp.DES
		case "aes192":
			security.PrivacyProtocol = gosnmp.AES192
		case "aes192c":
			security.PrivacyProtocol = gosnmp.AES192C
		case "aes256":
			security.PrivacyProtocol = gosnmp.AES256
		case "aes256c":
			security.PrivacyProtocol = gosnmp.AES256C
		default:
			return fmt.Errorf("unknown privacy protocol %q", s.PrivProtocol)
		}

		// Set credentials
		secnameSecret, err := s.SecName.Get()
		if err != nil {
			return fmt.Errorf("getting secname failed: %w", err)
		}
		security.UserName = secnameSecret.String()
		secnameSecret.Destroy()

		authPasswdSecret, err := s.AuthPassword.Get()
		if err != nil {
			return fmt.Errorf("getting auth-password failed: %w", err)
		}
		security.AuthenticationPassphrase = authPasswdSecret.String()
		authPasswdSecret.Destroy()

		privPasswdSecret, err := s.PrivPassword.Get()
		if err != nil {
			return fmt.Errorf("getting priv-password failed: %w", err)
		}
		security.PrivacyPassphrase = privPasswdSecret.String()
		privPasswdSecret.Destroy()

		// Enable security settings
		params.SecurityParameters = &security
	default:
		return fmt.Errorf("unknown version %q", s.Version)
	}

	// Initialize the listener
	s.listener = gosnmp.NewTrapListener()
	s.listener.OnNewTrap = s.handler
	s.listener.Params = &params

	return nil
}

func (s *SnmpTrap) Start(acc telegraf.Accumulator) error {
	s.acc = acc

	u, err := url.Parse(s.ServiceAddress)
	if err != nil {
		return fmt.Errorf("invalid service address: %s", s.ServiceAddress)
	}

	// The gosnmp package currently only supports UDP
	if u.Scheme != "udp" {
		return fmt.Errorf("unknown protocol for service address %q", s.ServiceAddress)
	}

	// If the listener immediately returns an error we need to return it
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.listener.Listen(u.Host)
	}()

	select {
	case <-s.listener.Listening():
		s.Log.Infof("Listening on %s", s.ServiceAddress)
	case err := <-errCh:
		return fmt.Errorf("listening failed: %w", err)
	}

	return nil
}

func (*SnmpTrap) Gather(telegraf.Accumulator) error {
	return nil
}

func (s *SnmpTrap) Stop() {
	s.listener.Close()
}

func setTrapOid(tags map[string]string, oid string, e snmp.MibEntry) {
	tags["oid"] = oid
	tags["name"] = e.OidText
	tags["mib"] = e.MibName
}

func (s *SnmpTrap) handler(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	tm := time.Now()
	fields := make(map[string]interface{}, len(packet.Variables)+1)
	tags := map[string]string{
		"version": packet.Version.String(),
		"source":  addr.IP.String(),
	}

	if packet.Version == gosnmp.Version1 {
		// Follow the procedure described in RFC 2576 3.1 to
		// translate a v1 trap to v2.
		var trapOid string

		if packet.GenericTrap >= 0 && packet.GenericTrap < 6 {
			trapOid = ".1.3.6.1.6.3.1.1.5." + strconv.Itoa(packet.GenericTrap+1)
		} else if packet.GenericTrap == 6 {
			trapOid = packet.Enterprise + ".0." + strconv.Itoa(packet.SpecificTrap)
		}

		if trapOid != "" {
			e, err := s.transl.lookup(trapOid)
			if err != nil {
				s.Log.Errorf("Error resolving V1 OID, oid=%s, source=%s: %v", trapOid, tags["source"], err)
				return
			}
			setTrapOid(tags, trapOid, e)
		}

		if packet.AgentAddress != "" {
			tags["agent_address"] = packet.AgentAddress
		}

		fields["sysUpTimeInstance"] = packet.Timestamp
	}

	for _, v := range packet.Variables {
		var value interface{}

		// Use system mibs to resolve oids. Don't fall back to numeric oid
		// because it's not useful enough to the end user and can be difficult
		// to translate or remove from the database later.
		//
		// TODO: format the pdu value based on its snmp type and the mib's
		// textual convention. The snmp input plugin only handles textual
		// convention for ip and mac addresses
		switch v.Type {
		case gosnmp.ObjectIdentifier:
			val, ok := v.Value.(string)
			if !ok {
				s.Log.Errorf("Error getting value OID")
				return
			}

			var e snmp.MibEntry
			var err error
			e, err = s.transl.lookup(val)
			if nil != err {
				s.Log.Errorf("Error resolving value OID, oid=%s, source=%s: %v", val, tags["source"], err)
				return
			}

			value = e.OidText

			// 1.3.6.1.6.3.1.1.4.1.0 is SNMPv2-MIB::snmpTrapOID.0.
			// If v.Name is this oid, set a tag of the trap name.
			if v.Name == ".1.3.6.1.6.3.1.1.4.1.0" {
				setTrapOid(tags, val, e)
				continue
			}
		case gosnmp.OctetString:
			// OctetStrings may contain hex data that needs its own conversion
			if !utf8.Valid(v.Value.([]byte)[:]) {
				value = hex.EncodeToString(v.Value.([]byte))
			} else {
				value = v.Value
			}
		default:
			value = v.Value
		}

		e, err := s.transl.lookup(v.Name)
		if nil != err {
			s.Log.Errorf("Error resolving OID oid=%s, source=%s: %v", v.Name, tags["source"], err)
			return
		}

		fields[e.OidText] = value
	}

	if packet.Version == gosnmp.Version3 {
		if packet.ContextName != "" {
			tags["context_name"] = packet.ContextName
		}
		if packet.ContextEngineID != "" {
			// SNMP RFCs like 3411 and 5343 show engine ID as a hex string
			tags["engine_id"] = fmt.Sprintf("%x", packet.ContextEngineID)
		}
	} else {
		if packet.Community != "" {
			tags["community"] = packet.Community
		}
	}

	s.acc.AddFields("snmp_trap", fields, tags, tm)
}

func init() {
	inputs.Add("snmp_trap", func() telegraf.Input {
		return &SnmpTrap{
			ServiceAddress: "udp://:162",
			Timeout:        defaultTimeout,
			Path:           []string{"/usr/share/snmp/mibs"},
			Version:        "2c",
		}
	})
}
