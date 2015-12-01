package trig

import (
	"math"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

func TestTrig(t *testing.T) {
	s := &Trig{
		Amplitude: 10.0,
	}

	for i := 0.0; i < 10.0; i++ {

		var acc testutil.Accumulator

		sine := math.Sin((i*math.Pi)/5.0) * s.Amplitude
		cosine := math.Cos((i*math.Pi)/5.0) * s.Amplitude

		s.Gather(&acc)

		fields := make(map[string]interface{})
		fields["sine"] = sine
		fields["cosine"] = cosine

		assert.True(t, acc.CheckFieldsValue("trig", fields))

	}
}
