package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	s := &ConfigAPIService{}
	srv := httptest.NewServer(s.mux())

	resp, err := http.Get(srv.URL + "/status")
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.EqualValues(t, "ok", body)
	require.EqualValues(t, 200, resp.StatusCode)
}

func TestStartPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := config.NewConfig()
	a := agent.NewAgent(ctx, c)
	api, outputCancel := newAPI(ctx, c, a)
	go a.RunWithAPI(outputCancel)

	s := &ConfigAPIService{
		api: api,
	}
	srv := httptest.NewServer(s.mux())

	buf := bytes.NewBufferString(`{
		"name": "inputs.file",
		"config": {
			"files": ["testdata.lp"],
			"data_format": "influx"
		}
}`)
	resp, err := http.Post(srv.URL+"/plugins/create", "application/json", buf)
	require.NoError(t, err)
	createResp := struct {
		ID string
	}{}
	require.EqualValues(t, 200, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err)
	_ = resp.Body.Close()

	require.Regexp(t, `^[\da-f]{8}\d{8}$`, createResp.ID)

	statusResp := struct {
		Status string
		Reason string
	}{}

	for statusResp.Status != "running" {
		resp, err = http.Get(srv.URL + "/plugins/" + createResp.ID + "/status")
		require.NoError(t, err)

		require.EqualValues(t, 200, resp.StatusCode)
		err = json.NewDecoder(resp.Body).Decode(&statusResp)
		require.NoError(t, err)
		_ = resp.Body.Close()
	}

	require.EqualValues(t, "running", statusResp.Status)

	resp, err = http.Get(srv.URL + "/plugins/list")
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)
	listResp := []PluginConfigTypeInfo{}
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	_ = resp.Body.Close()

	if len(listResp) < 20 {
		require.FailNow(t, "expected there to be more than 20 plugins loaded, was only", len(listResp))
	}

	resp, err = http.Get(srv.URL + "/plugins/running")
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)
	runningList := []Plugin{}
	err = json.NewDecoder(resp.Body).Decode(&runningList)
	require.NoError(t, err)
	_ = resp.Body.Close()

	if len(runningList) != 1 {
		require.FailNow(t, "expected there to be 1 running plugin, was", len(runningList))
	}
}
