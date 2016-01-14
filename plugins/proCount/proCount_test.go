package proCount

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
)

func TestProcs(t *testing.T) {
	s := &proCount {
		count: 200,
	}
	
	var acc testutil.Accumulator

/*
	I don't know how I can test to check that the 
	number of processes running is correct without 
	running the same command I ran within the plugin
	and that would only give the same result and not
	prove anything. I can't iterate through a set of
	integer values and show that it won't have errors
	accepting ints but that still doesn't prove that 
	number of processes listed is correct.
	

	fieldData := ???

	fields := make(map[string]interface{})
	fields["Processes"] = fieldData

	assert.True(t, acc.CheckFieldsValue("processes", fields)) 
*/

}
