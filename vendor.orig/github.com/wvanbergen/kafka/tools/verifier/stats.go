package main

import (
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
)

type MessageMetadata struct {
	Enqueued, Produced, Consumed time.Time
}

func (mm *MessageMetadata) ProduceLatency() time.Duration {
	return time.Duration(mm.Produced.UnixNano() - mm.Enqueued.UnixNano())
}

func (mm *MessageMetadata) ConsumeLatency() time.Duration {
	return time.Duration(mm.Consumed.UnixNano() - mm.Produced.UnixNano())
}

func (mm *MessageMetadata) TotalLatency() time.Duration {
	return time.Duration(mm.Consumed.UnixNano() - mm.Enqueued.UnixNano())
}

type Stats struct {
	enqueued int64
	produced int64
	consumed int64

	produceLatency int64
	consumeLatency int64
}

func (s *Stats) LogEnqueued(msg *sarama.ProducerMessage) {
	atomic.AddInt64(&s.enqueued, 1)
	metadata := msg.Metadata.(*MessageMetadata)
	metadata.Enqueued = time.Now()
}

func (s *Stats) LogProduced(msg *sarama.ProducerMessage) {
	atomic.AddInt64(&s.produced, 1)
	metadata := msg.Metadata.(*MessageMetadata)
	metadata.Produced = time.Now()
	atomic.AddInt64(&s.produceLatency, int64(metadata.ProduceLatency()))
}

func (s *Stats) LogConsumed(msg *sarama.ProducerMessage) {
	atomic.AddInt64(&s.consumed, 1)
	metadata := msg.Metadata.(*MessageMetadata)
	metadata.Consumed = time.Now()
	atomic.AddInt64(&s.consumeLatency, int64(metadata.ConsumeLatency()))
}

func (s *Stats) meanProduceLatency() time.Duration {
	return meanLatency(
		atomic.LoadInt64(&stats.produceLatency),
		atomic.LoadInt64(&stats.produced),
	)
}

func (s *Stats) meanConsumeLatency() time.Duration {
	return meanLatency(
		atomic.LoadInt64(&stats.consumeLatency),
		atomic.LoadInt64(&stats.consumed),
	)
}

func (s *Stats) meanLatency() time.Duration {
	return meanLatency(
		atomic.LoadInt64(&stats.consumeLatency)+atomic.LoadInt64(&stats.produceLatency),
		atomic.LoadInt64(&stats.consumed),
	)
}

func (s *Stats) Print() {
	logger.Printf("Enqueued: %d, produced: %d, consumed: %d.\n",
		atomic.LoadInt64(&stats.enqueued),
		atomic.LoadInt64(&stats.produced),
		atomic.LoadInt64(&stats.consumed),
	)

	logger.Printf("Produce latency: %0.2fms, consume latency: %0.2fms, total latency: %0.2fms.\n",
		float64(s.meanProduceLatency())/float64(time.Millisecond),
		float64(s.meanConsumeLatency())/float64(time.Millisecond),
		float64(s.meanLatency())/float64(time.Millisecond),
	)
}

func meanLatency(totalDuration int64, samples int64) (result time.Duration) {
	defer func() {
		if e := recover(); e != nil {
			result = 0
		}
	}()

	avg := totalDuration / samples
	return time.Duration(avg)
}
