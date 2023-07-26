//go:generate ../../../tools/readme_config_includer/generator
package internet_speed

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/showwin/speedtest-go/speedtest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// InternetSpeed is used to store configuration values.
type InternetSpeed struct {
	ServerIDInclude    []string `toml:"server_id_include"`
	ServerIDExclude    []string `toml:"server_id_exclude"`
	EnableFileDownload bool     `toml:"enable_file_download" deprecated:"1.25.0;use 'memory_saving_mode' instead"`
	MemorySavingMode   bool     `toml:"memory_saving_mode"`
	Cache              bool     `toml:"cache"`
	Connections        int      `toml:"connections"`
	TestMode           string   `toml:"test_mode"`

	Log telegraf.Logger `toml:"-"`

	server       *speedtest.Server // The main(best) server
	servers      speedtest.Servers // Auxiliary servers
	serverFilter filter.Filter
}

const (
	measurement    = "internet_speed"
	testModeSingle = "single"
	testModeMulti  = "multi"
)

func (*InternetSpeed) SampleConfig() string {
	return sampleConfig
}

func (is *InternetSpeed) Init() error {
	switch is.TestMode {
	case testModeSingle, testModeMulti:
	case "":
		is.TestMode = testModeSingle
	default:
		return fmt.Errorf("unrecognized test mode: %q", is.TestMode)
	}

	is.MemorySavingMode = is.MemorySavingMode || is.EnableFileDownload

	var err error
	is.serverFilter, err = filter.NewIncludeExcludeFilterDefaults(is.ServerIDInclude, is.ServerIDExclude, false, false)
	if err != nil {
		return fmt.Errorf("error compiling server ID filters: %w", err)
	}

	return nil
}

func (is *InternetSpeed) Gather(acc telegraf.Accumulator) error {
	// If not caching, go find the closest server each time.
	// We will find the best server as the main server. And
	// the remaining servers will be auxiliary candidates.
	if !is.Cache || is.server == nil {
		if err := is.findClosestServer(); err != nil {
			return fmt.Errorf("unable to find closest server: %w", err)
		}
	}

	err := is.server.PingTest(nil)
	if err != nil {
		return fmt.Errorf("ping test failed: %w", err)
	}

	if is.TestMode == testModeMulti {
		err = is.server.MultiDownloadTestContext(context.Background(), is.servers)
		if err != nil {
			return fmt.Errorf("download test failed: %w", err)
		}
		err = is.server.MultiUploadTestContext(context.Background(), is.servers)
		if err != nil {
			return fmt.Errorf("upload test failed failed: %w", err)
		}
	} else {
		err = is.server.DownloadTest()
		if err != nil {
			return fmt.Errorf("download test failed: %w", err)
		}
		err = is.server.UploadTest()
		if err != nil {
			return fmt.Errorf("upload test failed failed: %w", err)
		}
	}

	fields := map[string]any{
		"download": is.server.DLSpeed,
		"upload":   is.server.ULSpeed,
		"latency":  timeDurationMillisecondToFloat64(is.server.Latency),
		"jitter":   timeDurationMillisecondToFloat64(is.server.Jitter),
		"location": is.server.Name,
	}
	tags := map[string]string{
		"server_id": is.server.ID,
		"source":    is.server.Host,
		"test_mode": is.TestMode,
	}
	// Recycle the history of each test to prevent data backlog.
	is.server.Context.Reset()
	acc.AddFields(measurement, fields, tags)
	return nil
}

func (is *InternetSpeed) findClosestServer() error {
	client := speedtest.New(speedtest.WithUserConfig(&speedtest.UserConfig{
		UserAgent:  internal.ProductToken(),
		ICMP:       os.Geteuid() == 0 || os.Geteuid() == -1,
		SavingMode: is.MemorySavingMode,
	}))
	if is.Connections > 0 {
		client.SetNThread(is.Connections)
	}

	var err error
	is.servers, err = client.FetchServers()
	if err != nil {
		return fmt.Errorf("fetching server list failed: %w", err)
	}

	if len(is.servers) < 1 {
		return fmt.Errorf("no servers found")
	}

	// Return the first match or the server with the lowest latency
	// when filter mismatch all servers.
	var min int64 = math.MaxInt64
	selectIndex := -1
	for index, server := range is.servers {
		if is.serverFilter.Match(server.ID) {
			selectIndex = index
			break
		}
		if server.Latency > 0 {
			if min > server.Latency.Milliseconds() {
				min = server.Latency.Milliseconds()
				selectIndex = index
			}
		}
	}

	if selectIndex != -1 {
		is.server = is.servers[selectIndex]
		is.Log.Debugf("using server %s in %s (%s)\n", is.server.ID, is.server.Name, is.server.Host)
		return nil
	}

	return fmt.Errorf("no server set: filter excluded all servers or no available server found")
}

func init() {
	inputs.Add("internet_speed", func() telegraf.Input {
		return &InternetSpeed{}
	})
}

func timeDurationMillisecondToFloat64(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
