package clock

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

func TestUnalignedTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	clk.Add(1 * time.Second)

	startup := make(chan bool, 1)

	ticker := NewTicker(interval, jitter, offset, WithClock(clk), WithStartupNotification(startup))
	defer ticker.Stop()

	expected := []time.Time{
		time.Unix(1, 0).UTC(),
		time.Unix(11, 0).UTC(),
		time.Unix(21, 0).UTC(),
		time.Unix(31, 0).UTC(),
		time.Unix(41, 0).UTC(),
		time.Unix(51, 0).UTC(),
		time.Unix(61, 0).UTC(),
	}

	actual := make([]time.Time, 0, len(expected))

	// Wait for the ticker to startup
	<-startup

	// The first tick fires immediately inside the Timer() constructor
	// (Timer(0) with deadline == now), so no clock advance is needed.
	// Blocking on each read ensures the ticker goroutine has called
	// timer.Reset() before the next clock advance, preventing the
	// race between clk.Add() and timer.Reset() that causes flaky
	// deadline shifts.
	actual = append(actual, (<-ticker.C).UTC())
	for i := 1; i < len(expected); i++ {
		clk.Add(interval)
		actual = append(actual, (<-ticker.C).UTC())
	}
	require.Equal(t, expected, actual)
}

// TestUnalignedTickerJitterBehavior shows UnalignedTicker behavior with jitter.
// UnalignedTicker uses a fixed interval ticker internally, so jitter only adds
// delay but doesn't cause cumulative drift.
func TestUnalignedTickerJitterBehavior(t *testing.T) {
	interval := 60 * time.Second
	jitter := 10 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	start := clk.Now()

	ticker := NewTicker(interval, jitter, offset, WithClock(clk))
	defer ticker.Stop()

	// Collect 60 ticks
	const numTicks = 60
	var triggers []time.Time

	for len(triggers) < numTicks {
		select {
		case ts := <-ticker.C:
			triggers = append(triggers, ts)
		default:
			clk.Add(1 * time.Second)
		}
	}

	firstTrigger := triggers[0]
	lastTrigger := triggers[numTicks-1]
	totalElapsed := lastTrigger.Sub(firstTrigger)
	expectedTime := time.Duration(numTicks-1) * interval
	drift := totalElapsed - expectedTime

	t.Logf("=== UnalignedTicker (fixed ticker + jitter sleep) ===")
	t.Logf("Start time:      %s", start.Format("15:04:05"))
	t.Logf("First trigger:   %s", firstTrigger.Format("15:04:05"))
	t.Logf("Last trigger:    %s", lastTrigger.Format("15:04:05"))
	t.Logf("Total elapsed:   %s", totalElapsed)
	t.Logf("Expected:        %s", expectedTime)
	t.Logf("Drift:           %s", drift)
	t.Logf("Avg interval:    %.2fs", totalElapsed.Seconds()/float64(numTicks-1))

	// UnalignedTicker uses clk.Ticker(interval) which fires at fixed intervals
	// The jitter is added as sleep AFTER each tick, but the ticker rhythm is fixed
	// So drift should be minimal (jitter variations average out)
	if drift < 0 {
		drift = -drift
	}
	require.Less(t, drift, 1*time.Minute, "UnalignedTicker should have minimal drift due to fixed internal ticker")
}

// Simulates running the Ticker for an hour and displays stats about the
// operation.
func TestUnalignedTickerDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()

	ticker := NewTicker(interval, jitter, offset, WithClock(clk))
	defer ticker.Stop()

	dist := simulatedTickerDist(ticker, clk)
	dist.print()
	require.Less(t, 350, dist.count)
	require.True(t, 9 < dist.mean() && dist.mean() < 11)
}

func TestUnalignedTickerDistributionWithOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()

	ticker := NewTicker(interval, jitter, offset, WithClock(clk))
	defer ticker.Stop()

	dist := simulatedTickerDist(ticker, clk)
	dist.print()
	require.Less(t, 350, dist.count)
	require.True(t, 9 < dist.mean() && dist.mean() < 11)
}
