package clock

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

func TestAlignedTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	start := clk.Now()
	end := start.Add(60 * time.Second)

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()

	expected := []time.Time{
		time.Unix(10, 0).UTC(),
		time.Unix(20, 0).UTC(),
		time.Unix(30, 0).UTC(),
		time.Unix(40, 0).UTC(),
		time.Unix(50, 0).UTC(),
		time.Unix(60, 0).UTC(),
	}

	actual := make([]time.Time, 0)
	clk.Add(10 * time.Second)
	for !clk.Now().After(end) {
		tm := <-ticker.Elapsed()
		actual = append(actual, tm.UTC())

		clk.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

func TestAlignedTickerJitter(t *testing.T) {
	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	start := clk.Now()
	end := start.Add(61 * time.Second)

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()

	last := start
	for !clk.Now().After(end) {
		select {
		case tm := <-ticker.Elapsed():
			dur := tm.Sub(last)
			// 10s interval + 5s jitter + up to 1s late firing.
			require.LessOrEqual(t, dur, 16*time.Second, "expected elapsed time to be less than 16 seconds, but was %s", dur)
			require.GreaterOrEqual(t, dur, 5*time.Second, "expected elapsed time to be more than 5 seconds, but was %s", dur)
			last = last.Add(interval)
		default:
			clk.Add(1 * time.Second)
		}
	}
}

func TestAlignedTickerOffset(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()
	start := clk.Now()
	end := start.Add(61 * time.Second)

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()

	expected := []time.Time{
		time.Unix(13, 0).UTC(),
		time.Unix(23, 0).UTC(),
		time.Unix(33, 0).UTC(),
		time.Unix(43, 0).UTC(),
		time.Unix(53, 0).UTC(),
	}

	actual := make([]time.Time, 0)
	clk.Add(10*time.Second + offset)
	for !clk.Now().After(end) {
		tm := <-ticker.Elapsed()
		actual = append(actual, tm.UTC())
		clk.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

func TestAlignedTickerMissedTick(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	start := clk.Now()

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()

	clk.Add(25 * time.Second)
	tm := <-ticker.Elapsed()
	require.Equal(t, time.Unix(10, 0).UTC(), tm.UTC())
	clk.Add(5 * time.Second)
	tm = <-ticker.Elapsed()
	require.Equal(t, time.Unix(30, 0).UTC(), tm.UTC())
}

// TestAlignedTickerJitterBehavior shows that AlignedTicker has different behavior.
// It realigns to interval boundaries, so jitter doesn't accumulate as drift.
// However, the average interval is still affected by jitter.
//
// Scenario:
//   - interval = 60s
//   - jitter = 10s
//   - start time = 12:02:22
//
// Behavior:
//   - First trigger: next 60s boundary (12:03:00) + jitter = ~12:03:05
//   - Second trigger: next 60s boundary (12:04:00) + jitter = ~12:04:07
//   - The jitter variation averages out because of realignment
func TestAlignedTickerJitterBehavior(t *testing.T) {
	interval := 60 * time.Second
	jitter := 10 * time.Second
	offset := 0 * time.Second

	// Start at 12:02:22
	start := time.Date(2024, 1, 1, 12, 2, 22, 0, time.UTC)
	clk := clock.NewMock()
	clk.Set(start)

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
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

	t.Logf("=== AlignedTicker (realigns to boundaries) ===")
	t.Logf("Start time:      %s", start.Format("15:04:05"))
	t.Logf("First trigger:   %s", firstTrigger.Format("15:04:05"))
	t.Logf("Last trigger:    %s", lastTrigger.Format("15:04:05"))
	t.Logf("Total elapsed:   %s", totalElapsed)
	t.Logf("Expected:        %s", expectedTime)
	t.Logf("Drift:           %s", drift)
	t.Logf("Avg interval:    %.2fs", totalElapsed.Seconds()/float64(numTicks-1))

	// AlignedTicker realigns to boundaries, so drift is minimal
	// The jitter variations cancel out over time
	if drift < 0 {
		drift = -drift
	}
	require.Less(t, drift, 1*time.Minute,
		"AlignedTicker should have minimal drift due to boundary realignment")
}

// Simulates running the Ticker for an hour and displays stats about the
// operation.
func TestAlignedTickerDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	start := clk.Now()

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	dist.print()
	require.Less(t, 350, dist.count)
	require.True(t, 9 < dist.mean() && dist.mean() < 11)
}

func TestAlignedTickerDistributionWithOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()
	start := clk.Now()

	ticker := &aligned{
		clk:         clk,
		schedule:    start,
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start()
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	dist.print()
	require.Less(t, 350, dist.count)
	require.True(t, 9 < dist.mean() && dist.mean() < 11)
}
