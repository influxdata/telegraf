package snmp_trap

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/soniah/gosnmp"
)

var defaultTimeout = internal.Duration{Duration: time.Second * 5}

type handler func(*gosnmp.SnmpPacket, *net.UDPAddr)
type execer func(internal.Duration, string, ...string) ([]byte, error)

type mibEntry struct {
	mibName string
	oidText string
}

type SnmpTrap struct {
	ServiceAddress string            `toml:"service_address"`
	Timeout        internal.Duration `toml:"timeout"`
	Version        string            `toml:"version"`

	// Settings for version 3
	// Values: "noAuthNoPriv", "authNoPriv", "authPriv"
	SecLevel string `toml:"sec_level"`
	SecName  string `toml:"sec_name"`
	// Values: "MD5", "SHA", "". Default: ""
	AuthProtocol string `toml:"auth_protocol"`
	AuthPassword string `toml:"auth_password"`
	// Values: "DES", "AES", "". Default: ""
	PrivProtocol string `toml:"priv_protocol"`
	PrivPassword string `toml:"priv_password"`

	acc      telegraf.Accumulator
	listener *gosnmp.TrapListener
	timeFunc func() time.Time
	errCh    chan error

	makeHandlerWrapper func(handler) handler

	Log telegraf.Logger `toml:"-"`

	cacheLock sync.Mutex
	cache     map[string]mibEntry

	execCmd execer
}

var sampleConfig = `
  ## Transport, local address, and port to listen on.  Transport must
  ## be "udp://".  Omit local address to listen on all interfaces.
  ##   example: "udp://127.0.0.1:1234"
  ##
  ## Special permissions may be required to listen on a port less than
  ## 1024.  See README.md for details
  ##
  # service_address = "udp://:162"
  ## Timeout running snmptranslate command
  # timeout = "5s"
  ## Snmp version, defaults to 2c
  # version = "2c"
  ## SNMPv3 authentication and encryption options.
  ##
  ## Security Name.
  # sec_name = "myuser"
  ## Authentication protocol; one of "MD5", "SHA" or "".
  # auth_protocol = "MD5"
  ## Authentication password.
  # auth_password = "pass"
  ## Security Level; one of "noAuthNoPriv", "authNoPriv", or "authPriv".
  # sec_level = "authNoPriv"
  ## Privacy protocol used for encrypted messages; one of "DES", "AES", "AES192", "AES192C", "AES256", "AES256C" or "".
  # priv_protocol = ""
  ## Privacy password used for encrypted messages.
  # priv_password = ""
`

func (s *SnmpTrap) SampleConfig() string {
	return sampleConfig
}

func (s *SnmpTrap) Description() string {
	return "Receive SNMP traps"
}

func (s *SnmpTrap) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("snmp_trap", func() telegraf.Input {
		return &SnmpTrap{
			timeFunc:       time.Now,
			ServiceAddress: "udp://:162",
			Timeout:        defaultTimeout,
			Version:        "2c",
		}
	})
}

func realExecCmd(Timeout internal.Duration, arg0 string, args ...string) ([]byte, error) {
	cmd := exec.Command(arg0, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (s *SnmpTrap) Init() error {
	s.cache = map[string]mibEntry{}
	s.execCmd = realExecCmd
	return nil
}

func (s *SnmpTrap) Start(acc telegraf.Accumulator) error {
	s.acc = acc
	s.listener = gosnmp.NewTrapListener()
	s.listener.OnNewTrap = makeTrapHandler(s)
	s.listener.Params = gosnmp.Default

	switch s.Version {
	case "3":
		s.listener.Params.Version = gosnmp.Version3
	case "2c":
		s.listener.Params.Version = gosnmp.Version2c
	case "1":
		s.listener.Params.Version = gosnmp.Version1
	default:
		s.listener.Params.Version = gosnmp.Version2c
	}

	if s.listener.Params.Version == gosnmp.Version3 {
		s.listener.Params.SecurityModel = gosnmp.UserSecurityModel

		switch strings.ToLower(s.SecLevel) {
		case "noauthnopriv", "":
			s.listener.Params.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			s.listener.Params.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			s.listener.Params.MsgFlags = gosnmp.AuthPriv
		default:
			return fmt.Errorf("unknown security level '%s'", s.SecLevel)
		}

		var authenticationProtocol gosnmp.SnmpV3AuthProtocol
		switch strings.ToLower(s.AuthProtocol) {
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
			return fmt.Errorf("unknown authentication protocol '%s'", s.AuthProtocol)
		}

		var privacyProtocol gosnmp.SnmpV3PrivProtocol
		switch strings.ToLower(s.PrivProtocol) {
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
			return fmt.Errorf("unknown privacy protocol '%s'", s.PrivProtocol)
		}

		s.listener.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 s.SecName,
			PrivacyProtocol:          privacyProtocol,
			PrivacyPassphrase:        s.PrivPassword,
			AuthenticationPassphrase: s.AuthPassword,
			AuthenticationProtocol:   authenticationProtocol,
		}

	}

	// wrap the handler, used in unit tests
	if nil != s.makeHandlerWrapper {
		s.listener.OnNewTrap = s.makeHandlerWrapper(s.listener.OnNewTrap)
	}

	split := strings.SplitN(s.ServiceAddress, "://", 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid service address: %s", s.ServiceAddress)
	}

	protocol := split[0]
	addr := split[1]

	// gosnmp.TrapListener currently supports udp only.  For forward
	// compatibility, require udp in the service address
	if protocol != "udp" {
		return fmt.Errorf("unknown protocol '%s' in '%s'", protocol, s.ServiceAddress)
	}

	// If (*TrapListener).Listen immediately returns an error we need
	// to return it from this function.  Use a channel to get it here
	// from the goroutine.  Buffer one in case Listen returns after
	// Listening but before our Close is called.
	s.errCh = make(chan error, 1)
	go func() {
		s.errCh <- s.listener.Listen(addr)
	}()

	select {
	case <-s.listener.Listening():
		s.Log.Infof("Listening on %s", s.ServiceAddress)
	case err := <-s.errCh:
		return err
	}

	return nil
}

func (s *SnmpTrap) Stop() {
	s.listener.Close()
	err := <-s.errCh
	if nil != err {
		s.Log.Errorf("Error stopping trap listener %v", err)
	}
}

func setTrapOid(tags map[string]string, oid string, e mibEntry) {
	tags["oid"] = oid
	tags["name"] = e.oidText
	tags["mib"] = e.mibName
}

func makeTrapHandler(s *SnmpTrap) handler {
	return func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		tm := s.timeFunc()
		fields := map[string]interface{}{}
		tags := map[string]string{}

		tags["version"] = packet.Version.String()
		tags["source"] = addr.IP.String()

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
				e, err := s.lookup(trapOid)
				if err != nil {
					s.Log.Errorf("Error resolving V1 OID: %v", err)
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
			// Use system mibs to resolve oids.  Don't fall back to
			// numeric oid because it's not useful enough to the end
			// user and can be difficult to translate or remove from
			// the database later.

			var value interface{}

			// todo: format the pdu value based on its snmp type and
			// the mib's textual convention.  The snmp input plugin
			// only handles textual convention for ip and mac
			// addresses

			switch v.Type {
			case gosnmp.ObjectIdentifier:
				val, ok := v.Value.(string)
				if !ok {
					s.Log.Errorf("Error getting value OID")
					return
				}

				var e mibEntry
				var err error
				e, err = s.lookup(val)
				if nil != err {
					s.Log.Errorf("Error resolving value OID: %v", err)
					return
				}

				value = e.oidText

				// 1.3.6.1.6.3.1.1.4.1.0 is SNMPv2-MIB::snmpTrapOID.0.
				// If v.Name is this oid, set a tag of the trap name.
				if v.Name == ".1.3.6.1.6.3.1.1.4.1.0" {
					setTrapOid(tags, val, e)
					continue
				}
			default:
				value = v.Value
			}

			e, err := s.lookup(v.Name)
			if nil != err {
				s.Log.Errorf("Error resolving OID: %v", err)
				return
			}

			name := e.oidText

			fields[name] = value
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
}

func (s *SnmpTrap) lookup(oid string) (e mibEntry, err error) {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	var ok bool
	if e, ok = s.cache[oid]; !ok {
		// cache miss.  exec snmptranslate
		e, err = s.snmptranslate(oid)
		if err == nil {
			s.cache[oid] = e
		}
		return e, err
	}
	return e, nil
}

func (s *SnmpTrap) clear() {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache = map[string]mibEntry{}
}

func (s *SnmpTrap) load(oid string, e mibEntry) {
	s.cacheLock.Lock()
	defer s.cacheLock.Unlock()
	s.cache[oid] = e
}

func (s *SnmpTrap) snmptranslate(oid string) (e mibEntry, err error) {
	var out []byte
	out, err = s.execCmd(s.Timeout, "snmptranslate", "-Td", "-Ob", "-m", "all", oid)

	if err != nil {
		return e, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	ok := scanner.Scan()
	if err = scanner.Err(); !ok && err != nil {
		return e, err
	}

	e.oidText = scanner.Text()

	i := strings.Index(e.oidText, "::")
	if i == -1 {
		return e, fmt.Errorf("not found")
	}
	e.mibName = e.oidText[:i]
	e.oidText = e.oidText[i+2:]
	return e, nil
}
