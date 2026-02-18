package clock

import (
	"fmt"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
)

type distribution struct {
	buckets  [60]int
	count    int
	waittime float64
}

func simulatedDist(ticker Ticker, clk *clock.Mock) distribution {
	start := clk.Now()
	end := start.Add(1 * time.Hour)

	var dist distribution

	last := clk.Now()
	for !clk.Now().After(end) {
		select {
		case tm := <-ticker.Elapsed():
			dist.buckets[tm.Second()]++
			dist.count++
			dist.waittime += tm.Sub(last).Seconds()
			last = tm
		default:
			clk.Add(1 * time.Second)
		}
	}

	return dist
}

func (d *distribution) mean() float64 {
	return d.waittime / float64(d.count)
}

func (d distribution) print() {
	for i, count := range d.buckets {
		fmt.Printf("%2d %s\n", i, strings.Repeat("x", count))
	}
	fmt.Printf("Average interval: %f\n", d.mean())
	fmt.Printf("Count: %d\n", d.count)
}
