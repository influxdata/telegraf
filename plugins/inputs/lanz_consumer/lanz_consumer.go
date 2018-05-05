package lanz_consumer

import (
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/aristanetworks/goarista/lanz"
	pb "github.com/aristanetworks/goarista/lanz/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  	## URL to Arista LANZ endpoint
  	# servers = "tcp://localhost:50001"
  	servers = [
		"tcp://switch02.int.example.com:50001",
		"tcp://switch02.int.example.com:50001"
	]	
`

func init() {
	inputs.Add("lanz_consumer", func() telegraf.Input {
		return NewLanzConsumer()
	})
}

type LanzConsumer struct {
	Servers []string
	Clients map[string]*LanzClient
}

func NewLanzConsumer() *LanzConsumer {
	return &LanzConsumer{}
}

func (l *LanzConsumer) SampleConfig() string {
	return sampleConfig
}

func (l *LanzConsumer) Description() string {
	return "Read metrics off Arista LANZ, via socket"
}

func (l *LanzConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (l *LanzConsumer) Start(acc telegraf.Accumulator) error {

	if len(l.Servers) == 0 {
		l.Servers = append(l.Servers, "tcp://127.0.0.1:50001")
	}

	for _, server := range l.Servers {
		c := NewLanzClient()
		c.Server = server
		c.Start(acc)
	}
	return nil
}

func (l *LanzConsumer) Stop() {
	for _, client := range l.Clients {
		client.Stop()
	}
}

type LanzClient struct {
	Server string
	in     chan *pb.LanzRecord
	done   chan bool
	acc    telegraf.Accumulator
	client *lanz.Client
	sync.Mutex
}

func NewLanzClient() *LanzClient {
	return &LanzClient{}
}

func (c *LanzClient) Start(acc telegraf.Accumulator) error {
	c.Lock()
	defer c.Unlock()
	c.acc = acc
	u, err := url.Parse(c.Server)
	if err != nil {
		log.Fatal(err)
	}
	client := lanz.New(lanz.WithAddr(u.Host), lanz.WithBackoff(1*time.Second), lanz.WithTimeout(10*time.Second))
	c.client = &client
	go func() {
		client.Run(c.in)
		c.done <- true
	}()

	c.in = make(chan *pb.LanzRecord)
	c.done = make(chan bool)
	go c.receiver()
	return nil
}

func (c *LanzClient) receiver() {

	u, err := url.Parse(c.Server)
	if err != nil {
		log.Fatal(err)
	}
	for {

		select {
		case <-c.done:
			return
		case msg := <-c.in:
			cr := msg.GetCongestionRecord()
			if cr != nil {
				vals := map[string]interface{}{
					"timestamp":     int64(cr.GetTimestamp()),
					"queueSize":     int64(cr.GetQueueSize()),
					"timeOfMaxQLen": int64(cr.GetTimeOfMaxQLen()),
					"txLatency":     int64(cr.GetTxLatency()),
					"qDropCount":    int64(cr.GetQDropCount()),
				}
				tags := map[string]string{
					"intfName":           cr.GetIntfName(),
					"switchId":           strconv.FormatInt(int64(cr.GetSwitchId()), 10),
					"portId":             strconv.FormatInt(int64(cr.GetPortId()), 10),
					"entryType":          strconv.FormatInt(int64(cr.GetEntryType()), 10),
					"trafficClass":       strconv.FormatInt(int64(cr.GetTrafficClass()), 10),
					"fabricPeerIntfName": cr.GetFabricPeerIntfName(),
					"host":               u.Host,
				}
				c.acc.AddFields("congestionRecord", vals, tags)
			}

			gbur := msg.GetGlobalBufferUsageRecord()
			if gbur != nil {
				vals := map[string]interface{}{
					"timestamp":  int64(gbur.GetTimestamp()),
					"bufferSize": int64(gbur.GetBufferSize()),
					"duration":   int64(gbur.GetDuration()),
				}
				tags := map[string]string{
					"entryType": strconv.FormatInt(int64(gbur.GetEntryType()), 10),
					"host":      u.Host,
				}
				c.acc.AddFields("globalBufferUsageRecord", vals, tags)
			}
		}
	}
}

func (c *LanzClient) Stop() {
	c.Lock()
	defer c.Unlock()
	close(c.done)
}
