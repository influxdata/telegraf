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
	start := clk.Now()
	end := start.Add(60 * time.Second)

	ticker := &unaligned{
		clk:      clk,
		schedule: start,
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start()
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

	// Wait for the first tick to avoid race conditions between updating the
	// time and starting the timer.
	tm := <-ticker.Elapsed()
	actual = append(actual, tm.UTC())

	// Advance the clock and collect all ticks on the way
	for !clk.Now().After(end) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
			clk.Add(1 * time.Second)
		}
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

	ticker := &unaligned{
		clk:      clk,
		schedule: start,
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start()
	defer ticker.Stop()

	// Collect 60 ticks
	const numTicks = 60
	var triggers []time.Time

	for len(triggers) < numTicks {
		select {
		case tm := <-ticker.Elapsed():
			triggers = append(triggers, tm)
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
	start := clk.Now()

	ticker := &unaligned{
		clk:      clk,
		schedule: start,
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start()
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
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
	start := clk.Now()

	ticker := &unaligned{
		clk:      clk,
		schedule: start,
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start()
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	dist.print()
	require.Less(t, 350, dist.count)
	require.True(t, 9 < dist.mean() && dist.mean() < 11)
}
