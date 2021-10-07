package transmission

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const sampleConfig = `
  ## An URL where the Transmission RPC API is available
  url = "http://127.0.0.1:9091/transmission/rpc"

  ## Timeout for HTTP requests
  # timeout = "5s"

  ## Optional HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

const description = "Collect Transmission client statistics about bandwidth usage and torrent status"

// Transmission is an input that collects client statistics
// about bandwidth usage and torrent status of the Transmission BitTorrent client.
type Transmission struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`

	Timeout config.Duration `toml:"timeout"`

	// Used as tags
	rpcHost string
	rpcPort string

	tls.ClientConfig
	client     *http.Client
	csrfHeader string
}

type requestPayload struct {
	Method    string      `json:"method"`
	Arguments interface{} `json:"arguments,omitempty"`
	Tag       int         `json:"tag,omitempty"`
}

type responsePayload struct {
	Arguments interface{} `json:"arguments"`
	Result    string      `json:"result"`
	Tag       *int        `json:"tag"`
}

type sessionInformation struct {
	PeerPort int64 `json:"peer-port"`
}

type torrentStats struct {
	Active             int64
	Stopped            int64
	QueuedChecking     int64
	Checking           int64
	QueuedDownloading  int64
	Downloading        int64
	QueuedSeeding      int64
	Seeding            int64
	Size               int64
	PeersConnected     int64
	PeersGettingFromUs int64
	PeersSendingToUs   int64
	TotalDownloadSpeed int64
	TotalUploadSpeed   int64
}

func (t *Transmission) Description() string {
	return description
}

func (t *Transmission) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (t *Transmission) Init() error {
	if t.client == nil {
		client, err := t.createHTTPClient()

		if err != nil {
			return err
		}
		t.client = client
	}

	u, err := url.Parse(t.URL)
	if err != nil {
		return err
	}

	t.rpcHost, t.rpcPort, err = net.SplitHostPort(u.Host)
	if err != nil {
		t.rpcHost = u.Host
		if u.Scheme == "http" {
			t.rpcPort = "80"
		} else if u.Scheme == "https" {
			t.rpcPort = "443"
		} else {
			t.rpcPort = ""
		}
	}

	return nil
}

// createHTTPClient creates an HTTP client to access the RPC API.
func (t *Transmission) createHTTPClient() (*http.Client, error) {
	tlsConfig, err := t.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	timeout := t.Timeout
	if timeout == 0 {
		timeout = config.Duration(time.Second * 5)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(timeout),
	}

	return client, nil
}

// doRPC executes the RPC using the given method and optional arguments for that method.
// It tries to unmarshal the specific RPC response for that method using the given result interface.
// If retry is true it will retry the call once, using a new X-Transmission-Session-Id in case of a http.StatusConflict.
func (t *Transmission) doRPC(method string, arguments interface{}, result interface{}, retry bool) error {
	// Random number used by clients to track responses
	tag := rand.Int()

	reqPayload, err := json.Marshal(requestPayload{
		Method:    method,
		Arguments: arguments,
		Tag:       tag,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", t.URL, bytes.NewBuffer(reqPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", t.csrfHeader)
	if t.Username != "" && t.Password != "" {
		req.SetBasicAuth(t.Username, t.Password)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusConflict {
		t.csrfHeader = resp.Header.Get("X-Transmission-Session-Id")
		if retry {
			return t.doRPC(method, arguments, result, retry)
		}
		return errors.New("CSRF header invalid twice in a row, aborting")
	} else if resp.StatusCode != http.StatusOK {
		return errors.New("HTTP status code is not 200 OK")
	}

	defer resp.Body.Close()
	respPayload := responsePayload{Arguments: result}
	if err = json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return errors.New("can not decode JSON response")
	}

	if respPayload.Tag == nil || *respPayload.Tag != tag {
		return errors.New("tag mismatch")
	}

	if respPayload.Result != "success" {
		return errors.New("RPC unsuccessful")
	}

	return nil
}

// getSessionInformation calls the session-get method to gather information about the Transmission client itself.
func (t *Transmission) getSessionInformation() (*sessionInformation, error) {
	arguments := map[string]interface{}{
		"fields": []string{"peer-port"},
	}

	si := &sessionInformation{}
	err := t.doRPC("session-get", arguments, si, true)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// getTorrentStats calls the torrent-get method to gather information about individual torrents.
func (t *Transmission) getTorrentStats() (*torrentStats, error) {
	arguments := map[string]interface{}{
		"format": "table",
		"fields": []string{"status", "totalSize", "peersConnected", "peersGettingFromUs", "peersSendingToUs", "rateDownload", "rateUpload"},
	}

	type torrentResponse struct {
		Torrents []json.RawMessage `json:"torrents"`
	}
	tr := &torrentResponse{}

	err := t.doRPC("torrent-get", arguments, tr, true)
	if err != nil {
		return nil, err
	}

	row := [7]int64{}
	ts := &torrentStats{}
	for _, v := range tr.Torrents[1:] {
		err := json.Unmarshal(v, &row)
		if err != nil {
			continue
		}
		switch row[0] {
		case 0:
			ts.Stopped++
		case 1:
			ts.QueuedChecking++
		case 2:
			ts.Checking++
		case 3:
			ts.QueuedDownloading++
		case 4:
			ts.Downloading++
		case 5:
			ts.QueuedSeeding++
		case 6:
			ts.Seeding++
		}
		ts.Size += row[1]
		ts.PeersConnected += row[2]
		ts.PeersGettingFromUs += row[3]
		ts.PeersSendingToUs += row[4]
		if row[5] > 0 || row[6] > 0 {
			ts.Active++
		}
		ts.TotalDownloadSpeed += row[5]
		ts.TotalUploadSpeed += row[6]
	}

	return ts, nil
}

func (t *Transmission) Gather(acc telegraf.Accumulator) error {
	si, err := t.getSessionInformation()
	if err != nil {
		return err
	}

	ts, err := t.getTorrentStats()
	if err != nil {
		return err
	}

	tags := map[string]string{
		"url":       t.URL,
		"rpc_host":  t.rpcHost,
		"rpc_port":  t.rpcPort,
		"peer_port": strconv.FormatInt(si.PeerPort, 10),
	}

	fields := map[string]interface{}{
		"torrents_active":             ts.Active,
		"torrents_stopped":            ts.Stopped,
		"torrents_queued_checking":    ts.QueuedChecking,
		"torrents_checking":           ts.Checking,
		"torrents_queued_downloading": ts.QueuedDownloading,
		"torrents_downloading":        ts.Downloading,
		"torrents_queued_seeding":     ts.QueuedSeeding,
		"torrents_seeding":            ts.Seeding,
		"torrents_size":               ts.Size,
		"peers_connected":             ts.PeersConnected,
		"peers_getting_from_us":       ts.PeersGettingFromUs,
		"peers_sending_to_us":         ts.PeersSendingToUs,
		"download_speed":              ts.TotalDownloadSpeed,
		"upload_speed":                ts.TotalUploadSpeed,
	}

	acc.AddFields("transmission", fields, tags)

	return nil
}

func init() {
	inputs.Add("transmission", func() telegraf.Input { return &Transmission{} })
}
