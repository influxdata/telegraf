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

	user, err := speedtest.FetchUserInfo()
	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}
	serverList, err := speedtest.FetchServerList(user)
	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}
	targets, err := serverList.FindServer([]int{})

	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}

	s := targets[0]

	logger.Info("Starting test...")
	logger.Info("Running Ping...")
	err = s.PingTest()
	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}
	logger.Info("Running Download...")
	err = s.DownloadTest(enableFileDownload)
	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}
	logger.Info("Running Upload...")
	err = s.UploadTest(enableFileDownload)
	if err != nil {
		c <- InternetSpeedMonitor{Error: err}
	}
	logger.Info("Test finished.")

	c <- InternetSpeedMonitor{Data: SpeedData{
		Latency:  (math.Round(float64(s.Latency)/float64(time.Millisecond)*delimeter) / delimeter),
		Download: (math.Round(s.DLSpeed*delimeter) / delimeter),
		Upload:   (math.Round(s.ULSpeed*delimeter) / delimeter),
	}}
}
