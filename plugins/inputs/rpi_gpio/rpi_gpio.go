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
  ## Map data fields to pin numbers
  pins = {"motion_sensor"=1, "button"=2, "light_sensor"=3}
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
		s.Log.Debugf("%s=%d", field, val)
		fields[field] = val
	}
	s.Log.Debugf("Closing GPIO connections")
	rpio.Close()

	acc.AddFields("GPIO", fields, tags)
	return nil
}

func init() {
	inputs.Add("rpi_gpio", func() telegraf.Input { return &RPiGPIO{} })
}
