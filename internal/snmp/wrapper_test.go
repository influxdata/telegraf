package snmp

import "github.com/gosnmp/gosnmp"

type testSNMPConnection struct {
	host   string
	values map[string]interface{}
}

func (tsc *testSNMPConnection) Host() string {
	return tsc.host
}

func (tsc *testSNMPConnection) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	sp := &gosnmp.SnmpPacket{}
	for _, oid := range oids {
		v, ok := tsc.values[oid]
		if !ok {
			sp.Variables = append(sp.Variables, gosnmp.SnmpPDU{
				Name: oid,
				Type: gosnmp.NoSuchObject,
			})
			continue
		}
		sp.Variables = append(sp.Variables, gosnmp.SnmpPDU{
			Name:  oid,
			Value: v,
		})
	}
	return sp, nil
}
func (tsc *testSNMPConnection) Walk(oid string, wf gosnmp.WalkFunc) error {
	for void, v := range tsc.values {
		if void == oid || (len(void) > len(oid) && void[:len(oid)+1] == oid+".") {
			if err := wf(gosnmp.SnmpPDU{
				Name:  void,
				Value: v,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (tsc *testSNMPConnection) Reconnect() error {
	return nil
}

var tsc = &testSNMPConnection{
	host: "tsc",
	values: map[string]interface{}{
		".1.0.0.0.1.1.0":         "foo",
		".1.0.0.0.1.1.1":         []byte("bar"),
		".1.0.0.0.1.1.2":         []byte(""),
		".1.0.0.0.1.102":         "bad",
		".1.0.0.0.1.2.0":         1,
		".1.0.0.0.1.2.1":         2,
		".1.0.0.0.1.2.2":         0,
		".1.0.0.0.1.3.0":         "0.123",
		".1.0.0.0.1.3.1":         "0.456",
		".1.0.0.0.1.3.2":         "0.000",
		".1.0.0.0.1.3.3":         "9.999",
		".1.0.0.0.1.5.0":         123456,
		".1.0.0.0.1.6.0":         ".1.0.0.0.1.7",
		".1.0.0.1.1":             "baz",
		".1.0.0.1.2":             234,
		".1.0.0.1.3":             []byte("byte slice"),
		".1.0.0.2.1.5.0.9.9":     11,
		".1.0.0.2.1.5.1.9.9":     22,
		".1.0.0.3.1.1.10":        "instance",
		".1.0.0.3.1.1.11":        "instance2",
		".1.0.0.3.1.1.12":        "instance3",
		".1.0.0.3.1.2.10":        10,
		".1.0.0.3.1.2.11":        20,
		".1.0.0.3.1.2.12":        20,
		".1.0.0.3.1.3.10":        1,
		".1.0.0.3.1.3.11":        2,
		".1.0.0.3.1.3.12":        3,
		".1.3.6.1.2.1.3.1.1.1.0": "foo",
		".1.3.6.1.2.1.3.1.1.1.1": []byte("bar"),
		".1.3.6.1.2.1.3.1.1.1.2": []byte(""),
		".1.3.6.1.2.1.3.1.1.102": "bad",
		".1.3.6.1.2.1.3.1.1.2.0": 1,
		".1.3.6.1.2.1.3.1.1.2.1": 2,
		".1.3.6.1.2.1.3.1.1.2.2": 0,
		".1.3.6.1.2.1.3.1.1.3.0": "1.3.6.1.2.1.3.1.1.3",
		".1.3.6.1.2.1.3.1.1.5.0": 123456,
	},
}
