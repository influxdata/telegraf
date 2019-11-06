package snmp_trap

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/snmp"

	"github.com/soniah/gosnmp"
)

type SnmpTrap struct {
	Port uint16 `toml:"port"`

	acc      telegraf.Accumulator
	listener *gosnmp.TrapListener
	wg       sync.WaitGroup
	timeFunc func() time.Time

	makeHandlerWrapper func(func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr)) func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr)
	Errch              chan error
}

var sampleConfig = `
  ## Port to listen on.  Default 162
  #port = 162
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
			timeFunc: time.Now,
			Port:     162,
		}
	})
}

func (s *SnmpTrap) Init() error {
	return nil
}

func (s *SnmpTrap) Start(acc telegraf.Accumulator) error {
	s.acc = acc
	s.listener = gosnmp.NewTrapListener()
	s.listener.OnNewTrap = makeTrapHandler(s)
	s.listener.Params = gosnmp.Default
	//s.listener.Params.Logger = log.New(os.Stdout, "", 0)

	//wrap the handler, used in unit tests
	if nil != s.makeHandlerWrapper {
		s.listener.OnNewTrap = s.makeHandlerWrapper(s.listener.OnNewTrap)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		//no ip means listen on all interfaces, ipv4 and ipv6
		err := s.listener.Listen(":" + strconv.FormatUint(uint64(s.Port), 10))
		if err != nil {
			s.Errch <- err
			log.Panicf("error in listen: %s", err)
		}
	}()

	return nil
}

func (s *SnmpTrap) Stop() {
	s.listener.Close()
	s.wg.Wait()
}

func makeTrapHandler(s *SnmpTrap) func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	return func(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
		tm := s.timeFunc()
		fields := map[string]interface{}{}
		tags := map[string]string{}

		tags["trap_version"] = packet.Version.String()

		for _, v := range packet.Variables {
			//build a name and value for each variable to use as tags
			//and fields.  defaults are the uninterpreted values
			name := v.Name
			value := v.Value

			//use system mibs to resolve the name if possible
			_, _, oidText, _, err := snmp.SnmpTranslate(v.Name)
			if nil == err {
				name = oidText //would mib name be useful here?
			}

			//todo: format the pdu value based on its snmp type and
			//the mib's textual convention.  The snmp input plugin
			//only handles textual convention for ip and mac addresses

			switch v.Type {
			//case gosnmp.OctetString:
				//b := v.Value.([]byte)
			case gosnmp.ObjectIdentifier:
				s, ok := v.Value.(string)
				if (ok) {
					if _, _, oidText, _, err := snmp.SnmpTranslate(s); ok && nil == err {
						value = oidText //would mib name be useful here?
					}
				}
				//1.3.6.1.6.3.1.1.4.1.0 is SNMPv2-MIB::snmpTrapOID.0.  If
				//v.Name is this oid, set a tag of the trap name.
				if (v.Name == ".1.3.6.1.6.3.1.1.4.1.0") {
					tags["trap_name"] = fmt.Sprintf("%v", value)
					continue
				}
			}

			fields[name] = value
			fields[name+"_type"] = v.Type.String()
		}

		s.acc.AddFields("snmp_trap", fields, tags, tm)
	}
}

func (s *SnmpTrap) Listening() <-chan bool {
	return s.listener.Listening()
}
