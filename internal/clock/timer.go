package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

// Timer delivers ticks at regular but unaligned intervals.
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
type Timer struct {
	C        chan time.Time
	clk      clock.Clock
	interval time.Duration
	jitter   time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	cfg *config
}

func NewTimer(interval, jitter time.Duration, opt ...Option) *Timer {
	// Apply the options
	cfg := &config{
		clk: clock.New(),
	}
	for _, o := range opt {
		o(cfg)
	}

	// Initialize the timer instance and start it
	t := &Timer{
		C:        make(chan time.Time, 1),
		clk:      cfg.clk,
		interval: interval,
		jitter:   jitter,
		cfg:      cfg,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx)
	}()

	return t
}

func (t *Timer) Stop() {
	t.cancel()
	t.wg.Wait()
}

func (t *Timer) run(ctx context.Context) {
	timer := t.clk.Timer(t.interval + internal.RandomDuration(t.jitter))
	defer timer.Stop()

	if t.cfg.notifier != nil {
		t.cfg.notifier <- true
	}

	for {
		select {
		case ts := <-timer.C:
			// Compute the next tick by adding the interval and randomizing the
			// timing with the given jitter (if any). We don't ensure evenly
			// spaced ticks here but rather guarantee the minimum time between
			// ticks being 'interval' long. Note, on average the space between
			// ticks will be interval plus jitter/2!
			timer.Reset(t.interval + internal.RandomDuration(t.jitter))

			// Fire our event in a non-blocking fashion to avoid blocking the
			// timer if the agent code did not read the channel yet
			select {
			case t.C <- ts:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}
