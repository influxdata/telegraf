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

	"github.com/soniah/gosnmp"
)

type SnmpTrap struct {
	Port uint16 `toml:"port"`
	//todo add mib settings

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

		for _, v := range packet.Variables {
			//todo: determine which snmp variables are tags and which
			//are fields.  for now everything's a tag

			//todo: look up v.Name smi
			var name string
			name = v.Name

			//todo: format value based on its snmp type
			var value string
			switch v.Type {
			//case gosnmp.OctetString:
			//b := v.Value.([]byte)
			//case gosnmp.ObjectIdentifier:
			//todo: look up v.Value smi
			default:
				value = fmt.Sprintf("%v", v.Value)
			}

			tags[name] = value //fmt.Sprintf("%v", v.Value)
			tags[name+"_type"] = fmt.Sprintf("%v", v.Type)
		}
		fields["foo"] = "bar"
		s.acc.AddFields("snmp_trap", fields, tags, tm)
	}
}

func (s *SnmpTrap) Listening() <-chan bool {
	return s.listener.Listening()
}
