package clock

import (
	"time"

	"github.com/benbjohnson/clock"
)

type Ticker interface {
	Elapsed() <-chan time.Time
	Stop()
}

func NewTicker(start time.Time, interval, jitter, offset time.Duration, align bool) Ticker {
	if align {
		return newAlignedTicker(start, interval, jitter, offset)
	}
	return newUnalignedTicker(interval, jitter, offset)
}

func newAlignedTicker(start time.Time, interval, jitter, offset time.Duration) *aligned {
	t := &aligned{
		clk:         clock.New(),
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	t.start()
	return t
}

func newUnalignedTicker(interval, jitter, offset time.Duration) *unaligned {
	t := &unaligned{
		clk:      clock.New(),
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	t.start()
	return t
}
