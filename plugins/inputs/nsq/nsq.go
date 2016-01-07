// The MIT License (MIT)
//
// Copyright (c) 2015 Jeff Nickoloff (jeff@allingeek.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nsq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/influxdb/telegraf/plugins"
)

// Might add Lookupd endpoints for cluster discovery
type NSQ struct {
	Endpoints []string
}

var sampleConfig = `
  # An array of NSQD HTTP API endpoints
  endpoints = ["http://localhost:4151","http://otherhost:4151"]
`

const (
	requestPattern = `%s/stats?format=json`
)

func init() {
	plugins.Add("nsq", func() plugins.Plugin {
		return &NSQ{}
	})
}

func (n *NSQ) SampleConfig() string {
	return sampleConfig
}

func (n *NSQ) Description() string {
	return "Read NSQ topic and channel statistics."
}

func (n *NSQ) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup
	var outerr error

	for _, e := range n.Endpoints {
		wg.Add(1)
		go func(e string) {
			defer wg.Done()
			outerr = n.gatherEndpoint(e, acc)
		}(e)
	}

	wg.Wait()

	return outerr
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{Transport: tr}

func (n *NSQ) gatherEndpoint(e string, acc plugins.Accumulator) error {
	u, err := buildURL(e)
	if err != nil {
		return err
	}
	r, err := client.Get(u.String())
	if err != nil {
		return fmt.Errorf("Error while polling %s: %s", u.String(), err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", u.String(), r.Status)
	}

	s := &NSQStats{}
	err = json.NewDecoder(r.Body).Decode(s)
	if err != nil {
		return fmt.Errorf(`Error parsing response: %s`, err)
	}

	tags := map[string]string{
		`server_host`:    u.Host,
		`server_version`: s.Data.Version,
	}

	if s.Data.Health == `OK` {
		acc.Add(`nsq_server_count`, int64(1), tags)
	} else {
		acc.Add(`nsq_server_count`, int64(0), tags)
	}

	acc.Add(`nsq_server_topic_count`, int64(len(s.Data.Topics)), tags)
	for _, t := range s.Data.Topics {
		topicStats(t, acc, u.Host, s.Data.Version)
	}

	return nil
}

func buildURL(e string) (*url.URL, error) {
	u := fmt.Sprintf(requestPattern, e)
	addr, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse address '%s': %s", u, err)
	}
	return addr, nil
}

func topicStats(t TopicStats, acc plugins.Accumulator, host, version string) {

	// per topic overall (tag: name, paused, channel count)
	tags := map[string]string{
		`server_host`:    host,
		`server_version`: version,
		`topic`:          t.Name,
	}

	acc.Add(`nsq_topic_depth`, t.Depth, tags)
	acc.Add(`nsq_topic_backend_depth`, t.BackendDepth, tags)
	acc.Add(`nsq_topic_message_count`, t.MessageCount, tags)

	acc.Add(`nsq_topic_channel_count`, int64(len(t.Channels)), tags)
	for _, c := range t.Channels {
		channelStats(c, acc, host, version, t.Name)
	}
}

func channelStats(c ChannelStats, acc plugins.Accumulator, host, version, topic string) {
	tags := map[string]string{
		`server_host`:    host,
		`server_version`: version,
		`topic`:          topic,
		`channel`:        c.Name,
	}

	acc.Add("nsq_channel_depth", c.Depth, tags)
	acc.Add("nsq_channel_backend_depth", c.BackendDepth, tags)
	acc.Add("nsq_channel_inflight_count", c.InFlightCount, tags)
	acc.Add("nsq_channel_deferred_count", c.DeferredCount, tags)
	acc.Add("nsq_channel_message_count", c.MessageCount, tags)
	acc.Add("nsq_channel_requeue_count", c.RequeueCount, tags)
	acc.Add("nsq_channel_timeout_count", c.TimeoutCount, tags)

	acc.Add("nsq_channel_client_count", int64(len(c.Clients)), tags)
	for _, cl := range c.Clients {
		clientStats(cl, acc, host, version, topic, c.Name)
	}
}

func clientStats(c ClientStats, acc plugins.Accumulator, host, version, topic, channel string) {
	tags := map[string]string{
		`server_host`:       host,
		`server_version`:    version,
		`topic`:             topic,
		`channel`:           channel,
		`client_name`:       c.Name,
		`client_id`:         c.ID,
		`client_hostname`:   c.Hostname,
		`client_version`:    c.Version,
		`client_address`:    c.RemoteAddress,
		`client_user_agent`: c.UserAgent,
		`client_tls`:        strconv.FormatBool(c.TLS),
		`client_snappy`:     strconv.FormatBool(c.Snappy),
		`client_deflate`:    strconv.FormatBool(c.Deflate),
	}
	acc.Add("nsq_client_ready_count", c.ReadyCount, tags)
	acc.Add("nsq_client_inflight_count", c.InFlightCount, tags)
	acc.Add("nsq_client_message_count", c.MessageCount, tags)
	acc.Add("nsq_client_finish_count", c.FinishCount, tags)
	acc.Add("nsq_client_requeue_count", c.RequeueCount, tags)
}

type NSQStats struct {
	Code int64        `json:"status_code"`
	Txt  string       `json:"status_txt"`
	Data NSQStatsData `json:"data"`
}

type NSQStatsData struct {
	Version   string       `json:"version"`
	Health    string       `json:"health"`
	StartTime int64        `json:"start_time"`
	Topics    []TopicStats `json:"topics"`
}

// e2e_processing_latency is not modeled
type TopicStats struct {
	Name         string         `json:"topic_name"`
	Depth        int64          `json:"depth"`
	BackendDepth int64          `json:"backend_depth"`
	MessageCount int64          `json:"message_count"`
	Paused       bool           `json:"paused"`
	Channels     []ChannelStats `json:"channels"`
}

// e2e_processing_latency is not modeled
type ChannelStats struct {
	Name          string        `json:"channel_name"`
	Depth         int64         `json:"depth"`
	BackendDepth  int64         `json:"backend_depth"`
	InFlightCount int64         `json:"in_flight_count"`
	DeferredCount int64         `json:"deferred_count"`
	MessageCount  int64         `json:"message_count"`
	RequeueCount  int64         `json:"requeue_count"`
	TimeoutCount  int64         `json:"timeout_count"`
	Paused        bool          `json:"paused"`
	Clients       []ClientStats `json:"clients"`
}

type ClientStats struct {
	Name                          string `json:"name"`
	ID                            string `json:"client_id"`
	Hostname                      string `json:"hostname"`
	Version                       string `json:"version"`
	RemoteAddress                 string `json:"remote_address"`
	State                         int64  `json:"state"`
	ReadyCount                    int64  `json:"ready_count"`
	InFlightCount                 int64  `json:"in_flight_count"`
	MessageCount                  int64  `json:"message_count"`
	FinishCount                   int64  `json:"finish_count"`
	RequeueCount                  int64  `json:"requeue_count"`
	ConnectTime                   int64  `json:"connect_ts"`
	SampleRate                    int64  `json:"sample_rate"`
	Deflate                       bool   `json:"deflate"`
	Snappy                        bool   `json:"snappy"`
	UserAgent                     string `json:"user_agent"`
	TLS                           bool   `json:"tls"`
	TLSCipherSuite                string `json:"tls_cipher_suite"`
	TLSVersion                    string `json:"tls_version"`
	TLSNegotiatedProtocol         string `json:"tls_negotiated_protocol"`
	TLSNegotiatedProtocolIsMutual bool   `json:"tls_negotiated_protocol_is_mutual"`
}
