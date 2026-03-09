package clock

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/influxdata/telegraf/internal"
)

type Ticker struct {
	C chan time.Time

	clk      clock.Clock
	schedule time.Time
	interval time.Duration
	jitter   time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	cfg *config
}

func NewTicker(interval, jitter, offset time.Duration, opt ...Option) *Ticker {
	// Apply the options
	cfg := &config{
		clk: clock.New(),
	}
	for _, o := range opt {
		o(cfg)
	}

	schedule := cfg.clk.Now()

	// Align the scheduled trigger time to interval borders
	if cfg.align {
		// Add minimum interval size to avoid scheduling exceptionally short
		// intervals. This avoids an issue that can occur where the previous
		// interval ends slightly early due to very minor clock changes.
		schedule = internal.AlignTime(cfg.start.Add(interval/100), interval)
	}

	// Compute the scheduled first tick by adding the offset. By doing so, we
	// do not need to take the offset into account later.
	schedule = schedule.Add(offset)

	// Initialize the ticker instance and start it
	t := &Ticker{
		C:        make(chan time.Time, 1),
		clk:      cfg.clk,
		schedule: schedule,
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

func (t *Ticker) Stop() {
	t.cancel()
	t.wg.Wait()
}

func (t *Ticker) run(ctx context.Context) {
	// Start with the first scheduled tick
	timer := t.clk.Timer(t.clk.Until(t.schedule) + internal.RandomDuration(t.jitter))
	defer timer.Stop()

	if t.cfg.notifier != nil {
		t.cfg.notifier <- true
	}

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
			// ticker if the agent code did not read the channel yet
			select {
			case t.C <- ts:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}
