package internet_speed

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/showwin/speedtest-go/speedtest"
)

// InternetSpeed is used to store configuration values.
type InternetSpeed struct {
	EnableFileDownload bool            `toml:"enable_file_download"`
	Log                telegraf.Logger `toml:"-"`
}

const sampleConfig = `
  ## Sets if runs file download test
  ## Default: false  
  enable_file_download = false
`

// Description returns information about the plugin.
func (is *InternetSpeed) Description() string {
	return "Monitors internet speed using speedtest.net service"
}

// SampleConfig displays configuration instructions.
func (is *InternetSpeed) SampleConfig() string {
	return sampleConfig
}

const measurement = "internet_speed"

func (is *InternetSpeed) Gather(acc telegraf.Accumulator) error {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		return fmt.Errorf("fetching user info failed: %v", err)
	}
	serverList, err := speedtest.FetchServerList(user)
	if err != nil {
		return fmt.Errorf("fetching server list failed: %v", err)
	}

	if len(serverList.Servers) < 1 {
		return fmt.Errorf("no servers found")
	}
	s := serverList.Servers[0]
	is.Log.Debug("Starting Speed Test")
	is.Log.Debug("Running Ping...")
	err = s.PingTest()
	if err != nil {
		return fmt.Errorf("ping test failed: %v", err)
	}
	is.Log.Debug("Running Download...")
	err = s.DownloadTest(is.EnableFileDownload)
	if err != nil {
		return fmt.Errorf("download test failed: %v", err)
	}
	is.Log.Debug("Running Upload...")
	err = s.UploadTest(is.EnableFileDownload)
	if err != nil {
		return fmt.Errorf("upload test failed failed: %v", err)
	}

	is.Log.Debug("Test finished.")

	fields := make(map[string]interface{})
	fields["download"] = s.DLSpeed
	fields["upload"] = s.ULSpeed
	fields["latency"] = s.Latency

	tags := make(map[string]string)

	acc.AddFields(measurement, fields, tags)
	return nil
}
func init() {
	inputs.Add("internet_speed", func() telegraf.Input {
		return &InternetSpeed{}
	})
}
