package azure_iothub_consumer

// azure_iothub_consumer.go

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	iotcommon "github.com/amenzhinsky/iothub/common"
	iothub "github.com/amenzhinsky/iothub/iotdevice"
	iotmqtt "github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/influxdata/telegraf/plugins/parsers"
)

// IothubConsumer
type IothubConsumer struct {
	Client              iothub.ModuleClient
	UseGateway          bool   `toml:"use_gateway"`
	ConnectionString    string `toml:"connection_string"`
	HubName             string `toml:"hub_name"`
	DeviceID            string `toml:"device_id"`
	ModuleID            string `toml:"module_id"`
	SharedAccessKey     string `toml:"shared_access_key"`
	SharedAccessKeyName string `toml:"shared_access_key_name"`

	sub *iothub.EventSub

	wg     *sync.WaitGroup
	cancel context.CancelFunc

	parser parsers.Parser
}

// Description -
func (i *IothubConsumer) Description() string {
	return "Input plugin for Azure IoT Hub Edge Module"
}

// SampleConfig -
func (i *IothubConsumer) SampleConfig() string {
	return `
## One of the following sets required for configuration:
#  
#  # 1.
#  connection_string = ""
#  use_gateway = true
#
#  # 2.
#  hub_name = ""
#  device_id = ""
#  module_id = ""
#  shared_access_key = ""
#  use_gateway = true
#
#  # 3.
#  Provide no configuration for IoT Edge module, and it will self-configure from environment variables present in edge modules.

`
}

func (i *IothubConsumer) hasConnectionString() bool {

	if len(strings.TrimSpace(i.ConnectionString)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) hasHubName() bool {

	if len(strings.TrimSpace(i.HubName)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) hasSharedAccessKey() bool {

	if len(strings.TrimSpace(i.SharedAccessKey)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) hasSharedAccessKeyName() bool {

	if len(strings.TrimSpace(i.SharedAccessKeyName)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) hasDeviceID() bool {

	if len(strings.TrimSpace(i.DeviceID)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) hasModuleID() bool {

	if len(strings.TrimSpace(i.ModuleID)) > 0 {
		return true
	}

	return false
}

func (i *IothubConsumer) createConnectionString() {
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

func (i *IothubConsumer) validateConfiguration() bool {
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

	return valid
}

func messageToMetric(msg *iotcommon.Message, acc telegraf.Accumulator) {
	acc.AddFields("azure_iothub_consumer", fields(msg), tags(msg), time.Now().UTC())
}

func tags(msg *iotcommon.Message) map[string]string {
	ts := map[string]string{}

	if len(msg.MessageID) > 0 {
		ts["MessageID"] = msg.MessageID
	}

	if len(msg.To) > 0 {
		ts["To"] = msg.To
	}

	if msg.ExpiryTime != nil && len(msg.ExpiryTime.String()) > 0 {
		ts["ExpiryTime"] = msg.ExpiryTime.String()
	}

	if msg.EnqueuedTime != nil && len(msg.EnqueuedTime.String()) > 0 {
		ts["EnqueuedTime"] = msg.EnqueuedTime.String()
	}

	if len(msg.UserID) > 0 {
		ts["UserID"] = msg.UserID
	}

	if len(msg.ConnectionDeviceID) > 0 {
		ts["ConnectionDeviceID"] = msg.ConnectionDeviceID
	}

	if len(msg.ConnectionDeviceGenerationID) > 0 {
		ts["ConnectionDeviceGenerationID"] = msg.ConnectionDeviceGenerationID
	}

	if len(msg.MessageSource) > 0 {
		ts["MessageSource"] = msg.MessageSource
	}

	if msg.Payload != nil && len(string(msg.Payload)) > 0 {
		ts["Payload"] = string(msg.Payload)
	}

	for k, v := range msg.Properties {
		ts[fmt.Sprintf("Properties.%s", k)] = v
	}

	return ts
}

func fields(msg *iotcommon.Message) map[string]interface{} {
	flds := make(map[string]interface{})

	flds["Source"] = msg.MessageSource
	flds["Body"] = msg.Payload

	return flds
}

// Init IoT Hub
func (i *IothubConsumer) Init() error {

	// check for a valid configuration
	valid := i.validateConfiguration()

	// if connection parameters supplied
	if valid {

		// if there's no explict connection string given
		if !i.hasConnectionString() {
			// create connection string from IoT Hub configuration
			i.createConnectionString()
		}

		// create a new client from connection string

		gwhn := os.Getenv("IOTEDGE_GATEWAYHOSTNAME")
		mgid := os.Getenv("IOTEDGE_MODULEGENERATIONID")
		wluri := os.Getenv("IOTEDGE_WORKLOADURI")

		c, err := iothub.NewModuleFromConnectionString(
			iotmqtt.NewModuleTransport(), i.ConnectionString, gwhn, mgid, wluri, true,
		)

		// set IoT Hub client
		i.Client = *c
		return err

	} else {

		// create from environment
		c, err := iothub.NewModuleFromEnvironment(
			iotmqtt.NewModuleTransport(), true,
		)

		// set IoT Hub client
		i.Client = *c
		return err
	}
}

// Gather is not implemented
func (i *IothubConsumer) Gather(acc telegraf.Accumulator) error {

	return nil
}

// Start the IothubConsumer
func (i *IothubConsumer) Start(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithCancel(context.Background())
	i.cancel = cancel
	i.wg = &sync.WaitGroup{}

	if err := i.Client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	sub, err := i.Client.SubscribeEvents(ctx)
	if err != nil {
		log.Fatal(err)
	}

	i.sub = sub

	i.wg.Add(1)
	go i.listen(acc)

	return nil
}

func (i *IothubConsumer) listen(acc telegraf.Accumulator) error {
	defer i.wg.Done()
	var err error
	msgs := i.sub.C()
	subErr := i.sub.Err()

	for err == nil {

		if subErr != nil {
			err = subErr
		}
		msg := <-msgs
		messageToMetric(msg, acc)

	}

	return err
}

// Stop the IotHubConsumer
func (i *IothubConsumer) Stop() {
	i.cancel()
	i.Client.Close()
	i.wg.Wait()
}

func init() {
	inputs.Add("azure_iothub_consumer", func() telegraf.Input { return &IothubConsumer{} })
}
