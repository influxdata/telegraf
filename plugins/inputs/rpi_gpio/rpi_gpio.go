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
	gpio GPIO
}

type GPIO interface {
	Open() error
	Close() error
	ReadPin(pin_num int) int
}

type RPIO struct {
}

func (s *RPIO) Open() error {
	return rpio.Open()
}

func (s *RPIO) Close() error {
	return rpio.Close()
}

func (s *RPIO) ReadPin(pin_num int) int {
	pin := rpio.Pin(pin_num)
	pin.Input()
	val := pin.Read()
	return int(val)
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
	err := s.gpio.Open()
	if err != nil {
		return fmt.Errorf("error opening GPIO connection: %s", err)
	}
	defer s.gpio.Close()

	for field, pin_num := range s.Pins {
		s.Log.Debugf("Reading %s from pin %d", field, pin_num)
		val := s.gpio.ReadPin(pin_num)
		s.Log.Debugf("%s=%v (%T)", field, val, val)
		fields[field] = val
	}

	s.Log.Debugf("Fields: %s", fields)
	acc.AddFields("gpio", fields, tags)
	return nil
}

func init() {
	inputs.Add("rpi_gpio", func() telegraf.Input { return &RPiGPIO{gpio: &RPIO{}} })
}
