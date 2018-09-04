package appinsights

import (
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/clock/fakeclock"
)

var fakeClock *fakeclock.FakeClock

func mockClock(timestamp ...time.Time) {
	if len(timestamp) > 0 {
		fakeClock = fakeclock.NewFakeClock(timestamp[0])
	} else {
		fakeClock = fakeclock.NewFakeClock(time.Now().Round(time.Minute))
	}

	currentClock = fakeClock
}

func resetClock() {
	fakeClock = nil
	currentClock = clock.NewClock()
}

func slowTick(seconds int) {
	const delay = time.Millisecond * time.Duration(5)

	// Sleeps in tests are evil, but with all the async nonsense going
	// on, no callbacks, and minimal control of the clock, I'm not
	// really sure I have another choice.

	time.Sleep(delay)
	for i := 0; i < seconds; i++ {
		fakeClock.Increment(time.Second)
		time.Sleep(delay)
	}
}
