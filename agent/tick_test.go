package agent

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
)

var format = "2006-01-02T15:04:05.999Z07:00"

func TestAlignedTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second

	clock := clock.NewMock()
	since := clock.Now()
	until := since.Add(60 * time.Second)

	ticker := newAlignedTicker(since, interval, jitter, clock)

	expected := []time.Time{
		time.Unix(10, 0).UTC(),
		time.Unix(20, 0).UTC(),
		time.Unix(30, 0).UTC(),
		time.Unix(40, 0).UTC(),
		time.Unix(50, 0).UTC(),
		time.Unix(60, 0).UTC(),
	}

	actual := []time.Time{}
	for !clock.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
		}
		clock.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

func TestAlignedTickerJitter(t *testing.T) {
	interval := 10 * time.Second
	jitter := 5 * time.Second

	clock := clock.NewMock()
	since := clock.Now()
	until := since.Add(60 * time.Second)

	ticker := newAlignedTicker(since, interval, jitter, clock)

	last := since
	for !clock.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			require.True(t, tm.Sub(last) <= 15*time.Second)
			require.True(t, tm.Sub(last) >= 5*time.Second)
			last = last.Add(interval)
		default:
		}
		clock.Add(5 * time.Second)
	}
}

func TestAlignedTickerMissedTick(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second

	clock := clock.NewMock()
	since := clock.Now()

	ticker := newAlignedTicker(since, interval, jitter, clock)

	clock.Add(25 * time.Second)
	tm := <-ticker.Elapsed()
	require.Equal(t, time.Unix(10, 0).UTC(), tm.UTC())
	clock.Add(5 * time.Second)
	tm = <-ticker.Elapsed()
	require.Equal(t, time.Unix(30, 0).UTC(), tm.UTC())
}

func TestUnalignedTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second

	clock := clock.NewMock()
	clock.Add(1 * time.Second)
	since := clock.Now()
	until := since.Add(60 * time.Second)

	ticker := newUnalignedTicker(interval, jitter, clock)

	expected := []time.Time{
		time.Unix(1, 0).UTC(),
		time.Unix(11, 0).UTC(),
		time.Unix(21, 0).UTC(),
		time.Unix(31, 0).UTC(),
		time.Unix(41, 0).UTC(),
		time.Unix(51, 0).UTC(),
		time.Unix(61, 0).UTC(),
	}

	actual := []time.Time{}
	for !clock.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
		}
		clock.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

func TestRollingTicker(t *testing.T) {
	interval := 10 * time.Second
	jitter := 0 * time.Second

	clock := clock.NewMock()
	clock.Add(1 * time.Second)
	since := clock.Now()
	until := since.Add(60 * time.Second)

	ticker := newUnalignedTicker(interval, jitter, clock)

	expected := []time.Time{
		time.Unix(1, 0).UTC(),
		time.Unix(11, 0).UTC(),
		time.Unix(21, 0).UTC(),
		time.Unix(31, 0).UTC(),
		time.Unix(41, 0).UTC(),
		time.Unix(51, 0).UTC(),
		time.Unix(61, 0).UTC(),
	}

	actual := []time.Time{}
	for !clock.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			actual = append(actual, tm.UTC())
		default:
		}
		clock.Add(10 * time.Second)
	}

	require.Equal(t, expected, actual)
}

// Simulates running the Ticker for an hour and displays stats about the
// operation.
func TestAlignedTickerDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	interval := 10 * time.Second
	jitter := 5 * time.Second

	clock := clock.NewMock()
	since := clock.Now()

	ticker := newAlignedTicker(since, interval, jitter, clock)
	dist := simulatedDist(ticker, clock)
	printDist(dist)
	require.True(t, 350 < dist.Count)
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

	clock := clock.NewMock()

	ticker := newUnalignedTicker(interval, jitter, clock)
	dist := simulatedDist(ticker, clock)
	printDist(dist)
	require.True(t, 350 < dist.Count)
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

	clock := clock.NewMock()

	ticker := newRollingTicker(interval, jitter, clock)
	dist := simulatedDist(ticker, clock)
	printDist(dist)
	require.True(t, 275 < dist.Count)
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

func simulatedDist(ticker Ticker, clock *clock.Mock) Distribution {
	since := clock.Now()
	until := since.Add(1 * time.Hour)

	var dist Distribution

	last := clock.Now()
	for !clock.Now().After(until) {
		select {
		case tm := <-ticker.Elapsed():
			dist.Buckets[tm.Second()] += 1
			dist.Count++
			dist.Waittime += tm.Sub(last).Seconds()
			last = tm
		default:
			clock.Add(1 * time.Second)
		}
	}

	return dist
}
