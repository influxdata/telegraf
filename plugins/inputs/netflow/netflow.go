package netflow

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type inData struct {
	remote *net.UDPAddr
	data   []byte
}

type Netflow struct {
	ServiceAddress             string
	AllowedPendingMessages     int
	ResolveApplicationNameByID bool `toml:"resolve_application_name_by_id"`
	ResolveIfnameByIfindex     bool `toml:"resolve_ifname_by_ifindex"`

	sync.Mutex
	wg sync.WaitGroup

	in   chan inData
	done chan struct{}

	// for v9
	v9WriteTemplate       chan *V9TemplateWriteOp
	v9WriteOptionTemplate chan *V9OptionTemplateWriteOp
	v9ReadTemplate        chan *V9TemplateReadOp
	v9ReadFlowField       chan *V9FlowFieldReadOp

	// for ipfix
	ipfixWriteTemplate          chan *IpfixTemplateWriteOp
	ipfixWriteOptionTemplate    chan *IpfixOptionTemplateWriteOp
	ipfixReadTemplate           chan *IpfixTemplateReadOp
	ipfixReadInformationElement chan *IpfixInformationElementReadOp

	// for common
	readApplication chan *ApplicationReadOp
	writeIfname     chan *IfnameWriteOp
	readIfname      chan *IfnameReadOp

	acc telegraf.Accumulator

	listener *net.UDPConn
}

const MAX_PACKET_SIZE int = 64 * 2048

var dropwarn = "ERROR: Message queue full. Discarding line [%s] " +
	"You may want to increase allowed_pending_messages in the config\n"

const sampleConfig = `
  ## Address and port to host Netflow listener on
  service_address = ":2055"
  ## Number of Netflow messages allowed to queue up. Once filled, the
  ## Netflow listener will start dropping packets.
  allowed_pending_messages = 10000

  resolve_application_name_by_id = true
  resolve_ifname_by_ifindex = true

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
`

func (n *Netflow) SampleConfig() string {
	return sampleConfig
}

func (n *Netflow) Description() string {
	return "Netflow listener"
}

func (n *Netflow) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (n *Netflow) Start(acc telegraf.Accumulator) error {
	n.Lock()
	defer n.Unlock()

	n.acc = acc
	n.in = make(chan inData, n.AllowedPendingMessages)
	n.done = make(chan struct{})

	// for v9
	n.v9WriteTemplate = make(chan *V9TemplateWriteOp)
	n.v9WriteOptionTemplate = make(chan *V9OptionTemplateWriteOp)
	n.v9ReadTemplate = make(chan *V9TemplateReadOp)
	n.v9ReadFlowField = make(chan *V9FlowFieldReadOp)

	// for ipfix
	n.ipfixWriteTemplate = make(chan *IpfixTemplateWriteOp)
	n.ipfixWriteOptionTemplate = make(chan *IpfixOptionTemplateWriteOp)
	n.ipfixReadTemplate = make(chan *IpfixTemplateReadOp)
	n.ipfixReadInformationElement = make(chan *IpfixInformationElementReadOp)

	// for v9 & ipfix
	n.readApplication = make(chan *ApplicationReadOp)
	n.writeIfname = make(chan *IfnameWriteOp)
	n.readIfname = make(chan *IfnameReadOp)

	n.wg.Add(8)
	go n.netflowListen()
	go n.netflowParser()

	// v9
	go n.v9TemplatePoller()
	go n.v9FlowFieldPoller()

	// ipfix
	go n.ipfixTemplatePoller()
	go n.ipfixInformationElementPoller()

	// for v9 & ipfix
	go n.applicationPoller()
	go n.ifnamePoller()

	log.Printf("I! Started Netflow listener service on %s\n", n.ServiceAddress)
	return nil
}

func (n *Netflow) Stop() {
	close(n.done)
	n.listener.Close()
	n.wg.Wait()
	close(n.in)
	log.Println("I! Stopped Netflow listener service on ", n.ServiceAddress)
}

func (n *Netflow) netflowListen() error {
	defer n.wg.Done()
	var err error
	address, _ := net.ResolveUDPAddr("udp", n.ServiceAddress)
	n.listener, err = net.ListenUDP("udp", address)
	if err != nil {
		log.Printf("E! %s\n", err.Error())
	}
	log.Println("I! Netflow listening on: ", n.listener.LocalAddr().String())

	buf := make([]byte, MAX_PACKET_SIZE)
	for {
		select {
		case <-n.done:
			return nil
		default:
			s, remote, err := n.listener.ReadFromUDP(buf)
			if err != nil && !strings.Contains(err.Error(), "closed network") {
				log.Printf("E! %s\n", err.Error())
				continue
			}
			bufCopy := make([]byte, s)
			copy(bufCopy, buf[:s])
			select {
			case n.in <- inData{remote: remote, data: bufCopy}:
			default:
				log.Printf("W! %s, buf=%s", dropwarn, string(bufCopy))
			}
		}
	}
}

func (n *Netflow) netflowParser() error {
	defer n.wg.Done()
	for {
		select {
		case <-n.done:
			return nil
		case inData := <-n.in:
			log.Printf("D! Total Length=%d", len(inData.data))
			log.Printf("D! Raw Data: %v", inData.data)
			version := binary.BigEndian.Uint16(inData.data)
			log.Printf("D! Version=%d", version)

			var frame *bytes.Buffer
			frame = bytes.NewBuffer(inData.data)

			switch version {
			case 5:
				var header = new(V5Header)
				if err := binary.Read(frame, binary.BigEndian, header); err != nil {
					log.Printf("E! %s\n", err.Error())
				}
				var metrics []telegraf.Metric
				for i := uint16(0); i < header.Count; i++ {
					fields := make(map[string]interface{})
					tags := make(map[string]string)
					var record = new(V5FlowRecord)
					if err := binary.Read(frame, binary.BigEndian, record); err != nil {
						log.Printf("E! %s\n", err.Error())
					}
					fields["src_addr"] = record.SrcAddr
					fields["dst_addr"] = record.DstAddr
					fields["nexthop"] = record.Nexthop
					fields["input"] = record.Input
					fields["output"] = record.Output
					fields["packets"] = record.Packets
					fields["bytes"] = record.Bytes
					fields["first"] = record.First
					fields["last"] = record.Last
					fields["src_port"] = record.SrcPort
					fields["dst_port"] = record.DstPort
					fields["tcp_flags"] = record.TCPFlags
					fields["protocol"] = record.Protocol
					fields["tos"] = record.ToS
					fields["src_as"] = record.SrcAS
					fields["dst_as"] = record.DstAS
					fields["src_mask"] = record.SrcMask
					fields["dst_mask"] = record.DstMask
					tags["exporter"] = inData.remote.IP.String()
					m, err := metric.New("netflow", tags, fields, time.Now())
					if err != nil {
						log.Printf("E! %s\n", err.Error())
					}
					log.Printf("I! fields: %v", fields)
					log.Printf("I! tags: %v", tags)
					metrics = append(metrics, m)
				}
				n.storeMetrics(metrics)
			case 9:
				var header = new(V9Header)
				if err := binary.Read(frame, binary.BigEndian, header); err != nil {
					log.Printf("E! %s\n", err.Error())
				}
				log.Printf("D! Version=%d", header.Version)
				log.Printf("D! Count=%d", header.Count)

				var count = uint16(0)
			Loop:
				for count < header.Count {
					var fsId uint16
					var fsLen uint16
					if err := binary.Read(frame, binary.BigEndian, &fsId); err != nil {
						log.Printf("E! %s\n", err.Error())
					}
					if err := binary.Read(frame, binary.BigEndian, &fsLen); err != nil {
						log.Printf("E! %s\n", err.Error())
					}
					log.Printf("D! Flowset ID=%d", fsId)
					log.Printf("D! Flowset Length=%d", fsLen)
					switch {
					case fsId == 0: // template flowset
						log.Printf("I! Template Flowset!")
						count += n.parseV9TemplateFlowset(frame, inData.remote, fsLen)
					case fsId == 1: // template flowset
						log.Printf("I! Option Template Flowset!")
						count += n.parseV9OptionTemplateFlowset(frame, inData.remote, fsLen)
					case fsId >= 256: // data flowset
						log.Print("I! Data Flowset!")
						recordCount, metrics, ok := n.parseV9DataFlowset(frame, inData.remote, fsId, fsLen)
						if !ok {
							break Loop
						}
						if metrics != nil {
							log.Printf("D! store metrics: %v", metrics)
							n.storeMetrics(metrics)
						}
						count += recordCount
					default:
						log.Printf("W! Flowset ID=%d is not supported", fsId)
					}
				}
			case 10:
				var header = new(IpfixHeader)
				if err := binary.Read(frame, binary.BigEndian, header); err != nil {
					log.Fatal(err)
				}
				log.Printf("D! Header Length=%d", header.Length)
				var current = uint16(16) // version + length + export time + sequence number + observation domain id
				for current < header.Length {
					var setHeader = new(IpfixSetHeader)
					if err := binary.Read(frame, binary.BigEndian, setHeader); err != nil {
						log.Printf("E! %s\n", err.Error())
					}
					log.Printf("D! Set Header ID=%d", setHeader.SetID)
					log.Printf("D! Set Header Length=%d", setHeader.Length)
					switch {
					case setHeader.SetID == 2:
						log.Printf("I! Template Set!")
						n.parseIpfixTemplateSet(frame, inData.remote, setHeader.Length)
					case setHeader.SetID == 3:
						log.Printf("I! Option Template Set!")
						n.parseIpfixOptionTemplateSet(frame, inData.remote, setHeader.Length)
					case setHeader.SetID >= 256:
						log.Printf("I! Data Set!")
						metrics := n.parseIpfixDataSet(frame, inData.remote, setHeader.SetID, setHeader.Length)
						if metrics != nil {
							log.Printf("D! Store metrics: %v", metrics)
							n.storeMetrics(metrics)
						}
					default:
						log.Printf("W! Set ID=%d is not supported", setHeader.SetID)
					}
					current += setHeader.Length
				}
			}
		}
	}
}

func (n *Netflow) storeMetrics(metrics []telegraf.Metric) error {
	n.Lock()
	defer n.Unlock()
	for _, m := range metrics {
		n.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	return nil
}

func init() {
	inputs.Add("netflow", func() telegraf.Input {
		return &Netflow{}
	})
}
