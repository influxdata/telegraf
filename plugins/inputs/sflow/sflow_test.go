package sflow

import (
	"testing"
)

// test writing to listener and then metrics are generated

func Test_stochasicPacketGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")

		// run for a specified period of time, maybe 10 minutes?

		// obnjective is to check there are no slice overruns or deadlocks etc.
	}

}
