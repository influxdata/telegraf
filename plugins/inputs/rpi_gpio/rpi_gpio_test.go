package rpi_gpio

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

type MockGPIO struct {
	MockData map[int]int
}

func (s *MockGPIO) Open() error {
	return nil
}

func (s *MockGPIO) Close() error {
	return nil
}

func (s *MockGPIO) ReadPin(pin_num int) int {
	return int(s.MockData[pin_num])
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	// Create an instance of the RPiGPIOPlugin
	p := RPiGPIO{
		// Specify pin mapping for this test
		// Field gpio2 will read from pin GPIO02
		// Field gpio3 will read from pin GPIO03
		Pins: map[string]int{
			"gpio2": 2,
			"gpio3": 3,
		},
		Log: testutil.Logger{},
	}

	// Replace the gpio interface with a mock implementation
	gpio = &MockGPIO{
		// Provide mock data readings for this test
		// GPIO02 will read LOW (0)
		// GPIO03 will read HIGH (1)
		MockData: map[int]int{
			2: 0,
			3: 1,
		},
	}

	// Let the plugin gather readings
	err := p.Gather(&acc)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the gathered readings match the mocked data
	fields := map[string]interface{}{
		"gpio2": 0,
		"gpio3": 1,
	}
	acc.AssertContainsFields(t, "gpio", fields)

}
