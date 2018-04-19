package lanz

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

type Lanz struct {
	Server string
	in     chan *pb.LanzRecord
	done   chan bool
	acc    telegraf.Accumulator
	client *lanz.Client
	sync.Mutex
}

var sampleConfig = `
  ## URL to Arista LANZ endpoint
  # server = "tcp://localhost:50001"
  server = "tcp://localhost:50001"
`

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

	l.Lock()
	defer l.Unlock()

	l.acc = acc

	//for _, serv := range l.Servers {
	//	if _, ok := l.client[serv]; !ok {
	u, err := url.Parse(l.Server)
	if err != nil {
		log.Fatal(err)
	}
	c := lanz.New(lanz.WithAddr(u.Host), lanz.WithBackoff(1*time.Second), lanz.WithTimeout(10*time.Second))
	l.client = &c

	go func() {
		c.Run(l.in)
		l.done <- true
	}()
	//	}

	l.in = make(chan *pb.LanzRecord)
	l.done = make(chan bool)
	go l.receiver()
	//}
	return nil
}

func (l *Lanz) receiver() {

	u, err := url.Parse(l.Server)
	if err != nil {
		log.Fatal(err)
	}
	for {

		select {
		case <-l.done:
			return
		case msg := <-l.in:
			//fmt.Println(msg)
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
				l.acc.AddFields("congestionRecord", vals, tags)
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
				l.acc.AddFields("globalBufferUsageRecord", vals, tags)
			}
		}
	}
}

func (l *Lanz) Stop() {
	l.Lock()
	defer l.Unlock()
	close(l.done)
}

func NewLanz() *Lanz {
	return &Lanz{}
}

func init() {
	inputs.Add("lanz", func() telegraf.Input {
		return NewLanz()
	})
}
