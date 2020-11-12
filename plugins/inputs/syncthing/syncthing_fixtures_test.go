package syncthing_test

import (
	"text/template"
)

type TestValues struct {
	FolderID      string
	FolderLabel   string
	FolderPath    string
	DeviceID      string
	InBytesTotal  int
	OutBytesTotal int
	TotalNeeded   int
	DeviceAddress string
	DeviceName    string
	ClientVersion string
	Connected     bool
	Crypto        string
	Paused        bool
	MyID          string
}

var (
	sysconfigJSON    = template.Must(template.New("sysconfig").Parse(systemConfigTemplate))
	folderNeedJSON   = template.Must(template.New("folderNeed").Parse(folderNeedTemplate))
	connectionJSON   = template.Must(template.New("connection").Parse(connectionsTemplate))
	systemStatusJSON = template.Must(template.New("sysstatus").Parse(systemStatusTemplate))
)

func init() {
}

const (
	systemStatusTemplate = `
	{
		"myID": "{{ .MyID }}",
		"startTime": "2016-06-06T19:41:43.039284753+02:00"
	}
	`

	systemConfigTemplate = `
	{
	  "version": 30,
	  "folders": [
	    {
	      "id": "{{ .FolderID }}",
	      "label": "{{ .FolderLabel }}",
	      "filesystemType": "basic",
	      "path": "{{ .FolderPath }}",
	      "type": "sendreceive",
	      "devices": [
		{
		  "deviceID": "{{ .DeviceID }}",
		  "introducedBy": ""
		}
	      ],
	      "rescanIntervalS": 60,
	      "fsWatcherEnabled": false,
	      "fsWatcherDelayS": 10,
	      "ignorePerms": false,
	      "autoNormalize": true,
	      "minDiskFree": {
		"value": 1,
		"unit": "%"
	      },
	      "versioning": {
		"type": "simple",
		"params": {
		  "keep": "5"
		}
	      },
	      "copiers": 0,
	      "pullerMaxPendingKiB": 0,
	      "hashers": 0,
	      "order": "random",
	      "ignoreDelete": false,
	      "scanProgressIntervalS": 0,
	      "pullerPauseS": 0,
	      "maxConflicts": 10,
	      "disableSparseFiles": false,
	      "disableTempIndexes": false,
	      "paused": false,
	      "weakHashThresholdPct": 25,
	      "markerName": ".stfolder",
	      "copyOwnershipFromParent": false,
	      "modTimeWindowS": 0
	    }
	  ],
	  "devices": [
	    {
	      "deviceID": "{{ .DeviceID }}",
	      "name": "{{ .DeviceName }}",
	      "addresses": [
		"dynamic",
		"tcp://192.168.1.2:22000"
	      ],
	      "compression": "metadata",
	      "certName": "",
	      "introducer": false,
	      "skipIntroductionRemovals": false,
	      "introducedBy": "",
	      "paused": false,
	      "allowedNetworks": [],
	      "autoAcceptFolders": false,
	      "maxSendKbps": 0,
	      "maxRecvKbps": 0,
	      "ignoredFolders": [],
	      "pendingFolders": [],
	      "maxRequestKiB": 0
	    }
	  ]
	}`

	folderNeedTemplate = `
	{
	  "progress": [
	    {
	      "flags": "0755",
	      "sequence": 6,
	      "modified": "2015-04-20T23:06:12+09:00",
	      "name": "ls",
	      "size": 34640,
	      "version": [
		"5157751870738175669:1"
	      ]
	    }
	  ],
	  "queued": [],
	  "rest": [],
	  "page": 1,
	  "perpage": 100,
	  "total": {{ .TotalNeeded }}
	}`

	connectionsTemplate = `
	{
	   "total" : {
		  "paused" : false,
		  "clientVersion" : "",
		  "at" : "2015-11-07T17:29:47.691637262+01:00",
		  "connected" : false,
		  "inBytesTotal" : 1479,
		  "type" : "",
		  "outBytesTotal" : 1318,
		  "address" : ""
	   },
	   "connections" : {
		  "{{ .DeviceID }}" : {
		     "connected" : {{ .Connected }},
		     "inBytesTotal" : {{ .InBytesTotal }},
		     "paused" : {{ .Paused }},
		     "at" : "2015-11-07T17:29:47.691548971+01:00",
		     "clientVersion" : "{{ .ClientVersion }}",
		     "address" : "{{ .DeviceAddress }}",
		     "type" : "TCP (Client)",
		     "crypto" : "{{ .Crypto }}",
		     "outBytesTotal" : {{ .OutBytesTotal }}
		  },
		  "DOVII4U-SQEEESM-VZ2CVTC-CJM4YN5-QNV7DCU-5U3ASRL-YVFG6TH-W5DV5AA" : {
		     "outBytesTotal" : 0,
		     "type" : "",
		     "address" : "",
		     "at" : "0001-01-01T00:00:00Z",
		     "clientVersion" : "",
		     "paused" : false,
		     "inBytesTotal" : 0,
		     "connected" : false
		  }
	   }
	}
	`
)
