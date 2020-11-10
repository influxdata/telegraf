package csgo

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/james4k/rcon"
)

type CSGO struct {
	Servers [][]string `toml:"servers"`
}

func (_ *CSGO) Description() string {
	return "Fetch metrics from a CSGO SRCDS"
}

var sampleConfig = `
  ## specify servers using the following format:
  ##    servers = [
  ##      ["ip1:port1", "rcon_password1"],
  ##      ["ip2:port2", "rcon_password2"],
  ##    ]
  #
  ## If no servers are specified, no data will be collected
  servers = []
`

func (_ *CSGO) SampleConfig() string {
	return sampleConfig
}

func (s *CSGO) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range s.Servers {
		wg.Add(1)
		go func(ss []string) {
			defer wg.Done()
			acc.AddError(s.gatherServer(ss, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func init() {
	inputs.Add("csgo", func() telegraf.Input {
		return &CSGO{}
	})
}

func (s *CSGO) gatherServer(server []string, acc telegraf.Accumulator) error {
	if len(server) != 2 {
		return errors.New("wrong argument length")
	}

	url, rconPw := server[0], server[1]
	remoteConsole, err := rcon.Dial(url, rconPw)
	if err != nil {
		return err
	}
	defer remoteConsole.Close()

	reqId, err := remoteConsole.Write("stats")
	resp, respReqId, err := remoteConsole.Read()
	if err != nil {
		return err
	} else if reqId != respReqId {
		return errors.New("response/request mismatch")
	}

	rows := strings.Split(resp, "\n")
	if len(rows) < 2 {
		return errors.New("bad response")
	}

	fields := strings.Fields(rows[1])
	cpu, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return err
	}
	netIn, err := strconv.ParseFloat(fields[1], 32)
	if err != nil {
		return err
	}
	netOut, err := strconv.ParseFloat(fields[2], 32)
	if err != nil {
		return err
	}
	uptimeMinutes, err := strconv.ParseFloat(fields[3], 32)
	if err != nil {
		return err
	}
	maps, err := strconv.ParseFloat(fields[4], 32)
	if err != nil {
		return err
	}
	fps, err := strconv.ParseFloat(fields[5], 32)
	if err != nil {
		return err
	}
	players, err := strconv.ParseFloat(fields[6], 32)
	if err != nil {
		return err
	}
	svms, err := strconv.ParseFloat(fields[7], 32)
	if err != nil {
		return err
	}
	msVar, err := strconv.ParseFloat(fields[8], 32)
	if err != nil {
		return err
	}
	tick, err := strconv.ParseFloat(fields[9], 32)
	if err != nil {
		return err
	}

	now := time.Now()
	fieldsG := map[string]interface{}{
		"csgo_cpu":            cpu,
		"csgo_net_in":         netIn,
		"csgo_net_out":        netOut,
		"csgo_uptime_minutes": uptimeMinutes,
		"csgo_maps":           maps,
		"csgo_fps":            fps,
		"csgo_players":        players,
		"csgo_svms":           svms,
		"csgo_ms_var":         msVar,
		"csgo_tick":           tick,
	}

	tags := map[string]string{
		"host": url,
	}
	acc.AddGauge("csgo", fieldsG, tags, now)
	return nil
}
