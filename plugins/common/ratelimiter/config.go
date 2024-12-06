package ratelimiter

import (
	"time"

	"github.com/influxdata/telegraf/config"
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
