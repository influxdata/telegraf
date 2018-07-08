// +build !freebsd

package nats_streaming

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"

	nss "github.com/nats-io/nats-streaming-server/server"
)

type Nats struct {
	Server          string
	ResponseTimeout internal.Duration

	client *http.Client
}

var sampleConfig = `
  ## The address of the monitoring endpoint of the NATS streaming server
  server = "http://localhost:8222"

  ## Maximum time to receive response
  # response_timeout = "5s"
`

func (n *Nats) SampleConfig() string {
	return sampleConfig
}

func (n *Nats) Description() string {
	return "Provides metrics about the performance of the NATS streaming server"
}

func (n *Nats) Gather(acc telegraf.Accumulator) error {
	err := n.gatherServer(acc)
	if err != nil {
		return err
	}
	err = n.gatherChannels(acc)
	if err != nil {
		return err
	}
	return nil
}

func (n *Nats) gatherServer(acc telegraf.Accumulator) error {
	url, err := url.Parse(n.Server)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, nss.ServerPath)

	if n.client == nil {
		n.client = n.createHTTPClient()
	}
	resp, err := n.client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	stats := new(nss.Serverz)
	err = json.Unmarshal([]byte(bytes), &stats)
	if err != nil {
		return err
	}

	acc.AddFields("nats_streaming_server",
		map[string]interface{}{
			"clients":       stats.Clients,
			"subscriptions": stats.Subscriptions,
			"channels":      stats.Channels,
			"total_msgs":    stats.TotalMsgs,
			"total_bytes":   stats.TotalBytes,
			"uptime":        stats.Now.Sub(stats.Start).Nanoseconds(),
		},
		map[string]string{
			"server":     n.Server,
			"cluster_id": stats.ClusterID,
			"server_id":  stats.ServerID,
		},
		time.Now())

	return nil
}

func (n *Nats) gatherChannels(acc telegraf.Accumulator) error {
	url, err := url.Parse(n.Server)
	if err != nil {
		return err
	}
	url.Path = path.Join(url.Path, nss.ChannelsPath)
	url.RawQuery = "subs=1"

	if n.client == nil {
		n.client = n.createHTTPClient()
	}
	resp, err := n.client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	stats := new(nss.Channelsz)
	err = json.Unmarshal([]byte(bytes), &stats)
	if err != nil {
		return err
	}
	now := time.Now()

	for _, channel := range stats.Channels {
		acc.AddFields("nats_streaming_channel",
			map[string]interface{}{
				"msgs":      channel.Msgs,
				"bytes":     channel.Bytes,
				"first_seq": channel.FirstSeq,
				"last_seq":  channel.LastSeq,
			},
			map[string]string{
				"server":       n.Server,
				"cluster_id":   stats.ClusterID,
				"server_id":    stats.ServerID,
				"channel_name": channel.Name,
			},
			now)

		for _, sub := range channel.Subscriptions {
			acc.AddFields("nats_streaming_subscription",
				map[string]interface{}{
					"is_durable":    sub.IsDurable,
					"is_offline":    sub.IsOffline,
					"max_inflight":  sub.MaxInflight,
					"ack_wait":      sub.AckWait,
					"last_sent":     sub.LastSent,
					"pending_count": sub.PendingCount,
					"is_stalled":    sub.IsStalled,
				},
				map[string]string{
					"server":       n.Server,
					"cluster_id":   stats.ClusterID,
					"server_id":    stats.ServerID,
					"channel_name": channel.Name,
					"client_id":    sub.ClientID,
					"inbox":        sub.Inbox,
					"ack_inbox":    sub.AckInbox,
					"durable_name": sub.DurableName,
					"queue_name":   sub.QueueName,
				},
				now)
		}
	}

	return nil
}

func (n *Nats) createHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	timeout := n.ResponseTimeout.Duration
	if timeout == time.Duration(0) {
		timeout = 5 * time.Second
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func init() {
	inputs.Add("nats_streaming", func() telegraf.Input {
		return &Nats{
			Server: "http://localhost:8222",
		}
	})
}
