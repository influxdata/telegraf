package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

// AlignedTicker delivers ticks at aligned times plus an optional jitter.  Each
// tick is realigned to avoid drift and handle changes to the system clock.
//
// The ticks may have an jitter duration applied to them as an random offset to
// the interval.  However the overall pace of is that of the interval, so on
// average you will have one collection each interval.
//
// The first tick is emitted at the next alignment.
//
// Ticks are dropped for slow consumers.
//
// The implementation currently does not recalculate until the next tick with
// no maximum sleep, when using large intervals alignment is not corrected
// until the next tick.
type AlignedTicker struct {
	clk         clock.Clock
	schedule    time.Time
	interval    time.Duration
	jitter      time.Duration
	offset      time.Duration
	minInterval time.Duration
	ch          chan time.Time
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewAlignedTicker(start time.Time, interval, jitter, offset time.Duration) *AlignedTicker {
	t := &AlignedTicker{
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

func (t *AlignedTicker) start() {
	t.ch = make(chan time.Time, 1)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	d := t.next(t.schedule)
	timer := t.clk.Timer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, timer)
	}()
}

func (t *AlignedTicker) next(now time.Time) time.Duration {
	// Add minimum interval size to avoid scheduling an interval that is
	// exceptionally short.  This avoids an issue that can occur where the
	// previous interval ends slightly early due to very minor clock changes.
	next := now.Add(t.minInterval)

	next = internal.AlignTime(next, t.interval)
	d := next.Sub(now)
	if d == 0 {
		d = t.interval
	}
	d += t.offset
	d += internal.RandomDuration(t.jitter)
	return d
}

func (t *AlignedTicker) run(ctx context.Context, timer *clock.Timer) {
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

			d := t.next(now)
			timer.Reset(d)
		}
	}
}

func (t *AlignedTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *AlignedTicker) Stop() {
	t.cancel()
	t.wg.Wait()
}
