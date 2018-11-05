package agent

import (
	"context"
	"sync"
	"time"

	"github.com/influxdata/telegraf/internal"
)

type Ticker struct {
	C          chan time.Time
	ticker     *time.Ticker
	jitter     time.Duration
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

func NewTicker(
	interval time.Duration,
	jitter time.Duration,
) *Ticker {
	ctx, cancel := context.WithCancel(context.Background())

	t := &Ticker{
		C:          make(chan time.Time, 1),
		ticker:     time.NewTicker(interval),
		jitter:     jitter,
		cancelFunc: cancel,
	}

	t.wg.Add(1)
	go t.relayTime(ctx)

	return t
}

func (t *Ticker) Stop() {
	t.cancelFunc()
	t.wg.Wait()
}

func (t *Ticker) relayTime(ctx context.Context) {
	defer t.wg.Done()
	for {
		select {
		case tm := <-t.ticker.C:
			internal.SleepContext(ctx, internal.RandomDuration(t.jitter))
			select {
			case t.C <- tm:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}
