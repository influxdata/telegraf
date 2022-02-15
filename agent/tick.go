package agent

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/influxdata/telegraf/internal"
)

type Ticker interface {
	Elapsed() <-chan time.Time
	Stop()
}

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
	interval    time.Duration
	jitter      time.Duration
	offset      time.Duration
	minInterval time.Duration
	ch          chan time.Time
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewAlignedTicker(now time.Time, interval, jitter, offset time.Duration) *AlignedTicker {
	t := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	t.start(now, clock.New())
	return t
}

func (t *AlignedTicker) start(now time.Time, clk clock.Clock) {
	t.ch = make(chan time.Time, 1)

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	d := t.next(now)
	timer := clk.Timer(d)

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

// UnalignedTicker delivers ticks at regular but unaligned intervals.  No
// effort is made to avoid drift.
//
// The ticks may have an jitter duration applied to them as an random offset to
// the interval.  However the overall pace of is that of the interval, so on
// average you will have one collection each interval.
//
// The first tick is emitted immediately.
//
// Ticks are dropped for slow consumers.
type UnalignedTicker struct {
	interval time.Duration
	jitter   time.Duration
	offset   time.Duration
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewUnalignedTicker(interval, jitter, offset time.Duration) *UnalignedTicker {
	t := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	t.start(clock.New())
	return t
}

func (t *UnalignedTicker) start(clk clock.Clock) *UnalignedTicker {
	t.ch = make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	ticker := clk.Ticker(t.interval)
	if t.offset == 0 {
		// Perform initial trigger to stay backward compatible
		t.ch <- clk.Now()
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, ticker, clk)
	}()

	return t
}

func sleep(ctx context.Context, duration time.Duration, clk clock.Clock) error {
	if duration == 0 {
		return nil
	}

	t := clk.Timer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

func (t *UnalignedTicker) run(ctx context.Context, ticker *clock.Ticker, clk clock.Clock) {
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			jitter := internal.RandomDuration(t.jitter)
			err := sleep(ctx, t.offset+jitter, clk)
			if err != nil {
				ticker.Stop()
				return
			}
			select {
			case t.ch <- clk.Now():
			default:
			}
		}
	}
}

func (t *UnalignedTicker) InjectTick() {
	t.ch <- time.Now()
}

func (t *UnalignedTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *UnalignedTicker) Stop() {
	t.cancel()
	t.wg.Wait()
}

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

func (t *RollingTicker) start(clk clock.Clock) *RollingTicker {
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

	return t
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
