package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
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

	outputCtx, outputCancel := context.WithCancel(context.Background())
	defer outputCancel()

	c := config.NewConfig()
	a := agent.NewAgent(ctx, c)
	api := newAPI(ctx, outputCtx, c, a)
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

func TestStopPlugin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	outputCtx, outputCancel := context.WithCancel(context.Background())
	defer outputCancel()

	c := config.NewConfig()
	a := agent.NewAgent(ctx, c)
	api := newAPI(ctx, outputCtx, c, a)
	go a.RunWithAPI(outputCancel)

	s := &ConfigAPIService{
		api: api,
		Log: testutil.Logger{},
	}
	srv := httptest.NewServer(s.mux())

	// start plugin
	buf := bytes.NewBufferString(`{
		"name": "inputs.cpu",
		"config": {
			"percpu": true
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

	waitForStatus(t, api, models.PluginID(createResp.ID), models.PluginStateRunning.String(), 2*time.Second)

	resp, err = http.Get(srv.URL + "/plugins/running")
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)
	runningList := []Plugin{}
	err = json.NewDecoder(resp.Body).Decode(&runningList)
	require.NoError(t, err)
	_ = resp.Body.Close()

	// confirm plugin is running
	if len(runningList) != 1 {
		require.FailNow(t, "expected there to be 1 running plugin, was", len(runningList))
	}

	// stop plugin
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", srv.URL+"/plugins/"+createResp.ID, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	require.EqualValues(t, 200, resp.StatusCode)
	require.NoError(t, err)

	waitForStatus(t, api, models.PluginID(createResp.ID), models.PluginStateDead.String(), 2*time.Second)

	resp, err = http.Get(srv.URL + "/plugins/running")
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)
	runningList = []Plugin{}
	err = json.NewDecoder(resp.Body).Decode(&runningList)
	require.NoError(t, err)
	_ = resp.Body.Close()

	// confirm plugin has stopped
	if len(runningList) >= 1 {
		require.FailNow(t, "expected there to be no running plugin, was", len(runningList))
	}

	// try to delete a plugin which was already been deleted
	req, err = http.NewRequest("DELETE", srv.URL+"/plugins/"+createResp.ID, nil)
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	require.EqualValues(t, 404, resp.StatusCode)
	require.NoError(t, err)
}

func TestStatusCodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outputCtx, outputCancel := context.WithCancel(context.Background())
	defer outputCancel()

	c := config.NewConfig()
	a := agent.NewAgent(ctx, c)
	api := newAPI(ctx, outputCtx, c, a)
	go a.RunWithAPI(outputCancel)

	s := &ConfigAPIService{
		api: api,
		Log: testutil.Logger{},
	}
	srv := httptest.NewServer(s.mux())

	// Error finding plugin with wrong name
	buf := bytes.NewBufferString(`{
			"name": "inputs.blah",
			"config": {
				"files": ["testdata.lp"],
				"data_format": "influx"
			}
	}`)
	resp, err := http.Post(srv.URL+"/plugins/create", "application/json", buf)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.EqualValues(t, 404, resp.StatusCode)
	require.NoError(t, err)

	// Error creating plugin with wrong data format
	buf = bytes.NewBufferString(`{
			"name": "inputs.file",
			"config": {
				"files": ["testdata.lp"],
				"data_format": "blah"
			}
	}`)
	resp, err = http.Post(srv.URL+"/plugins/create", "application/json", buf)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)
	require.NoError(t, err)
}
