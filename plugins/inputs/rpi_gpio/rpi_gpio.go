package rpi_gpio

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/stianeikeland/go-rpio/v4"
)

type RPiGPIO struct {
	Pins map[string]int  `toml:pins`
	Log  telegraf.Logger `toml:"-"`
}

func (s *RPiGPIO) SampleConfig() string {
	return `
  ## Provide a data field name to gpio pin numbers
  ## Numbers correspond to the GPIO number, not the physical pin number
  [inputs.rpi_gpio.pins]
  button = 2
  motion_sensor = 3
  light_sensor = 4
	`
}

func (s *RPiGPIO) Description() string {
	return "Reads binary values from the GPIO pins of a RaspberryPi"
}

func (s *RPiGPIO) Gather(acc telegraf.Accumulator) error {

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	s.Log.Debugf("Opening GPIO connections")
	err := rpio.Open()
	if err != nil {
		return fmt.Errorf("error opening GPIO connection: %s", err)
	}
	for field, pin_num := range s.Pins {
		s.Log.Debugf("Reading %s from pin %d", field, pin_num)
		pin := rpio.Pin(pin_num)
		pin.Input()
		val := pin.Read()
		s.Log.Debugf("%s=%v (%T)", field, val, val)
		fields[field] = int(val)
	}
	s.Log.Debugf("Closing GPIO connections")
	rpio.Close()

	s.Log.Debugf("Fields: %s", fields)
	acc.AddFields("gpio", fields, tags)
	return nil
}

func init() {
	inputs.Add("rpi_gpio", func() telegraf.Input { return &RPiGPIO{} })
}
