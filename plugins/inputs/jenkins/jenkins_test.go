package jenkins

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestJenkins_Gather(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		switch r.RequestURI {

		case "/computer/api/json?depth=1":
			w.WriteHeader(http.StatusOK)
			resp = responseNode
		case "/job/JOB1/api/json":
			w.WriteHeader(http.StatusOK)
			resp = responseJob1
		case "/job/JOB1/1/api/json?depth=1":
			w.WriteHeader(http.StatusOK)
			resp = responseJob1Build1
		case "/job/JOB1/2/api/json?depth=1":
			w.WriteHeader(http.StatusOK)
			resp = responseJob1Build2
		case "/job/JOB1/api/json?tree=allBuilds%5Bnumber%2Curl%5D":
			w.WriteHeader(http.StatusOK)
			resp = responseJob1AllBuilds
		case "/api/json":
			w.WriteHeader(http.StatusOK)
			resp = response
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write([]byte(resp))
	}))

	defer ts.Close()

	var acc testutil.Accumulator

	j := &Jenkins{
		URL: ts.URL,
	}

	err := acc.GatherError(j.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"memory_total":     float64(6089498624),
		"swap_total":       float64(4294963200),
		"response_time":    int64(0),
		"disk_available":   float64(152393224192),
		"temp_available":   float64(152393224192),
		"memory_available": float64(463441920),
		"swap_available":   float64(3494731776),
	}
	tags := map[string]string{
		"node_name": "master",
		"arch":      "Linux (amd64)",
		"disk_path": "/var/jenkins_home",
		"temp_path": "/tmp",
		"online":    "true",
	}

	acc.AssertContainsTaggedFields(t, "jenkins_node", fields, tags)

	fields = map[string]interface{}{
		"duration_ms": int64(2022068),
		"result_code": int(1),
	}

	tags = map[string]string{
		"job_name": "JOB1",
		"result":   "FAILURE",
	}

	acc.AssertContainsTaggedFields(t, "jenkins_job", fields, tags)

	fields = map[string]interface{}{
		"duration_ms": int64(2137068),
		"result_code": int(0),
	}

	tags = map[string]string{
		"job_name": "JOB1",
		"result":   "SUCCESS",
	}

	acc.AssertContainsTaggedFields(t, "jenkins_job", fields, tags)

}

var response = `{
	"_class": "hudson.model.Hudson",
	"assignedLabels": [{}
	],
	"mode": "NORMAL",
	"nodeDescription": "the master Jenkins node",
	"nodeName": "",
	"numExecutors": 4,
	"description": null,
	"jobs": [{
			"_class": "hudson.model.FreeStyleProject",
			"name": "JOB1",
			"url": "http://localhost:8080/job/JOB1/",
			"color": "blue"
		}
	],
	"overallLoad": {},
	"primaryView": {
		"_class": "hudson.model.AllView",
		"name": "all",
		"url": "http://localhost:8080/"
	},
	"quietingDown": false,
	"slaveAgentPort": 50000,
	"unlabeledLoad": {
		"_class": "jenkins.model.UnlabeledLoadStatistics"
	},
	"useCrumbs": false,
	"useSecurity": true,
	"views": [{
			"_class": "hudson.model.ListView",
			"name": "Delivery",
			"url": "http://localhost:8080/view/Delivery/"
		}, {
			"_class": "hudson.model.ListView",
			"name": "Legacy",
			"url": "http://localhost:8080/view/Legacy/"
		}, {
			"_class": "hudson.model.AllView",
			"name": "all",
			"url": "http://localhost:8080"
		}
	]
}`

var responseNode = `{
	"_class": "hudson.model.ComputerSet",
	"busyExecutors": 0,
	"computer": [{
			"_class": "hudson.model.Hudson$MasterComputer",
			"actions": [{}
			],
			"displayName": "master",
			"executors": [{}, {}, {}, {}
			],
			"icon": "computer.png",
			"iconClassName": "icon-computer",
			"idle": true,
			"jnlpAgent": false,
			"launchSupported": true,
			"loadStatistics": {
				"_class": "hudson.model.Label$1"
			},
			"manualLaunchAllowed": true,
			"monitorData": {
				"hudson.node_monitors.SwapSpaceMonitor": {
					"_class": "hudson.node_monitors.SwapSpaceMonitor$MemoryUsage2",
					"availablePhysicalMemory": 463441920,
					"availableSwapSpace": 3494731776,
					"totalPhysicalMemory": 6089498624,
					"totalSwapSpace": 4294963200
				},
				"hudson.node_monitors.TemporarySpaceMonitor": {
					"_class": "hudson.node_monitors.DiskSpaceMonitorDescriptor$DiskSpace",
					"timestamp": 1515746819525,
					"path": "/tmp",
					"size": 152393224192
				},
				"hudson.node_monitors.DiskSpaceMonitor": {
					"_class": "hudson.node_monitors.DiskSpaceMonitorDescriptor$DiskSpace",
					"timestamp": 1515746819400,
					"path": "/var/jenkins_home",
					"size": 152393224192
				},
				"hudson.node_monitors.ArchitectureMonitor": "Linux (amd64)",
				"hudson.node_monitors.ResponseTimeMonitor": {
					"_class": "hudson.node_monitors.ResponseTimeMonitor$Data",
					"timestamp": 1515746819409,
					"average": 0
				},
				"hudson.node_monitors.ClockMonitor": {
					"_class": "hudson.util.ClockDifference",
					"diff": 0
				}
			},
			"numExecutors": 4,
			"offline": false,
			"offlineCause": null,
			"offlineCauseReason": "",
			"oneOffExecutors": [],
			"temporarilyOffline": false
		}
	],
	"displayName": "Nodes",
	"totalExecutors": 4
}`

var responseJob1 = `{
	"_class": "hudson.model.FreeStyleProject",
	"actions": [],
	"description": "",
	"displayName": "JOB1",
	"displayNameOrNull": null,
	"fullDisplayName": "JOB1",
	"fullName": "JOB1",
	"name": "JOB1",
	"url": "http://localhost:8080/job/JOB1/",
	"buildable": true,
	"builds": [{
			"_class": "hudson.model.FreeStyleBuild",
			"number": 2,
			"url": "http://localhost:8080/job/JOB1/2/"
		}, {
			"_class": "hudson.model.FreeStyleBuild",
			"number": 1,
			"url": "http://localhost:8080/job/JOB1/1/"
		}
	],
	"color": "blue",
	"firstBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 1,
		"url": "http://localhost:8080/job/JOB1/1/"
	},
	"healthReport": [{
			"description": "Build stability: No recent builds failed.",
			"iconClassName": "icon-health-80plus",
			"iconUrl": "health-80plus.png",
			"score": 100
		}
	],
	"inQueue": false,
	"keepDependencies": false,
	"lastBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 2,
		"url": "http://localhost:8080/job/JOB1/2/"
	},
	"lastCompletedBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 2,
		"url": "http://localhost:8080/job/JOB1/2/"
	},
	"lastFailedBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 1,
		"url": "http://localhost:8080/job/JOB1/1/"
	},
	"lastStableBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 2,
		"url": "http://localhost:8080/job/JOB1/2/"
	},
	"lastSuccessfulBuild": {
		"_class": "hudson.model.FreeStyleBuild",
		"number": 2,
		"url": "http://localhost:8080/job/JOB1/2/"
	},
	"lastUnstableBuild": null,
	"lastUnsuccessfulBuild": null,
	"nextBuildNumber": 3,
	"property": [],
	"queueItem": null,
	"concurrentBuild": false,
	"downstreamProjects": [],
	"scm": {
		"_class": "hudson.scm.NullSCM"
	},
	"upstreamProjects": []
}
`

var responseJob1Build1 = strings.Replace(`{
	"_class": "hudson.model.FreeStyleBuild",
	"actions": [],
	"artifacts": [],
	"building": false,
	"description": null,
	"displayName": "#1",
	"duration": 2022068,
	"estimatedDuration": 1250570,
	"executor": null,
	"fullDisplayName": "JOB1 #1",
	"id": "1",
	"keepLog": false,
	"number": 1,
	"queueId": 31541,
	"result": "FAILURE",
	"timestamp": %timestamp%,
	"url": "http://localhost:8080/job/JOB1/1/",
	"builtOn": "",
	"changeSet": {
		"_class": "hudson.scm.EmptyChangeLogSet",
		"items": [],
		"kind": null
	}
}`, "%timestamp%", strconv.Itoa(int(time.Now().Unix())), 1)

var responseJob1Build2 = strings.Replace(`{
	"_class": "hudson.model.FreeStyleBuild",
	"actions": [],
	"artifacts": [],
	"building": false,
	"description": null,
	"displayName": "#2",
	"duration": 2137068,
	"estimatedDuration": 1450370,
	"executor": null,
	"fullDisplayName": "JOB1 #2",
	"id": "2",
	"keepLog": false,
	"number": 2,
	"queueId": 31542,
	"result": "SUCCESS",
	"timestamp": %timestamp%,
	"url": "http://localhost:8080/job/JOB1/2/",
	"builtOn": "",
	"changeSet": {
		"_class": "hudson.scm.EmptyChangeLogSet",
		"items": [],
		"kind": null
	}
}`, "%timestamp%", strconv.Itoa(int(time.Now().Unix())), 1)

var responseJob1AllBuilds = `{
	"_class": "hudson.model.FreeStyleProject",
	"allBuilds": [
		{
			"_class": "hudson.model.FreeStyleBuild",
			"number": 2,
			"url": "http://localhost:8080/job/JOB1/2/"
		}, {
			"_class": "hudson.model.FreeStyleBuild",
			"number": 1,
			"url": "http://localhost:8080/job/JOB1/1/"
		}
	]
}`
