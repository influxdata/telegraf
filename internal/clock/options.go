package clock

import (
	"time"

	"github.com/benbjohnson/clock"
)

type config struct {
	clk   clock.Clock
	start time.Time
	align bool
}

type Option func(*config)

func WithClock(clk clock.Clock) Option {
	return func(c *config) {
		c.clk = clk
	}
}

func WithAlignment(start time.Time) Option {
	return func(c *config) {
		c.start = start
		c.align = true
	}
}
