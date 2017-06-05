// +build windows

package open_hardware_monitor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateQueryWithSensors(t *testing.T) {
	//var acc testutil.Accumulator
	p := OpenHardwareMonitorConfig{
		SensorsType: []string{"Temperature", "Voltage"},
	}

	query, _ := p.CreateQuery()

	assert.Equal(t, "SELECT * FROM SENSOR WHERE SensorType='Temperature' OR SensorType='Voltage'", query)
}

func TestCreateQueryEmpty(t *testing.T) {
	//var acc testutil.Accumulator
	var p OpenHardwareMonitorConfig

	query, _ := p.CreateQuery()

	assert.Equal(t, "SELECT * FROM SENSOR", query)
}
