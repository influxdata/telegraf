package teamspeak

import (
	"github.com/multiplay/go-ts3"

	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Teamspeak struct {
	Server         string
	Username       string
	Password       string
	VirtualServers []int `toml:"virtual_servers"`

	client    *ts3.Client
	connected bool
}

func (ts *Teamspeak) Gather(acc telegraf.Accumulator) error {
	var err error

	if !ts.connected {
		ts.client, err = ts3.NewClient(ts.Server)
		if err != nil {
			return err
		}

		err = ts.client.Login(ts.Username, ts.Password)
		if err != nil {
			return err
		}

		ts.connected = true
	}

	for _, vserver := range ts.VirtualServers {
		if err := ts.client.Use(vserver); err != nil {
			ts.connected = false
			return err
		}

		sm, err := ts.client.Server.Info()
		if err != nil {
			ts.connected = false
			return err
		}

		sc, err := ts.client.Server.ServerConnectionInfo()
		if err != nil {
			ts.connected = false
			return err
		}

		tags := map[string]string{
			"virtual_server": strconv.Itoa(sm.ID),
			"name":           sm.Name,
		}

		fields := map[string]interface{}{
			"uptime":                 sm.Uptime,
			"clients_online":         sm.ClientsOnline,
			"total_ping":             sm.TotalPing,
			"total_packet_loss":      sm.TotalPacketLossTotal,
			"packets_sent_total":     sc.PacketsSentTotal,
			"packets_received_total": sc.PacketsReceivedTotal,
			"bytes_sent_total":       sc.BytesSentTotal,
			"bytes_received_total":   sc.BytesReceivedTotal,
			"query_clients_online":   sm.QueryClientsOnline,
		}

		acc.AddFields("teamspeak", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("teamspeak", func() telegraf.Input {
		return &Teamspeak{
			Server:         "127.0.0.1:10011",
			VirtualServers: []int{1},
		}
	})
}
