package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	r := NewRateLimiter(5, time.Second)
	ticker := time.NewTicker(time.Millisecond * 75)

	// test that we can only get 5 receives from the rate limiter
	counter := 0
outer:
	for {
		select {
		case <-r.C:
			counter++
		case <-ticker.C:
			break outer
		}
	}

	assert.Equal(t, 5, counter)
	r.Stop()
	// verify that the Stop function closes the channel.
	_, ok := <-r.C
	assert.False(t, ok)
}

func TestRateLimiterMultipleIterations(t *testing.T) {
	r := NewRateLimiter(5, time.Millisecond*50)
	ticker := time.NewTicker(time.Millisecond * 250)

	// test that we can get 15 receives from the rate limiter
	counter := 0
outer:
	for {
		select {
		case <-ticker.C:
			break outer
		case <-r.C:
			counter++
		}
	}

	assert.True(t, counter > 10)
	r.Stop()
	// verify that the Stop function closes the channel.
	_, ok := <-r.C
	assert.False(t, ok)
}
