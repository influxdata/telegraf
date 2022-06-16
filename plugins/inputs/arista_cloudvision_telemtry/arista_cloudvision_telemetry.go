package arista_cloudvision_telemtry

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig string

// Cloudvision struct
type CVP struct {
	Cvpaddress    string         `toml:"addresses"`
	Subscriptions []Subscription `toml:"subscription"`

	Encoding string
	Origin   string
	Prefix   string

	Cvptoken string `toml:"cvptoken"`

	// Redial
	Redial config.Duration

	// Internal state
	internalAliases map[string]string
	acc             telegraf.Accumulator
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	// Lookup/device+name/key/value
	lookup      map[string]map[string]map[string]interface{}
	lookupMutex sync.Mutex

	Log telegraf.Logger
}

type Subscription struct {
	Name   string
	Origin string
	Path   string

	// Subscription mode and interval
	SubscriptionMode string          `toml:"subscription_mode"`
	SampleInterval   config.Duration `toml:"sample_interval"`

	// Duplicate suppression
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`
}

// Struct for cloudvision API to return device data.
type CvPDevices struct {
	Result struct {
		Value struct {
			Key struct {
				DeviceID string `json:"deviceId"`
			} `json:"key"`
			SoftwareVersion    string    `json:"softwareVersion"`
			ModelName          string    `json:"modelName"`
			HardwareRevision   string    `json:"hardwareRevision"`
			Fqdn               string    `json:"fqdn"`
			Hostname           string    `json:"hostname"`
			DomainName         string    `json:"domainName"`
			SystemMacAddress   string    `json:"systemMacAddress"`
			BootTime           time.Time `json:"bootTime"`
			StreamingStatus    string    `json:"streamingStatus"`
			ExtendedAttributes struct {
				FeatureEnabled struct {
					Danz bool `json:"Danz"`
					Mlag bool `json:"Mlag"`
				} `json:"featureEnabled"`
			} `json:"extendedAttributes"`
		} `json:"value"`
		Time time.Time `json:"time"`
		Type string    `json:"type"`
	} `json:"result"`
}

func (*CVP) SampleConfig() string {
	return sampleConfig
}

// Start the CVP gNMI telemetry service
func (c *CVP) Start(acc telegraf.Accumulator) error {
	//var err error
	//var ctx context.Context
	//var tlscfg *tls.Config
	//var request *gnmiLib.SubscribeRequest
	//c.acc = acc
	//ctx, c.cancel = context.WithCancel(context.Background())
	c.lookupMutex.Lock()
	c.lookup = make(map[string]map[string]map[string]interface{})
	c.lookupMutex.Unlock()

	cvdevs := make(map[string]string)

	for cvpdevice, devicetarget := range c.CvpDevices() {
		c.Log.Info("Connect to CVP and using Device ", cvpdevice, " With target of ", devicetarget)
		cvdevs[cvpdevice] = devicetarget
	}

	return nil

}

// Method to return all devices which are streaming so we can then use their targets as the gNMI target.
func (c *CVP) CvpDevices() map[string]string {
	var bearer = "Bearer " + c.Cvptoken
	//Connect to CVP resource api
	req, err := http.NewRequest("GET", "https://"+c.Cvpaddress+"/api/resources/inventory/v1/Device/all", nil)
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Accept", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		c.Log.Error("Cannot connect to CVP", err)
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Log.Error("Cannot marshall data", err)
	}

	f := strings.Split(string(responseData), "\n")
	//Create a map for devices
	devs := map[string]string{}
	//Loop through and add devices to devs map that are currently streaming.
	for _, i := range f {
		var Dev CvPDevices
		json.Unmarshal([]byte(i), &Dev)
		if Dev.Result.Value.StreamingStatus == "STREAMING_STATUS_ACTIVE" {
			devs[Dev.Result.Value.Fqdn] = Dev.Result.Value.Key.DeviceID
		}

	}
	//Return devices.
	return devs
}

// Stop listener and cleanup
func (c *CVP) Stop() {
	c.cancel()
	c.wg.Wait()
}

func New() telegraf.Input {
	return &CVP{
		Encoding: "proto",
		Redial:   config.Duration(10 * time.Second),
	}
}

// Gather plugin measurements (unused)
func (c *CVP) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("arista_cloudvision_telemtry", New)
}
