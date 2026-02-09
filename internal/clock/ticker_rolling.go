package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

// rolling delivers ticks at regular but unaligned intervals.
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
type rolling struct {
	clk      clock.Clock
	interval time.Duration
	jitter   time.Duration
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func (t *rolling) start() {
	t.ch = make(chan time.Time, 1)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	d := t.next()
	timer := t.clk.Timer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, timer)
	}()
}

func (t *rolling) next() time.Duration {
	return t.interval + internal.RandomDuration(t.jitter)
}

func (t *rolling) run(ctx context.Context, timer *clock.Timer) {
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

func (t *rolling) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *rolling) Stop() {
	t.cancel()
	t.wg.Wait()
}
