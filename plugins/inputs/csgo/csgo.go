package csgo

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/james4k/rcon"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type statsData struct {
	CPU           float64 `json:"cpu"`
	NetIn         float64 `json:"net_in"`
	NetOut        float64 `json:"net_out"`
	UptimeMinutes float64 `json:"uptime_minutes"`
	Maps          float64 `json:"maps"`
	FPS           float64 `json:"fps"`
	Players       float64 `json:"players"`
	Sim           float64 `json:"sv_ms"`
	Variance      float64 `json:"variance_ms"`
	Tick          float64 `json:"tick_ms"`
}

type CSGO struct {
	Servers [][]string `toml:"servers"`
}

func (s *CSGO) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range s.Servers {
		wg.Add(1)
		go func(ss []string) {
			defer wg.Done()
			acc.AddError(s.gatherServer(acc, ss, requestServer))
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

func (s *CSGO) gatherServer(
	acc telegraf.Accumulator,
	server []string,
	request func(string, string) (string, error),
) error {
	if len(server) != 2 {
		return errors.New("incorrect server config")
	}

	url, rconPw := server[0], server[1]
	resp, err := request(url, rconPw)
	if err != nil {
		return err
	}

	rows := strings.Split(resp, "\n")
	if len(rows) < 2 {
		return errors.New("bad response")
	}

	fields := strings.Fields(rows[1])
	if len(fields) != 10 {
		return errors.New("bad response")
	}

	cpu, err := strconv.ParseFloat(fields[0], 32)
	if err != nil {
		return err
	}
	netIn, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return err
	}
	netOut, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return err
	}
	uptimeMinutes, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return err
	}
	maps, err := strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return err
	}
	fps, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return err
	}
	players, err := strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return err
	}
	svms, err := strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return err
	}
	msVar, err := strconv.ParseFloat(fields[8], 64)
	if err != nil {
		return err
	}
	tick, err := strconv.ParseFloat(fields[9], 64)
	if err != nil {
		return err
	}

	now := time.Now()
	stats := statsData{
		CPU:           cpu,
		NetIn:         netIn,
		NetOut:        netOut,
		UptimeMinutes: uptimeMinutes,
		Maps:          maps,
		FPS:           fps,
		Players:       players,
		Sim:           svms,
		Variance:      msVar,
		Tick:          tick,
	}

	tags := map[string]string{
		"host": url,
	}

	var statsMap map[string]interface{}
	marshalled, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshalled, &statsMap)
	if err != nil {
		return err
	}

	acc.AddGauge("csgo", statsMap, tags, now)
	return nil
}

func requestServer(url string, rconPw string) (string, error) {
	remoteConsole, err := rcon.Dial(url, rconPw)
	if err != nil {
		return "", err
	}
	defer remoteConsole.Close()

	reqID, err := remoteConsole.Write("stats")
	if err != nil {
		return "", err
	}

	resp, respReqID, err := remoteConsole.Read()
	if err != nil {
		return "", err
	} else if reqID != respReqID {
		return "", errors.New("response/request mismatch")
	} else {
		return resp, nil
	}
}
