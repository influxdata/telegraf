package netdata

import (
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNetdataStat(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/charts", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, charts)
	})
	mux.HandleFunc("/api/v1/data", func(w http.ResponseWriter, req *http.Request) {
		chartname := req.FormValue("chart")
		switch chartname {
		case "ipv4.tcpconnaborts":
			fmt.Fprintln(w, datatcpconnaborts)
		case "ipv4.mcastpkts":
			fmt.Fprintln(w, datamcastpkts)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	n := &Netdata{
		Servers: []Server{
			{
				Url: ts.URL,
			},
		},
		Points: 10,
		Group:  "average",
		PrefixChartIDByHostname: true,
	}

	var acc testutil.Accumulator
	n.Gather(&acc)
	assert.Empty(t, acc.Errors, "Gather contains errors")

	tsURL, _ := url.Parse(ts.URL)

	fields := map[string]interface{}{
		"baddata":    float64(2),
		"userclosed": float64(4),
		"nomemory":   float64(8),
		"timeout":    float64(3),
		"linger":     float64(9),
		"failed":     float64(8),
	}
	tags := map[string]string{
		"server":     tsURL.Host,
		"hostname":   "testhost",
		"type":       "ipv4",
		"family":     "tcp",
		"dimensions": "baddata,userclosed,nomemory,timeout,linger,failed",
	}
	acc.AssertContainsTaggedFields(t, "netdata.testhost.ipv4.tcpconnaborts", fields, tags)

	fields = map[string]interface{}{
		"received": float64(42),
		"sent":     float64(34),
	}
	tags = map[string]string{
		"server":     tsURL.Host,
		"hostname":   "testhost",
		"type":       "ipv4",
		"family":     "multicast",
		"dimensions": "received,sent",
	}
	acc.AssertContainsTaggedFields(t, "netdata.testhost.ipv4.mcastpkts", fields, tags)
}

var charts = `
{
	"hostname": "testhost",
	"update_every": 2,
	"history": 3600,
	"charts": {
		"ipv4.tcpconnaborts": 		{
			"id": "ipv4.tcpconnaborts",
			"name": "ipv4.tcpconnaborts",
			"type": "ipv4",
			"family": "tcp",
			"context": "ipv4.tcpconnaborts",
			"title": "TCP Connection Aborts (ipv4.tcpconnaborts)",
			"priority": 3010,
			"enabled": true,
			"units": "connections/s",
			"data_url": "/api/v1/data?chart=ipv4.tcpconnaborts",
			"chart_type": "line",
			"duration": 7200,
			"first_entry": 1476120982,
			"last_entry": 1476125718,
			"update_every": 2,
			"dimensions": {
				"TCPAbortOnData": { "name": "baddata" },
				"TCPAbortOnClose": { "name": "userclosed" },
				"TCPAbortOnMemory": { "name": "nomemory" },
				"TCPAbortOnTimeout": { "name": "timeout" },
				"TCPAbortOnLinger": { "name": "linger" },
				"TCPAbortFailed": { "name": "failed" }
			},
			"green": null,
			"red": null
		},
		"ipv4.mcastpkts": 		{
			"id": "ipv4.mcastpkts",
			"name": "ipv4.mcastpkts",
			"type": "ipv4",
			"family": "multicast",
			"context": "ipv4.mcastpkts",
			"title": "IPv4 Multicast Packets (ipv4.mcastpkts)",
			"priority": 8600,
			"enabled": true,
			"units": "packets/s",
			"data_url": "/api/v1/data?chart=ipv4.mcastpkts",
			"chart_type": "line",
			"duration": 7200,
			"first_entry": 1476120920,
			"last_entry": 1476125718,
			"update_every": 2,
			"dimensions": {
				"InMcastPkts": { "name": "received" },
				"OutMcastPkts": { "name": "sent" }
			},
			"green": null,
			"red": null
		}
    }
}
`

var datatcpconnaborts = `
{
	"api": 1,
	"id": "ipv4.tcpconnaborts",
	"name": "ipv4.tcpconnaborts",
	"view_update_every": 8,
	"update_every": 2,
	"first_entry": 1478873812,
	"last_entry": 1478881012,
	"before": 1478880928,
	"after": 1478880842,
	"dimension_names": ["baddata", "userclosed", "nomemory", "timeout", "linger", "failed"],
	"dimension_ids": ["TCPAbortOnData", "TCPAbortOnClose", "TCPAbortOnMemory", "TCPAbortOnTimeout", "TCPAbortOnLinger", "TCPAbortFailed"],
	"latest_values": [0, 0, 0, 0, 0, 0],
	"view_latest_values": [0, 0, 0, 0, 0, 0],
	"dimensions": 6,
	"points": 11,
	"format": "json",
	"result": {
		"labels": ["time", "baddata", "userclosed", "nomemory", "timeout", "linger", "failed"],
		"data":
		[
			[1478880928000, 2, 4, 8, 3, 9, 8]
		]
	},
	"min": 0,
	"max": 0
}
`

var datamcastpkts = `
{
	"api": 1,
	"id": "ipv4.mcastpkts",
	"name": "ipv4.mcastpkts",
	"view_update_every": 8,
	"update_every": 2,
	"first_entry": 1478874070,
	"last_entry": 1478881270,
	"before": 1478880928,
	"after": 1478880842,
	"dimension_names": ["received", "sent"],
	"dimension_ids": ["InMcastPkts", "OutMcastPkts"],
	"latest_values": [0, 0],
	"view_latest_values": [0, 0],
	"dimensions": 2,
	"points": 11,
	"format": "json",
	"result": {
		"labels": ["time", "received", "sent"],
		"data":
		[
			[1478880928000, 42, 34]
		]
	},
	"min": 0,
	"max": 0
}
`
