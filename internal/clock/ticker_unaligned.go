package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

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
	clk      clock.Clock
	interval time.Duration
	jitter   time.Duration
	offset   time.Duration
	ch       chan time.Time
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewUnalignedTicker(interval, jitter, offset time.Duration) *UnalignedTicker {
	t := &UnalignedTicker{
		clk:      clock.New(),
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	t.start()
	return t
}

func (t *UnalignedTicker) start() {
	t.ch = make(chan time.Time, 1)
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	ticker := t.clk.Ticker(t.interval)
	if t.offset == 0 {
		// Perform initial trigger to stay backward compatible
		t.ch <- t.clk.Now()
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, ticker, t.clk)
	}()
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
