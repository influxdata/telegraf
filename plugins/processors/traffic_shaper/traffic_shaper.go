package traffic_shaper

import (
	_ "embed"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/limiter"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

type TrafficShaper struct {
	Samples          int `toml:"samples"`
	TimeUnit         time.Duration
	BufferSize       int `toml:"buffer_size"`
	Queue            chan *telegraf.Metric
	Acc              telegraf.Accumulator
	Log              telegraf.Logger `toml:"-"`
	wg               sync.WaitGroup
	messagesInFlight selfstat.Stat
	messagesDropped  selfstat.Stat
}

func (*TrafficShaper) SampleConfig() string {
	return sampleConfig
}

func (t *TrafficShaper) Start(acc telegraf.Accumulator) error {
	t.Queue = make(chan *telegraf.Metric, t.BufferSize)
	t.Acc = acc
	t.wg.Add(1)
	go t.ShapeTraffic()
	t.messagesInFlight = selfstat.Register("traffic_shaper", "messages_inflight", map[string]string{})
	t.messagesDropped = selfstat.Register("traffic_shaper", "messages_dropped", map[string]string{})
	return nil
}

func init() {
	processors.AddStreaming("traffic_shaper", func() telegraf.StreamingProcessor {
		return newTrafficShaper()
	})
}

func (t *TrafficShaper) Stop() {
	t.Log.Debugf("Got stop signal %s", time.Now().String())
	close(t.Queue)
	t.wg.Wait()
	t.Log.Debugf("Got stop signal done waiting %s", time.Now().String())
}

func (t *TrafficShaper) ShapeTraffic() {
	defer t.wg.Done()
	rateLimiter := limiter.NewRateLimiter(t.Samples, t.TimeUnit)
	defer rateLimiter.Stop()
	for metric := range t.Queue {
		<-rateLimiter.C
		t.Acc.AddMetric(*metric)
		t.messagesInFlight.Incr(-1)
	}
}

func (t *TrafficShaper) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	select {
	case t.Queue <- &metric:
		t.messagesInFlight.Incr(1)
		return nil
	default:
		t.messagesDropped.Incr(1)
		metric.Drop()
		return nil
	}
}

func newTrafficShaper() *TrafficShaper {
	return &TrafficShaper{
		TimeUnit: time.Second,
	}
}
