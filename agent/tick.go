package agent

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf/internal"
)

type RelativeTicker struct {
	ch         chan time.Time
	ticker     *time.Ticker
	jitter     time.Duration
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

func NewRelativeTicker(
	interval time.Duration,
	jitter time.Duration,
	leading bool,
) *RelativeTicker {
	ctx, cancel := context.WithCancel(context.Background())

	t := &RelativeTicker{
		ch:         make(chan time.Time, 1),
		ticker:     time.NewTicker(interval),
		jitter:     jitter,
		cancelFunc: cancel,
	}

	if leading {
		t.ch <- time.Now()
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.relayTime(ctx, leading)
	}()

	return t
}

func (t *RelativeTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *RelativeTicker) Stop() {
	t.cancelFunc()
	t.wg.Wait()
}

func (t *RelativeTicker) relayTime(ctx context.Context, leading bool) {
	for {
		select {
		case tm := <-t.ticker.C:
			internal.SleepContext(ctx, internal.RandomDuration(t.jitter))
			select {
			case t.ch <- tm:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}

type AlignedTicker struct {
	ch         chan time.Time
	interval   time.Duration
	jitter     time.Duration
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

func NewAlignedTicker(now time.Time, interval, jitter time.Duration) *AlignedTicker {
	ctx, cancel := context.WithCancel(context.Background())
	t := &AlignedTicker{
		ch:         make(chan time.Time, 1),
		interval:   interval,
		jitter:     jitter,
		cancelFunc: cancel,
	}

	d := t.next(now)
	timer := time.NewTimer(d)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.run(ctx, timer)
	}()

	return t
}

func (t *AlignedTicker) Elapsed() <-chan time.Time {
	return t.ch
}

func (t *AlignedTicker) Stop() {
	t.cancelFunc()
	t.wg.Wait()
}

func (t *AlignedTicker) next(now time.Time) time.Duration {
	next := internal.AlignTime(now, t.interval)
	d := next.Sub(now)
	d += internal.RandomDuration(t.jitter)
	return d
}

func (t *AlignedTicker) run(ctx context.Context, timer *time.Timer) {
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case now := <-timer.C:
			// Forward time to the elapsed channel.
			select {
			case t.ch <- now:
			default:
			}

			d := t.next(now)
			timer.Reset(d)
		}
	}
}
