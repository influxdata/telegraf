package dockerstate

import (
	"context"
	"encoding/json"
	"github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"net/http"
	"strings"
)

type Stats struct {
	Fields	map[string]interface{}
	Socket	string
}

var DefaultSocket = "/var/run/docker.sock"

func (s *Stats) Description() string {
	return "Basic Up/Down stats for docker containers"
}

func (s *Stats) SampleConfig() string {
	return `
  [inputs.docker_state]
	socket = /var/run/docker.sock
`
}

func (s *Stats) Gather(acc telegraf.Accumulator) error {
	c := http.Client{Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", s.Socket)
				},
			},
		 }
	var resp *http.Response
	var err error
	resp, err = c.Get("http://unix/containers/json?all=1&limit")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var containerobj interface{}
	if err := json.NewDecoder(resp.Body).Decode(&containerobj); err != nil {
		panic(err)
	}
	containers := containerobj.([]interface{})
	for _, v := range containers {
		v := v.(map[string]interface{})
		var state int
		if v["State"].(string) == "running" {
			state = 1
		} else {
			state = 0
		}
		container := strings.Split(v["Image"].(string), ":")
		container_ver := "latest"
		if len(container) > 1 {
			container_ver = container[1]
		}
		tags := map[string]string{
			"container": container[0],
			"version": container_ver,
			"status": v["Status"].(string),
		}
		fields := map[string]interface{}{
			"state": state,
			"created": v["Created"],
		}
		acc.AddGauge("dockerstate", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("dockerstate", func() telegraf.Input {
		return &Stats{
			Fields: make(map[string]interface{}),
			Socket: DefaultSocket,
		}
	})
}
