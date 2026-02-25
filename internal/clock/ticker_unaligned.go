package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

// unaligned delivers ticks at regular but unaligned intervals.  No
// effort is made to avoid drift.
//
// The ticks may have an jitter duration applied to them as an random offset to
// the interval.  However the overall pace of is that of the interval, so on
// average you will have one collection each interval.
//
// The first tick is emitted immediately.
//
// Ticks are dropped for slow consumers.
type unaligned struct {
	clk      clock.Clock
	schedule time.Time
	interval time.Duration
	jitter   time.Duration
	offset   time.Duration
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func (t *unaligned) start() {
	t.ch = make(chan time.Time, 1)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	// Compute the scheduled first tick by adding the offset. By doing so, we
	// do not need to take the offset into account later.
	t.schedule = t.schedule.Add(t.offset)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx)
	}()
}

func (t *unaligned) run(ctx context.Context) {
	// Start with the first scheduled
	timer := t.clk.Timer(t.clk.Until(t.schedule) + internal.RandomDuration(t.jitter))

	for {
		select {
		case ts := <-timer.C:
			// Compute the next scheduled interval by adding the interval and
			// randomizing the timing with the given jitter (if any). Note, we
			// need to remember the next scheduling without adding the ticker
			// to avoid drifting of the ticks by jitter/2 on average!
			t.schedule = t.schedule.Add(t.interval)
			timer.Reset(t.clk.Until(t.schedule) + internal.RandomDuration(t.jitter))

			// Fire our event in a non-blocking fashion to avoid blocking the
			// ticker if the agent code did not read the ticker channel yet.
			select {
			case t.ch <- ts:
			default:
			}
		case <-ctx.Done():
			// Someone stopped the ticker so cleanup and leave
			timer.Stop()
			return
		}
	}
}

func (t *unaligned) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *unaligned) Stop() {
	t.cancel()
	t.wg.Wait()
}
