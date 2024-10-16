package ratelimiter

import (
	"errors"
	"math"
	"time"

	"github.com/influxdata/telegraf/config"
)

var (
	ErrLimitExceeded = errors.New("not enough tokens")
)

type RateLimitConfig struct {
	Limit  config.Size     `toml:"rate_limit"`
	Period config.Duration `toml:"rate_limit_period"`
}

func (cfg *RateLimitConfig) CreateRateLimiter() *RateLimiter {
	return &RateLimiter{
		limit:  int64(cfg.Limit),
		period: time.Duration(cfg.Period),
	}
}

type RateLimiter struct {
	limit       int64
	period      time.Duration
	periodStart time.Time
	remaining   int64
}

func (r *RateLimiter) Remaining(t time.Time) int64 {
	if r.limit == 0 {
		return math.MaxInt64
	}

	// Check for corner case
	if !r.periodStart.Before(t) {
		return 0
	}

	// We are in a new period, so the complete size is available
	deltat := t.Sub(r.periodStart)
	if deltat >= r.period {
		return r.limit
	}

	return r.remaining
}

func (r *RateLimiter) Accept(t time.Time, used int64) {
	if r.limit == 0 || r.periodStart.After(t) {
		return
	}

	// Remember the first query and reset if we are in a new period
	if r.periodStart.IsZero() {
		r.periodStart = t
		r.remaining = r.limit
	} else if deltat := t.Sub(r.periodStart); deltat >= r.period {
		r.periodStart = r.periodStart.Add(deltat.Truncate(r.period))
		r.remaining = r.limit
	}

	// Update the state
	r.remaining = max(r.remaining-used, 0)
}

func (r *RateLimiter) Undo(t time.Time, used int64) {
	// Do nothing if we are not in the current period or unlimited because we
	// already reset the limit on a new window.
	if r.limit == 0 || r.periodStart.IsZero() || r.periodStart.After(t) || t.Sub(r.periodStart) >= r.period {
		return
	}

	// Undo the state update
	r.remaining = min(r.remaining+used, r.limit)
}
