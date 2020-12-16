package random_number

import (
	"fmt"
	"math/rand"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// RandomNumber struct should be named the same as the Plugin
type RandomNumber struct {
	Min int
	Max int
}

// Description will appear directly above the plugin definition in the config file
func (r *RandomNumber) Description() string {
	return `This is a random number generator plugin which takes in a min and a max`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (r *RandomNumber) SampleConfig() string {
	return `  Min = 1 Max = 10000`
}

// Init can be implemented to do one-time processing stuff like initializing variables
// func (r *RandomNumber) Init() error {
// r.randomNumber = rand.Int()
// 	return nil
// }

// Gather defines what data the plugin will gather.
func (r *RandomNumber) Gather(acc telegraf.Accumulator) error {
	randomNumber := rand.Intn(r.Max-r.Min) + r.Min
	fmt.Println("random", randomNumber)
	acc.AddFields("state", map[string]interface{}{"value": randomNumber}, nil)
	return nil
}

func init() {
	inputs.Add("random_number", func() telegraf.Input { return &RandomNumber{Min: 1, Max: 100000} })
}
