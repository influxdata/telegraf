package clock

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

func TestTimer(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second

	clk := clock.NewMock()
	clk.Add(1 * time.Second)

	start := clk.Now()
	end := start.Add(60 * time.Second)

	startup := make(chan bool, 1)

	timer := NewTimer(interval, jitter, WithClock(clk), WithStartupNotification(startup))
	defer timer.Stop()

	expected := []time.Time{
		time.Unix(11, 0).UTC(),
		time.Unix(21, 0).UTC(),
		time.Unix(31, 0).UTC(),
		time.Unix(41, 0).UTC(),
		time.Unix(51, 0).UTC(),
		time.Unix(61, 0).UTC(),
	}

	// Wait for the timer to startup
	<-startup

	actual := make([]time.Time, 0, len(expected))
	for !clk.Now().After(end) {
		select {
		case ts := <-timer.C:
			actual = append(actual, ts.UTC())
		default:
			clk.Add(1 * time.Second)
		}
	}

	require.Equal(t, expected, actual)
}

// TestTimerJitterDrift demonstrates that with a Timer, jitter causes drift
// over time. Each tick = interval + random(0, jitter), so the
// average tick distance is interval + jitter/2.
//
// Scenario from issue #17287:
//   - interval = 60s
//   - jitter = 10s
//
// Current behavior:
//   - Each tick: interval + random(0-10s)
//   - Average interval: 60s + 5s = 65s
//   - After 60 ticks: expected 60min, actual ~65min (5min drift)
//
// This is intentional as Timer is used for flushing and there we want to
// guarantee a minimum pause between flush cycles.
func TestTimerJitterDrift(t *testing.T) {
	interval := 60 * time.Second
	jitter := 10 * time.Second

	clk := clock.NewMock()
	start := clk.Now()

	timer := NewTimer(interval, jitter, WithClock(clk))
	defer timer.Stop()

	// Collect 60 ticks
	const numTicks = 60
	var triggers []time.Time

	for len(triggers) < numTicks {
		select {
		case ts := <-timer.C:
			triggers = append(triggers, ts)
		default:
			clk.Add(1 * time.Second)
		}
	}

	// Calculate total elapsed time
	firstTrigger := triggers[0]
	lastTrigger := triggers[numTicks-1]
	totalElapsed := lastTrigger.Sub(firstTrigger)

	// Expected time for 59 intervals: 59 * 60s = 59 minutes
	expectedTime := time.Duration(numTicks-1) * interval

	// Calculate drift
	drift := totalElapsed - expectedTime

	t.Logf("=== Timer (interval + jitter each tick) ===")
	t.Logf("Start time:      %s", start.Format("15:04:05"))
	t.Logf("First trigger:   %s", firstTrigger.Format("15:04:05"))
	t.Logf("Last trigger:    %s", lastTrigger.Format("15:04:05"))
	t.Logf("Total elapsed:   %s", totalElapsed)
	t.Logf("Expected:        %s (if no jitter drift)", expectedTime)
	t.Logf("Drift:           %s", drift)
	t.Logf("Avg interval:    %.2fs (expected ~65s with jitter)", totalElapsed.Seconds()/float64(numTicks-1))

	// Current behavior: drift should be ~5 minutes (59 intervals * 5s avg jitter)
	// This confirms the bug from issue #17287
	require.Greater(t, drift, 2*time.Minute,
		"Expected significant drift with Timer jitter behavior")
	require.Less(t, drift, 10*time.Minute,
		"Drift is larger than expected maximum")
}

// Simulates running the Ticker for an hour and displays stats about the
// operation.
func TestTimerDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second

	clk := clock.NewMock()

	timer := NewTimer(interval, jitter, WithClock(clk))
	defer timer.Stop()

	dist := simulatedTimerDist(timer, clk)
	dist.print()
	require.Less(t, 275, dist.count)
	require.True(t, 12 < dist.mean() && 13 > dist.mean())
}

func simulatedTimerDist(timer *Timer, clk *clock.Mock) distribution {
	start := clk.Now()
	end := start.Add(1 * time.Hour)

	var dist distribution

	last := clk.Now()
	for !clk.Now().After(end) {
		select {
		case ts := <-timer.C:
			dist.buckets[ts.Second()]++
			dist.count++
			dist.waittime += ts.Sub(last).Seconds()
			last = ts
		default:
			clk.Add(1 * time.Second)
		}
	}

	return dist
}
