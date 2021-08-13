package internetSpeedMonitor

import (
	"math"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedData struct {
	Latency  float64
	Download float64
	Upload   float64
}

type InternetSpeedMonitor struct {
	Data  SpeedData
	Error error
}

func testInternetSpeed(c chan InternetSpeedMonitor, enableFileDownload bool, logger telegraf.Logger) {
	delimeter := 1000.00

	user, error := speedtest.FetchUserInfo()
	if error != nil {
		c <- InternetSpeedMonitor{Error: error}
	}
	serverList, error := speedtest.FetchServerList(user)
	if error != nil {
		c <- InternetSpeedMonitor{Error: error}
	}
	targets, error := serverList.FindServer([]int{})

	if error != nil {
		c <- InternetSpeedMonitor{Error: error}
	}

	s := targets[0]

	logger.Info("Starting test...")
	logger.Info("Running Ping...")
	s.PingTest()
	logger.Info("Running Download...")
	s.DownloadTest(enableFileDownload)
	logger.Info("Running Upload...")
	s.UploadTest(enableFileDownload)
	logger.Info("Test finished.")

	c <- InternetSpeedMonitor{Data: SpeedData{
		Latency:  (math.Round(float64(s.Latency)/float64(time.Millisecond)*delimeter) / delimeter),
		Download: (math.Round(s.DLSpeed*delimeter) / delimeter),
		Upload:   (math.Round(s.ULSpeed*delimeter) / delimeter),
	}}

}
