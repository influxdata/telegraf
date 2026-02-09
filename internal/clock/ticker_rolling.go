package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

// RollingTicker delivers ticks at regular but unaligned intervals.
//
// Because the next interval is scheduled based on the interval + jitter, you
// are guaranteed at least interval seconds without missing a tick and ticks
// will be evenly scheduled over time.
//
// On average you will have one collection each interval + (jitter/2).
//
// The first tick is emitted after interval+jitter seconds.
//
// Ticks are dropped for slow consumers.
type RollingTicker struct {
	interval time.Duration
	jitter   time.Duration
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewRollingTicker(interval, jitter time.Duration) *RollingTicker {
	t := &RollingTicker{
		interval: interval,
		jitter:   jitter,
	}
	t.start(clock.New())
	return t
}

func (t *RollingTicker) start(clk clock.Clock) {
	t.ch = make(chan time.Time, 1)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	d := t.next()
	timer := clk.Timer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, timer)
	}()
}

func (t *RollingTicker) next() time.Duration {
	return t.interval + internal.RandomDuration(t.jitter)
}

func (t *RollingTicker) run(ctx context.Context, timer *clock.Timer) {
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case now := <-timer.C:
			select {
			case t.ch <- now:
			default:
			}

			d := t.next()
			timer.Reset(d)
		}
	}
}

func (t *RollingTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *RollingTicker) Stop() {
	t.cancel()
	t.wg.Wait()
}
