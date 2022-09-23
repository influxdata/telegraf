//go:generate ../../../tools/readme_config_includer/generator
package internet_speed

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/showwin/speedtest-go/speedtest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// InternetSpeed is used to store configuration values.
type InternetSpeed struct {
	EnableFileDownload bool            `toml:"enable_file_download" deprecated:"1.25.0;use 'memory_saving_mode' instead"`
	MemorySavingMode   bool            `toml:"memory_saving_mode"`
	Cache              bool            `toml:"cache"`
	Log                telegraf.Logger `toml:"-"`
	serverCache        *speedtest.Server
}

const measurement = "internet_speed"

func (*InternetSpeed) SampleConfig() string {
	return sampleConfig
}

func (is *InternetSpeed) Init() error {
	is.MemorySavingMode = is.MemorySavingMode || is.EnableFileDownload

	return nil
}

func (is *InternetSpeed) Gather(acc telegraf.Accumulator) error {
	// Get closest server
	s := is.serverCache
	if s == nil {
		user, err := speedtest.FetchUserInfo()
		if err != nil {
			return fmt.Errorf("fetching user info failed: %v", err)
		}
		serverList, err := speedtest.FetchServers(user)
		if err != nil {
			return fmt.Errorf("fetching server list failed: %v", err)
		}
		if len(serverList) < 1 {
			return fmt.Errorf("no servers found")
		}
		s = serverList[0]
		is.Log.Debugf("Found server: %v", s)
		if is.Cache {
			is.serverCache = s
		}
	}

	is.Log.Debug("Starting Speed Test")
	is.Log.Debug("Running Ping...")
	err := s.PingTest()
	if err != nil {
		return fmt.Errorf("ping test failed: %v", err)
	}
	is.Log.Debug("Running Download...")
	err = s.DownloadTest(is.MemorySavingMode)
	if err != nil {
		is.Log.Debug("try `memory_saving_mode = true` if this fails consistently")
		return fmt.Errorf("download test failed: %v", err)
	}
	is.Log.Debug("Running Upload...")
	err = s.UploadTest(is.MemorySavingMode)
	if err != nil {
		is.Log.Debug("try `memory_saving_mode = true` if this fails consistently")
		return fmt.Errorf("upload test failed failed: %v", err)
	}

	is.Log.Debug("Test finished.")

	fields := make(map[string]interface{})
	fields["download"] = s.DLSpeed
	fields["upload"] = s.ULSpeed
	fields["latency"] = timeDurationMillisecondToFloat64(s.Latency)

	tags := make(map[string]string)

	acc.AddFields(measurement, fields, tags)
	return nil
}

func init() {
	inputs.Add("internet_speed", func() telegraf.Input {
		return &InternetSpeed{}
	})
}

func timeDurationMillisecondToFloat64(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
