package azure_iothub

// azure_iothub.go

import (
	"context"
	"fmt"
	"log"
	"sync"

	iothub "github.com/amenzhinsky/iothub/iotservice"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Iothub struct {
	Client              iothub.Client
	UseGateway          bool   `toml:"use_gateway"`
	ConnectionString    string `toml:"connection_string"`
	HubName             string `toml:"hub_name"`
	DeviceID            string `toml:"device_id"`
	ModuleID            string `toml:"module_id"`
	SharedAccessKey     string `toml:"shared_access_key"`
	SharedAccessKeyName string `toml:"shared_access_key_name"`

	wg     *sync.WaitGroup
	cancel context.CancelFunc

	parser parsers.Parser
}

func (i *Iothub) Description() string {
	return "output plugin for Azure IoT Hub."
}

func (i *Iothub) SampleConfig() string {
	return `
## One of the following sets required for config:
#  
#  # 1.
#  connection_string = ""
#
#  # 2.
#  hub_name = ""
#  device_id = ""
#  shared_access_key = ""
#  
#  # 3.
#  hub_name = ""
#  shared_access_key = ""
#  shared_access_key_name = ""
#
## Optional:
#  module_id = ""
#  use_gateway = false
`
}

func (i *Iothub) hasConnectionString() bool {

	if len(i.ConnectionString) > 0 {
		return true
	}

	return false
}

func (i *Iothub) hasHubName() bool {

	if len(i.HubName) > 0 {
		return true
	}

	return false
}

func (i *Iothub) hasSharedAccessKey() bool {

	if len(i.SharedAccessKey) > 0 {
		return true
	}

	return false
}

func (i *Iothub) hasSharedAccessKeyName() bool {

	if len(i.SharedAccessKeyName) > 0 {
		return true
	}

	return false
}

func (i *Iothub) hasDeviceID() bool {

	if len(i.DeviceID) > 0 {
		return true
	}

	return false
}

func (i *Iothub) hasModuleID() bool {

	if len(i.ModuleID) > 0 {
		return true
	}

	return false
}

func (i *Iothub) createConnectionString() {
	conn := fmt.Sprintf("HostName=%s", i.HubName)

	if i.hasDeviceID() {
		conn = fmt.Sprintf("%s;DeviceId=%s", conn, i.DeviceID)
	}

	if i.hasModuleID() {
		conn = fmt.Sprintf("%s;ModuleId=%s", conn, i.ModuleID)
	}

	if i.hasSharedAccessKeyName() {
		conn = fmt.Sprintf("%s;SharedAccessKeyName=%s", conn, i.SharedAccessKeyName)
	}

	if i.hasSharedAccessKey() {
		conn = fmt.Sprintf("%s;SharedAccessKey=%s", conn, i.SharedAccessKey)
	}

	i.ConnectionString = conn
}

func (i *Iothub) validateConfiguration() error {
	valid := false

	// connection_string provided
	if i.hasConnectionString() {
		valid = true
	}

	// hub_name, shared_access_key, and shared_access_key_name provided
	if i.hasHubName() && i.hasSharedAccessKey() && i.hasSharedAccessKeyName() {
		valid = true
	}

	// hub_name, shared_access_key, and device_id provided
	if i.hasHubName() && i.hasSharedAccessKey() && i.hasDeviceID() {
		valid = true
	}

	// return
	if valid {
		return nil
	} else {
		return fmt.Errorf("invalid plugin configuration")
	}
}

// Init IoT Hub
func (i *Iothub) Init() error {
	// check for a valid configuration
	err := i.validateConfiguration()

	// if there's no explict connection string given
	if !i.hasConnectionString() {
		// create connection string from IoT Hub configuration
		i.createConnectionString()
	}

	// create a new client from connection string
	c, err := iothub.NewFromConnectionString(
		i.ConnectionString,
	)

	// set IoT Hub client
	i.Client = *c

	return err
}

func (i *Iothub) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (i *Iothub) Start(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	i.cancel = cancel

	i.Client.SubscribeEvents(ctx, func(msg *iothub.Event) error {
		//fmt.Printf("%q sends %q", msg.ConnectionDeviceID, msg.Payload)
		log.Printf("%q sends %q", msg.ConnectionDeviceID, msg.Payload)
		//i.process(ctx, msg, acc)
		return nil
	})

	return nil
}

func (i *Iothub) process(ctx context.Context, msg *iothub.Event, acc telegraf.Accumulator) error {
	metrics, err := i.parser.Parse(msg.Payload)

	for _, metric := range metrics {
		acc.AddMetric(metric)
	}

	return err
}

func (i *Iothub) Stop() {
	i.cancel()
	i.wg.Wait()
	i.Client.Close()
}

func init() {
	inputs.Add("azure_iothub", func() telegraf.Input { return &Iothub{} })
}
