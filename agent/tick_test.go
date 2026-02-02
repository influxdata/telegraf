package agent

import (
	"fmt"
	"strings"
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
	since := clk.Now()
	until := since.Add(60 * time.Second)

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
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
	for !clk.Now().After(until) {
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
	since := clk.Now()
	until := since.Add(61 * time.Second)

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
	defer ticker.Stop()

	last := since
	for !clk.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			dur := tm.Sub(last)
			// 10s interval + 5s jitter + up to 1s late firing.
			require.LessOrEqual(t, dur, 16*time.Second, "expected elapsed time to be less than 16 seconds, but was %s", dur)
			require.GreaterOrEqual(t, dur, 5*time.Second, "expected elapsed time to be more than 5 seconds, but was %s", dur)
			last = last.Add(interval)
		default:
		}
		clk.Add(1 * time.Second)
	}
}

func TestAlignedTickerOffset(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()
	since := clk.Now()
	until := since.Add(61 * time.Second)

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
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
	for !clk.Now().After(until) {
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
	since := clk.Now()

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
	defer ticker.Stop()

	clk.Add(25 * time.Second)
	tm := <-ticker.Elapsed()
	require.Equal(t, time.Unix(10, 0).UTC(), tm.UTC())
	clk.Add(5 * time.Second)
	tm = <-ticker.Elapsed()
	require.Equal(t, time.Unix(30, 0).UTC(), tm.UTC())
}

func TestUnalignedTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	clk.Add(1 * time.Second)
	since := clk.Now()
	until := since.Add(60 * time.Second)

	ticker := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start(clk)
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

	actual := make([]time.Time, 0)
	for !clk.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
		}
		clk.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

func TestRollingTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	clk.Add(1 * time.Second)
	since := clk.Now()
	until := since.Add(60 * time.Second)

	ticker := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start(clk)
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

	actual := make([]time.Time, 0)
	for !clk.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
		}
		clk.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

// TestRollingTickerJitterDrift demonstrates that with RollingTicker,
// jitter causes drift over time. Each tick = interval + random(0, jitter),
// so average interval = interval + jitter/2.
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
// This demonstrates the bug where jitter increases effective collection interval.
func TestRollingTickerJitterDrift(t *testing.T) {
	interval := 60 * time.Second
	jitter := 10 * time.Second

	clk := clock.NewMock()
	startTime := clk.Now()

	ticker := &RollingTicker{
		interval: interval,
		jitter:   jitter,
	}
	ticker.start(clk)
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

	// Calculate total elapsed time
	firstTrigger := triggers[0]
	lastTrigger := triggers[numTicks-1]
	totalElapsed := lastTrigger.Sub(firstTrigger)

	// Expected time for 59 intervals: 59 * 60s = 59 minutes
	expectedTime := time.Duration(numTicks-1) * interval

	// Calculate drift
	drift := totalElapsed - expectedTime

	t.Logf("=== RollingTicker (interval + jitter each tick) ===")
	t.Logf("Start time:      %s", startTime.Format("15:04:05"))
	t.Logf("First trigger:   %s", firstTrigger.Format("15:04:05"))
	t.Logf("Last trigger:    %s", lastTrigger.Format("15:04:05"))
	t.Logf("Total elapsed:   %s", totalElapsed)
	t.Logf("Expected:        %s (if no jitter drift)", expectedTime)
	t.Logf("Drift:           %s", drift)
	t.Logf("Avg interval:    %.2fs (expected ~65s with jitter)", totalElapsed.Seconds()/float64(numTicks-1))

	// Current behavior: drift should be ~5 minutes (59 intervals * 5s avg jitter)
	// This confirms the bug from issue #17287
	require.Greater(t, drift, 2*time.Minute,
		"Expected significant drift with RollingTicker jitter behavior")
	require.Less(t, drift, 10*time.Minute,
		"Drift is larger than expected maximum")
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
	startTime := time.Date(2024, 1, 1, 12, 2, 22, 0, time.UTC)
	clk := clock.NewMock()
	clk.Set(startTime)

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(startTime, clk)
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
	t.Logf("Start time:      %s", startTime.Format("15:04:05"))
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

// TestUnalignedTickerJitterBehavior shows UnalignedTicker behavior with jitter.
// Unlike RollingTicker, UnalignedTicker uses a fixed interval ticker internally,
// so jitter only adds delay but doesn't cause cumulative drift.
func TestUnalignedTickerJitterBehavior(t *testing.T) {
	interval := 60 * time.Second
	jitter := 10 * time.Second
	offset := 0 * time.Second

	clk := clock.NewMock()
	startTime := clk.Now()

	ticker := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start(clk)
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
	t.Logf("Start time:      %s", startTime.Format("15:04:05"))
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
	require.Less(t, drift, 1*time.Minute,
		"UnalignedTicker should have minimal drift due to fixed internal ticker")
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
	since := clk.Now()

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	printDist(dist)
	require.Less(t, 350, dist.Count)
	require.True(t, 9 < dist.Mean() && dist.Mean() < 11)
}

func TestAlignedTickerDistributionWithOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()
	since := clk.Now()

	ticker := &AlignedTicker{
		interval:    interval,
		jitter:      jitter,
		offset:      offset,
		minInterval: interval / 100,
	}
	ticker.start(since, clk)
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	printDist(dist)
	require.Less(t, 350, dist.Count)
	require.True(t, 9 < dist.Mean() && dist.Mean() < 11)
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

	ticker := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start(clk)
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	printDist(dist)
	require.Less(t, 350, dist.Count)
	require.True(t, 9 < dist.Mean() && dist.Mean() < 11)
}

func TestUnalignedTickerDistributionWithOffset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second
	offset := 3 * time.Second

	clk := clock.NewMock()

	ticker := &UnalignedTicker{
		interval: interval,
		jitter:   jitter,
		offset:   offset,
	}
	ticker.start(clk)
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	printDist(dist)
	require.Less(t, 350, dist.Count)
	require.True(t, 9 < dist.Mean() && dist.Mean() < 11)
}

// Simulates running the Ticker for an hour and displays stats about the
// operation.
func TestRollingTickerDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second

	clk := clock.NewMock()

	ticker := &RollingTicker{
		interval: interval,
		jitter:   jitter,
	}
	ticker.start(clk)
	defer ticker.Stop()
	dist := simulatedDist(ticker, clk)
	printDist(dist)
	require.Less(t, 275, dist.Count)
	require.True(t, 12 < dist.Mean() && 13 > dist.Mean())
}

type Distribution struct {
	Buckets  [60]int
	Count    int
	Waittime float64
}

func (d *Distribution) Mean() float64 {
	return d.Waittime / float64(d.Count)
}

func printDist(dist Distribution) {
	for i, count := range dist.Buckets {
		fmt.Printf("%2d %s\n", i, strings.Repeat("x", count))
	}
	fmt.Printf("Average interval: %f\n", dist.Mean())
	fmt.Printf("Count: %d\n", dist.Count)
}

func simulatedDist(ticker Ticker, clk *clock.Mock) Distribution {
	since := clk.Now()
	until := since.Add(1 * time.Hour)

	var dist Distribution

	last := clk.Now()
	for !clk.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			dist.Buckets[tm.Second()]++
			dist.Count++
			dist.Waittime += tm.Sub(last).Seconds()
			last = tm
		default:
			clk.Add(1 * time.Second)
		}
	}

	return dist
}
