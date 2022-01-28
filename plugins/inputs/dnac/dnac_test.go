package dnac_test

import (
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/dnac"
	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func Test_Gather(t *testing.T) {
	plugin := dnac.NewDnac()

	plugin.DnacBaseURL = "https://192.168.196.2"
	plugin.Username = "test"
	plugin.Password = "test123"
	plugin.Debug = "false"
	plugin.SSLVerify = "false"
	plugin.ClientHealth = true
	plugin.NetworkHealth = true

	plugin.InitClient()

	httpmock.ActivateNonDefault(plugin.Client.RestyClient().GetClient())

	var err error

	auth_responder, err := httpmock.NewJsonResponder(200, auth_fixture)
	auth_url := "https://192.168.196.2/dna/system/api/v1/auth/token"
	httpmock.RegisterResponder("POST", auth_url, auth_responder)

	if err != nil {
		t.Log("invalid auth fixture")
		t.FailNow()
	}

	json.Unmarshal([]byte(client_health_string), &client_health_fixture)
	client_health_responder, err := httpmock.NewJsonResponder(200, client_health_fixture)
	client_health_url := "https://192.168.196.2/dna/intent/api/v1/client-health"
	httpmock.RegisterResponder("GET", client_health_url, client_health_responder)

	if err != nil {
		t.Log("invalid client_health fixture")
		t.FailNow()
	}

	json.Unmarshal([]byte(network_health_string), &network_health_fixture)
	network_health_responder, err := httpmock.NewJsonResponder(200, network_health_fixture)
	network_health_url := "https://192.168.196.2/dna/intent/api/v1/network-health"
	httpmock.RegisterResponder("GET", network_health_url, network_health_responder)

	if err != nil {
		t.Log("invalid network_health fixture")
		t.FailNow()
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 2)
	client_health_metric := acc.Metrics[0]
	require.Len(t, client_health_metric.Fields, 21)
	require.Len(t, client_health_metric.Tags, 2)
	network_health_metric := acc.Metrics[1]
	require.Len(t, network_health_metric.Fields, 66)
	require.Len(t, network_health_metric.Tags, 2)

}

var auth_fixture = map[string]string{"Token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2MDY2NzE2NmU2YjNmNDAwYzJmMmNhOTciLCJhdXRoU291cmNlIjoiaW50ZXJuYWwiLCJ0ZW5hbnROYW1lIjoiVE5UMCIsInJvbGVzIjpbIjYwNjY3MTY1ZTZiM2Y0MDBjMmYyY2E5NiJdLCJ0ZW5hbnRJZCI6IjYwNjY3MTY1ZTZiM2Y0MDBjMmYyY2E5NCIsImV4cCI6MTY0MTgzNjIwNSwiaWF0IjoxNjQxODMyNjA1LCJqdGkiOiJkNDY5MWI4My1iZGIwLTQ3MWQtOTllZi1jOTQzZGQxMjA5ZTEiLCJ1c2VybmFtZSI6ImFkbWluIn0.HQCZ-EffKDgjb7jHmyHjfr1vNP6u4xBImcgS82AQK-4whsgIwyGu1xeJujT7zLRFdt-Ojz3pvRCiUhvXRaCIpSXnhR_s6otFJCsmZE_oJcFwVLWa7_7heWN4XAxjwUH37_2wiGRay4HFwkh85O4I_VW_NkrUfoLFeWpksBFX8edcbBLUC3iBtMG0wp-ogQdfebKxnpwaEy2o8iR38RhcAJDnOa4XCVGGDTTUZTsEID5yE_M_XRMdYkDA6g0FY3mbsj0VWI8fPELhHgjR39enLDOsykMefmquMkm7mqhufMslC5gls5MlLjmw53xzsAwmQe8NRYD9uuMrGEVFE0Ep3Q"}

var client_health_string = `{
	"response" : [ {
	  "siteId" : "global",
	  "scoreDetail" : [ {
		"scoreCategory" : {
		  "scoreCategory" : "CLIENT_TYPE",
		  "value" : "ALL"
		},
		"scoreValue" : 97,
		"clientCount" : 1165,
		"clientUniqueCount" : 1165,
		"starttime" : 1641830700000,
		"endtime" : 1641831000000
	  }, {
		"scoreCategory" : {
		  "scoreCategory" : "CLIENT_TYPE",
		  "value" : "WIRED"
		},
		"scoreValue" : 99,
		"clientCount" : 1047,
		"clientUniqueCount" : 1047,
		"starttime" : 1641830700000,
		"endtime" : 1641831000000,
		"scoreList" : [ {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "POOR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 14,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 14,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "DHCP"
			},
			"scoreValue" : -1,
			"clientCount" : 14,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "FAIR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "GOOD"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 1033,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 1033,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "IDLE"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NODATA"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NEW"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000
		} ]
	  }, {
		"scoreCategory" : {
		  "scoreCategory" : "CLIENT_TYPE",
		  "value" : "WIRELESS"
		},
		"scoreValue" : 79,
		"clientCount" : 118,
		"clientUniqueCount" : 118,
		"starttime" : 1641830700000,
		"endtime" : 1641831000000,
		"scoreList" : [ {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "POOR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 9,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 9,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "ASSOCIATION"
			},
			"scoreValue" : -1,
			"clientCount" : 7,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "AAA"
			},
			"scoreValue" : -1,
			"clientCount" : 2,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "FAIR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 15,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 15,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "GOOD"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 92,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 92,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "IDLE"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 2,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 2,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NODATA"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NEW"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 1,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700000,
		  "endtime" : 1641831000000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 1,
			"clientUniqueCount" : null,
			"starttime" : 1641830700000,
			"endtime" : 1641831000000
		  } ]
		} ]
	  } ]
	} ]
  }`

var client_health_fixture map[string]interface{}

var network_health_string = `{
	"version" : "1.0",
	"response" : [ {
	  "time" : "2022-01-10T17:50:00.000+0000",
	  "healthScore" : 92,
	  "totalCount" : 404,
	  "goodCount" : 373,
	  "noHealthCount" : 2,
	  "fairCount" : 29,
	  "badCount" : 0,
	  "entity" : null,
	  "timeinMillis" : 1641837000000
	} ],
	"measuredBy" : "global",
	"latestMeasuredByEntity" : null,
	"latestHealthScore" : 92,
	"monitoredDevices" : 402,
	"monitoredHealthyDevices" : 373,
	"monitoredUnHealthyDevices" : 29,
	"noHealthDevices" : 2,
	"healthDistirubution" : [ {
	  "category" : "Core",
	  "totalCount" : 1,
	  "healthScore" : 100,
	  "goodPercentage" : 100,
	  "badPercentage" : 0,
	  "fairPercentage" : 0,
	  "noHealthPercentage" : 0,
	  "goodCount" : 1,
	  "badCount" : 0,
	  "fairCount" : 0,
	  "noHealthCount" : 0
	}, {
	  "category" : "Access",
	  "totalCount" : 32,
	  "healthScore" : 100,
	  "goodPercentage" : 100,
	  "badPercentage" : 0,
	  "fairPercentage" : 0,
	  "noHealthPercentage" : 0,
	  "goodCount" : 32,
	  "badCount" : 0,
	  "fairCount" : 0,
	  "noHealthCount" : 0
	}, {
	  "category" : "Distribution",
	  "totalCount" : 5,
	  "healthScore" : 100,
	  "goodPercentage" : 100,
	  "badPercentage" : 0,
	  "fairPercentage" : 0,
	  "noHealthPercentage" : 0,
	  "goodCount" : 5,
	  "badCount" : 0,
	  "fairCount" : 0,
	  "noHealthCount" : 0
	}, {
	  "category" : "Router",
	  "totalCount" : 4,
	  "healthScore" : 100,
	  "goodPercentage" : 100,
	  "badPercentage" : 0,
	  "fairPercentage" : 0,
	  "noHealthPercentage" : 0,
	  "goodCount" : 4,
	  "badCount" : 0,
	  "fairCount" : 0,
	  "noHealthCount" : 0
	}, {
	  "category" : "WLC",
	  "totalCount" : 1,
	  "healthScore" : 100,
	  "goodPercentage" : 100,
	  "badPercentage" : 0,
	  "fairPercentage" : 0,
	  "noHealthPercentage" : 0,
	  "goodCount" : 1,
	  "badCount" : 0,
	  "fairCount" : 0,
	  "noHealthCount" : 0
	}, {
	  "category" : "AP",
	  "totalCount" : 361,
	  "healthScore" : 91,
	  "goodPercentage" : 91.41274,
	  "badPercentage" : 91.41274,
	  "fairPercentage" : 8.033241,
	  "noHealthPercentage" : 0.55401665,
	  "goodCount" : 330,
	  "badCount" : 0,
	  "fairCount" : 29,
	  "noHealthCount" : 2,
	  "kpiMetrics" : [ {
		"key" : "NOISE",
		"value" : "FAIR"
	  }, {
		"key" : "AIRQUALITY",
		"value" : "FAIR"
	  }, {
		"key" : "UTILIZATION",
		"value" : "FAIR"
	  }, {
		"key" : "INTERFERENCE",
		"value" : "FAIR"
	  } ]
	} ]
  }`

var network_health_fixture map[string]interface{}
