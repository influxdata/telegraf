package transmission

import (
	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testJSONTorrent = `{"arguments":{"torrents":[["status","totalSize","peersConnected","peersGettingFromUs","peersSendingToUs","rateDownload","rateUpload"],[6,33345888,0,0,0,0,0],[0,371033238,0,0,0,0,0],[6,1682767645,2,1,0,0,16000],[4,135178626499,50,15,10,20,3013000]]},"result":"success","tag":TAG}`
const testJSONSession = `{"arguments":{"peer-port":6881},"result":"success","tag":TAG}`

func TestTransmissionGather(t *testing.T) {
	var acc testutil.Accumulator

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		reqPayload := &requestPayload{}
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		require.NoError(t, err)

		w.Header().Add("Content-Type", "application/json")
		if reqPayload.Method == "torrent-get" {
			_, err := w.Write([]byte(strings.Replace(testJSONTorrent, "TAG", strconv.FormatInt(int64(reqPayload.Tag), 10), 1)))
			require.NoError(t, err)
		}
		if reqPayload.Method == "session-get" {
			_, err = w.Write([]byte(strings.Replace(testJSONSession, "TAG", strconv.FormatInt(int64(reqPayload.Tag), 10), 1)))
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		if u.Scheme == "http" {
			port = "80"
		} else if u.Scheme == "https" {
			port = "443"
		} else {
			port = ""
		}
	}

	trans := Transmission{
		URL: ts.URL,
	}
	err = trans.Init()
	require.NoError(t, err)

	err = trans.Gather(&acc)
	require.NoError(t, err)

	expectFields := map[string]interface{}{
		"torrents_stopped":            int64(1),
		"torrents_queued_checking":    int64(0),
		"torrents_checking":           int64(0),
		"torrents_queued_downloading": int64(0),
		"torrents_downloading":        int64(1),
		"torrents_queued_seeding":     int64(0),
		"torrents_seeding":            int64(2),
		"torrents_size":               int64(137265773270),
		"peers_connected":             int64(52),
		"peers_getting_from_us":       int64(16),
		"peers_sending_to_us":         int64(10),
		"torrents_active":             int64(2),
		"download_speed":              int64(20),
		"upload_speed":                int64(3029000),
	}

	expectTags := map[string]string{
		"url":       ts.URL,
		"rpc_host":  host,
		"rpc_port":  port,
		"peer_port": "6881",
	}

	acc.AssertContainsTaggedFields(t, "transmission", expectFields, expectTags)
}

func TestTransmissionGatherTagError(t *testing.T) {
	var acc testutil.Accumulator

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		reqPayload := &requestPayload{}
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		require.NoError(t, err)

		w.Header().Add("Content-Type", "application/json")
		if reqPayload.Method == "torrent-get" {
			_, err := w.Write([]byte(strings.Replace(testJSONTorrent, "TAG", strconv.FormatInt(int64(reqPayload.Tag)+1, 10), 1)))
			require.NoError(t, err)
		}
		if reqPayload.Method == "session-get" {
			_, err = w.Write([]byte(strings.Replace(testJSONSession, "TAG", strconv.FormatInt(int64(reqPayload.Tag)+1, 10), 1)))
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	trans := Transmission{
		URL: ts.URL,
	}
	err := trans.Init()
	require.NoError(t, err)

	err = trans.Gather(&acc)
	require.EqualError(t, err, "tag mismatch")
}

func TestTransmissionGatherInvalidJSON(t *testing.T) {
	var acc testutil.Accumulator

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		reqPayload := &requestPayload{}
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		require.NoError(t, err)

		w.Header().Add("Content-Type", "application/json")
		if reqPayload.Method == "torrent-get" {
			_, err := w.Write([]byte(testJSONTorrent))
			require.NoError(t, err)
		}
		if reqPayload.Method == "session-get" {
			_, err = w.Write([]byte(testJSONSession))
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	trans := Transmission{
		URL: ts.URL,
	}
	err := trans.Init()
	require.NoError(t, err)

	err = trans.Gather(&acc)
	require.EqualError(t, err, "can not decode JSON response")
}

func TestTransmissionGatherUsingBasicAuth(t *testing.T) {
	var acc testutil.Accumulator

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		reqPayload := &requestPayload{}
		err := json.NewDecoder(r.Body).Decode(&reqPayload)
		require.NoError(t, err)

		user, pass, ok := r.BasicAuth()
		if !ok || user != "testUser" || pass != "testPass" {
			w.WriteHeader(401)
			_, err := w.Write([]byte("Unauthorised.\n"))
			require.NoError(t, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if reqPayload.Method == "torrent-get" {
			_, err := w.Write([]byte(strings.Replace(testJSONTorrent, "TAG", strconv.FormatInt(int64(reqPayload.Tag), 10), 1)))
			require.NoError(t, err)
		}
		if reqPayload.Method == "session-get" {
			_, err = w.Write([]byte(strings.Replace(testJSONSession, "TAG", strconv.FormatInt(int64(reqPayload.Tag), 10), 1)))
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	transOK := Transmission{
		URL:      ts.URL,
		Username: "testUser",
		Password: "testPass",
	}
	err := transOK.Init()
	require.NoError(t, err)

	err = transOK.Gather(&acc)
	require.NoError(t, err)

	transErr := Transmission{
		URL:      ts.URL,
		Username: "invalid",
		Password: "invalid",
	}
	err = transErr.Init()
	require.NoError(t, err)

	err = transErr.Gather(&acc)
	require.EqualError(t, err, "HTTP status code is not 200 OK")
}
