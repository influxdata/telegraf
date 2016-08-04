package snmp

import (
	"fmt"

	"github.com/soniah/gosnmp"
)

// snmpConnection is an interface which wraps a *gosnmp.GoSNMP object.
// We interact through an interface so we can mock it out in tests.
type snmpConnection interface {
	Host() string
	//BulkWalkAll(string) ([]gosnmp.SnmpPDU, error)
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
}

// gosnmpWrapper wraps a *gosnmp.GoSNMP object so we can use it as a snmpConnection.
type gosnmpWrapper struct {
	*gosnmp.GoSNMP
}

// Host returns the value of GoSNMP.Target.
func (gsw gosnmpWrapper) Host() string {
	return gsw.Target
}

// Walk wraps GoSNMP.Walk() or GoSNMP.BulkWalk(), depending on whether the
// connection is using SNMPv1 or newer.
// Also, if any error is encountered, it will just once reconnect and try again.
func (gsw gosnmpWrapper) Walk(oid string, fn gosnmp.WalkFunc) error {
	var err error
	// On error, retry once.
	// Unfortunately we can't distinguish between an error returned by gosnmp, and one returned by the walk function.
	for i := 0; i < 2; i++ {
		if gsw.Version == gosnmp.Version1 {
			err = gsw.GoSNMP.Walk(oid, fn)
		} else {
			err = gsw.GoSNMP.BulkWalk(oid, fn)
		}
		if err == nil {
			return nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return fmt.Errorf("reconnecting %s", err)
		}
	}
	return err
}

// Get wraps GoSNMP.GET().
// If any error is encountered, it will just once reconnect and try again.
func (gsw gosnmpWrapper) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	var err error
	var pkt *gosnmp.SnmpPacket
	for i := 0; i < 2; i++ {
		pkt, err = gsw.GoSNMP.Get(oids)
		if err == nil {
			return pkt, nil
		}
		if err := gsw.GoSNMP.Connect(); err != nil {
			return nil, fmt.Errorf("reconnecting %s", err)
		}
	}
	return nil, err
}
