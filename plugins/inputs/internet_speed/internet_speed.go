package internet_speed

import (
	"fmt"
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/showwin/speedtest-go/speedtest"
)

// InternetSpeed is used to store configuration values.
type InternetSpeed struct {
	EnableFileDownload bool            `toml:"enable_file_download"`
	Log                telegraf.Logger `toml:"-"`
}

var InternetSpeedConfig = `
## Sets if runs file download test
## Default: false  
enable_file_download = false
`

// Description returns information about the plugin.
func (internetSpeed *InternetSpeed) Description() string {
	return "Monitors internet speed using speedtest.net service"
}

// SampleConfig displays configuration instructions.
func (internetSpeed *InternetSpeed) SampleConfig() string {
	return InternetSpeedConfig
}

const measurement = "internet_speed"
const delimeter = 1000.00

func (internetSpeed *InternetSpeed) Gather(acc telegraf.Accumulator) error {
	enableFileDownload := internetSpeed.EnableFileDownload
	log := internetSpeed.Log

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}
	serverList, err := speedtest.FetchServerList(user)
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}

	s := targets[0]

	log.Info("Starting Speed Test")
	log.Info("Running Ping...")
	err = s.PingTest()
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}
	log.Info("Running Download...")
	err = s.DownloadTest(enableFileDownload)
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}
	log.Info("Running Upload...")
	err = s.UploadTest(enableFileDownload)
	if err != nil {
		acc.AddError(err)
		return fmt.Errorf("gathering speed failed: %v", err)
	}

	log.Info("Test finished.")

	fields := make(map[string]interface{})
	fields["download"] = (math.Round(s.DLSpeed*delimeter) / delimeter)
	fields["upload"] = (math.Round(s.ULSpeed*delimeter) / delimeter)
	fields["latency"] = (math.Round(float64(s.Latency)/float64(time.Millisecond)*delimeter) / delimeter)

	tags := make(map[string]string)

	acc.AddFields(measurement, fields, tags)
	return nil
}
func init() {
	inputs.Add("internet_speed", func() telegraf.Input {
		return &InternetSpeed{
			EnableFileDownload: false,
		}
	})
}
