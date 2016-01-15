package proCount

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
)

func TestProcs(t *testing.T) {
	s := &proCount {
		count: 200,
	}

/*
	I don't know how I can test to check that the 
	number of processes running is correct without 
	running the same command I ran within the plugin
	and that would only give the same result and not
	prove anything. I can't iterate through a set of
	integer values and show that it won't have errors
	accepting ints but that still doesn't prove that 
	number of processes listed is correct. */

	for i:= 0; i < to; i++ {

		var acc testutil.Accumulator
		
	//This is where I don't know how to test it.  fieldData := 

		fields := make(map[string]interface{})
		fields["Processes"] = fieldData

		assert.True(t, acc.CheckFieldsValue("processes", fields)) 
	}
}
