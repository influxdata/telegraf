package snmp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

// GosnmpWrapper wraps a *gosnmp.GoSNMP object so we can use it as a snmpConnection.
type GosnmpWrapper struct {
	*gosnmp.GoSNMP
}

// Host returns the value of GoSNMP.Target.
func (gs GosnmpWrapper) Host() string {
	return gs.Target
}

// Walk wraps GoSNMP.Walk() or GoSNMP.BulkWalk(), depending on whether the
// connection is using SNMPv1 or newer.
// Also, if any error is encountered, it will just once reconnect and try again.
func (gs GosnmpWrapper) Walk(oid string, fn gosnmp.WalkFunc) error {
	var err error
	// On error, retry once.
	// Unfortunately we can't distinguish between an error returned by gosnmp, and one returned by the walk function.
	for i := 0; i < 2; i++ {
		if gs.Version == gosnmp.Version1 {
			err = gs.GoSNMP.Walk(oid, fn)
		} else {
			err = gs.GoSNMP.BulkWalk(oid, fn)
		}
		if err == nil {
			return nil
		}
		if err := gs.GoSNMP.Connect(); err != nil {
			return fmt.Errorf("reconnecting: %w", err)
		}
	}
	return err
}

// Get wraps GoSNMP.GET().
// If any error is encountered, it will just once reconnect and try again.
func (gs GosnmpWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	var err error
	var pkt *gosnmp.SnmpPacket
	for i := 0; i < 2; i++ {
		pkt, err = gs.GoSNMP.Get(oids)
		if err == nil {
			return pkt, nil
		}
		if err := gs.GoSNMP.Connect(); err != nil {
			return nil, fmt.Errorf("reconnecting: %w", err)
		}
	}
	return nil, err
}

func NewWrapper(s ClientConfig) (GosnmpWrapper, error) {
	gs := GosnmpWrapper{&gosnmp.GoSNMP{}}

	gs.Timeout = time.Duration(s.Timeout)

	gs.Retries = s.Retries

	switch s.Version {
	case 3:
		gs.Version = gosnmp.Version3
	case 2, 0:
		gs.Version = gosnmp.Version2c
	case 1:
		gs.Version = gosnmp.Version1
	default:
		return GosnmpWrapper{}, fmt.Errorf("invalid version")
	}

	if s.Version < 3 {
		if s.Community == "" {
			gs.Community = "public"
		} else {
			gs.Community = s.Community
		}
	}

	gs.MaxRepetitions = s.MaxRepetitions

	if s.Version == 3 {
		gs.ContextName = s.ContextName

		sp := &gosnmp.UsmSecurityParameters{}
		gs.SecurityParameters = sp
		gs.SecurityModel = gosnmp.UserSecurityModel

		switch strings.ToLower(s.SecLevel) {
		case "noauthnopriv", "":
			gs.MsgFlags = gosnmp.NoAuthNoPriv
		case "authnopriv":
			gs.MsgFlags = gosnmp.AuthNoPriv
		case "authpriv":
			gs.MsgFlags = gosnmp.AuthPriv
		default:
			return GosnmpWrapper{}, fmt.Errorf("invalid secLevel")
		}

		sp.UserName = s.SecName

		switch strings.ToLower(s.AuthProtocol) {
		case "md5":
			sp.AuthenticationProtocol = gosnmp.MD5
		case "sha":
			sp.AuthenticationProtocol = gosnmp.SHA
		case "sha224":
			sp.AuthenticationProtocol = gosnmp.SHA224
		case "sha256":
			sp.AuthenticationProtocol = gosnmp.SHA256
		case "sha384":
			sp.AuthenticationProtocol = gosnmp.SHA384
		case "sha512":
			sp.AuthenticationProtocol = gosnmp.SHA512
		case "":
			sp.AuthenticationProtocol = gosnmp.NoAuth
		default:
			return GosnmpWrapper{}, fmt.Errorf("invalid authProtocol")
		}

		sp.AuthenticationPassphrase = s.AuthPassword

		switch strings.ToLower(s.PrivProtocol) {
		case "des":
			sp.PrivacyProtocol = gosnmp.DES
		case "aes":
			sp.PrivacyProtocol = gosnmp.AES
		case "aes192":
			sp.PrivacyProtocol = gosnmp.AES192
		case "aes192c":
			sp.PrivacyProtocol = gosnmp.AES192C
		case "aes256":
			sp.PrivacyProtocol = gosnmp.AES256
		case "aes256c":
			sp.PrivacyProtocol = gosnmp.AES256C
		case "":
			sp.PrivacyProtocol = gosnmp.NoPriv
		default:
			return GosnmpWrapper{}, fmt.Errorf("invalid privProtocol")
		}

		sp.PrivacyPassphrase = s.PrivPassword

		sp.AuthoritativeEngineID = s.EngineID

		sp.AuthoritativeEngineBoots = s.EngineBoots

		sp.AuthoritativeEngineTime = s.EngineTime
	}
	return gs, nil
}

// SetAgent takes a url (scheme://host:port) and sets the wrapped
// GoSNMP struct's corresponding fields.  This shouldn't be called
// after using the wrapped GoSNMP struct, for example after
// connecting.
func (gs *GosnmpWrapper) SetAgent(agent string) error {
	if !strings.Contains(agent, "://") {
		agent = "udp://" + agent
	}

	u, err := url.Parse(agent)
	if err != nil {
		return err
	}

	// Only allow udp{4,6} and tcp{4,6}.
	// Allowing ip{4,6} does not make sense as specifying a port
	// requires the specification of a protocol.
	// gosnmp does not handle these errors well, which is why
	// they can result in cryptic errors by net.Dial.
	switch u.Scheme {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		gs.Transport = u.Scheme
	default:
		return fmt.Errorf("unsupported scheme: %v", u.Scheme)
	}

	gs.Target = u.Hostname()

	portStr := u.Port()
	if portStr == "" {
		portStr = "161"
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return fmt.Errorf("parsing port: %w", err)
	}
	gs.Port = uint16(port)
	return nil
}
