package lanz

import (
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
    # servers = [
	#   "tcp://127.0.0.1:50001"
	# ]
`

func init() {
	inputs.Add("lanz", func() telegraf.Input {
		return NewLanz()
	})
}

type Lanz struct {
	Servers []string
	Clients map[string]*LanzClient
}

func NewLanz() *Lanz {
	return &Lanz{}
}

func (l *Lanz) SampleConfig() string {
	return sampleConfig
}

func (l *Lanz) Description() string {
	return "Read metrics off Arista LANZ, via socket"
}

func (l *Lanz) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (l *Lanz) Start(acc telegraf.Accumulator) error {

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

func (l *Lanz) Stop() {
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
	c.acc = acc
	deviceUrl, err := url.Parse(c.Server)
	if err != nil {
		return err
	}
	client := lanz.New(lanz.WithAddr(deviceUrl.Host), lanz.WithBackoff(1*time.Second), lanz.WithTimeout(10*time.Second))
	c.client = &client
	c.in = make(chan *pb.LanzRecord)
	c.done = make(chan bool)
	go func() {
		client.Run(c.in)
		c.done <- true
	}()
	go c.receiver(deviceUrl)
	return nil
}

func (c *LanzClient) receiver(deviceUrl *url.URL) {

	for {

		select {
		case <-c.done:
			return
		case msg := <-c.in:
			cr := msg.GetCongestionRecord()
			if cr != nil {
				vals := map[string]interface{}{
					"timestamp":        int64(cr.GetTimestamp()),
					"queue_size":       int64(cr.GetQueueSize()),
					"time_of_max_qlen": int64(cr.GetTimeOfMaxQLen()),
					"tx_latency":       int64(cr.GetTxLatency()),
					"q_drop_count":     int64(cr.GetQDropCount()),
				}
				tags := map[string]string{
					"intf_name":             cr.GetIntfName(),
					"switch_id":             strconv.FormatInt(int64(cr.GetSwitchId()), 10),
					"port_id":               strconv.FormatInt(int64(cr.GetPortId()), 10),
					"entry_type":            strconv.FormatInt(int64(cr.GetEntryType()), 10),
					"traffic_class":         strconv.FormatInt(int64(cr.GetTrafficClass()), 10),
					"fabric_peer_intf_name": cr.GetFabricPeerIntfName(),
					"source":                deviceUrl.Hostname(),
					"port":                  deviceUrl.Port(),
				}
				c.acc.AddFields("lanz_congestion_record", vals, tags)
			}

			gbur := msg.GetGlobalBufferUsageRecord()
			if gbur != nil {
				vals := map[string]interface{}{
					"timestamp":   int64(gbur.GetTimestamp()),
					"buffer_size": int64(gbur.GetBufferSize()),
					"duration":    int64(gbur.GetDuration()),
				}
				tags := map[string]string{
					"entry_type": strconv.FormatInt(int64(gbur.GetEntryType()), 10),
					"source":     deviceUrl.Hostname(),
					"port":       deviceUrl.Port(),
				}
				c.acc.AddFields("lanz_global_buffer_usage_record", vals, tags)
			}
		}
	}
}

func (c *LanzClient) Stop() {
	c.Lock()
	defer c.Unlock()
	close(c.done)
}
