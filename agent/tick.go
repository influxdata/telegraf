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
	minInterval time.Duration
	ch          chan time.Time
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewAlignedTicker(now time.Time, interval, jitter time.Duration) *AlignedTicker {
	return newAlignedTicker(now, interval, jitter, clock.New())
}

func newAlignedTicker(now time.Time, interval, jitter time.Duration, clock clock.Clock) *AlignedTicker {
	ctx, cancel := context.WithCancel(context.Background())
	t := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		minInterval: interval / 100,
		ch:          make(chan time.Time, 1),
		cancel:      cancel,
	}

	d := t.next(now)
	timer := clock.Timer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, timer)
	}()

	return t
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
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewUnalignedTicker(interval, jitter time.Duration) *UnalignedTicker {
	return newUnalignedTicker(interval, jitter, clock.New())
}

func newUnalignedTicker(interval, jitter time.Duration, clock clock.Clock) *UnalignedTicker {
	ctx, cancel := context.WithCancel(context.Background())
	t := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		ch:       make(chan time.Time, 1),
		cancel:   cancel,
	}

	ticker := clock.Ticker(t.interval)
	t.ch <- clock.Now()

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, ticker, clock)
	}()

	return t
}

func sleep(ctx context.Context, duration time.Duration, clock clock.Clock) error {
	if duration == 0 {
		return nil
	}

	t := clock.Timer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

func (t *UnalignedTicker) run(ctx context.Context, ticker *clock.Ticker, clock clock.Clock) {
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			jitter := internal.RandomDuration(t.jitter)
			err := sleep(ctx, jitter, clock)
			if err != nil {
				ticker.Stop()
				return
			}
			select {
			case t.ch <- clock.Now():
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
	timer    *clock.Timer
}

func NewRollingTicker(interval, jitter time.Duration) *RollingTicker {
	return newRollingTicker(interval, jitter, clock.New())
}

func newRollingTicker(interval, jitter time.Duration, clock clock.Clock) *RollingTicker {
	ctx, cancel := context.WithCancel(context.Background())
	t := &RollingTicker{
		interval: interval,
		jitter:   jitter,
		ch:       make(chan time.Time, 1),
		cancel:   cancel,
	}

	d := t.next()
	t.timer = clock.Timer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx)
	}()

	return t
}

func (t *RollingTicker) next() time.Duration {
	return t.interval + internal.RandomDuration(t.jitter)
}

func (t *RollingTicker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.timer.Stop()
			return
		case now := <-t.timer.C:
			select {
			case t.ch <- now:
			default:
			}

			t.Reset()
		}
	}
}

// Reset the ticker to the next interval + jitter.
func (t *RollingTicker) Reset() {
	t.timer.Reset(t.next())
}

func (t *RollingTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *RollingTicker) Stop() {
	t.cancel()
	t.wg.Wait()
}
