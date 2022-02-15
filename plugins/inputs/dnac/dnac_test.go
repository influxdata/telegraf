package dnac_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/dnac"
	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func Test_Gather(t *testing.T) {
	var err error

	plugin := dnac.NewDnac()

	plugin.DnacBaseURL = "https://192.168.196.2"
	plugin.Username = "test"
	plugin.Password = "test123"
	plugin.SSLVerify = "false"
	plugin.Report = []string{"client", "network"}
	plugin.Log = testutil.Logger{}

	require.NoError(t, plugin.InitClient())

	httpmock.ActivateNonDefault(plugin.Client.RestyClient().GetClient())

	authResponder, err := httpmock.NewJsonResponder(200, authFixture)
	if err != nil {
		t.Log("invalid auth fixture")
		t.FailNow()
	}
	authURL := "https://192.168.196.2/dna/system/api/v1/auth/token"
	httpmock.RegisterResponder("POST", authURL, authResponder)

	// Used json Decoder instead of Unmarshall to preserve 64-bit integers on 32-bit machines
	clientHealthDecoder := json.NewDecoder(bytes.NewBuffer([]byte(clientHealthString)))
	clientHealthDecoder.UseNumber()
	require.NoError(t, clientHealthDecoder.Decode(&clientHealthFixture))
	clientHealthResponder, err := httpmock.NewJsonResponder(200, clientHealthFixture)
	if err != nil {
		t.Log("invalid clientHealth fixture")
		t.FailNow()
	}
	clientHealthURL := "https://192.168.196.2/dna/intent/api/v1/client-health"
	httpmock.RegisterResponder("GET", clientHealthURL, clientHealthResponder)

	// Used json Decoder instead of Unmarshall to preserve 64-bit integers on 32-bit machines
	networkHealthDecoder := json.NewDecoder(bytes.NewBuffer([]byte(networkHealthString)))
	networkHealthDecoder.UseNumber()
	require.NoError(t, networkHealthDecoder.Decode(&networkHealthFixture))
	networkHealthResponder, err := httpmock.NewJsonResponder(200, networkHealthFixture)
	if err != nil {
		t.Log("invalid networkHealthFixture")
		t.FailNow()
	}
	networkHealthURL := "https://192.168.196.2/dna/intent/api/v1/network-health"
	httpmock.RegisterResponder("GET", networkHealthURL, networkHealthResponder)

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 2)
	clientHealthMetric := acc.Metrics[0]
	require.Len(t, clientHealthMetric.Fields, 21)
	require.Len(t, clientHealthMetric.Tags, 2)
	networkHealthMetric := acc.Metrics[1]
	require.Len(t, networkHealthMetric.Fields, 66)
	require.Len(t, networkHealthMetric.Tags, 2)
}

var authFixture = map[string]string{"Token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2MDY2NzE2NmU2YjNmNDAwYzJmMmNhOTciLCJhdXRoU291cmNlIjoiaW50ZXJuYWwiLCJ0ZW5hbnROYW1lIjoiVE5UMCIsInJvbGVzIjpbIjYwNjY3MTY1ZTZiM2Y0MDBjMmYyY2E5NiJdLCJ0ZW5hbnRJZCI6IjYwNjY3MTY1ZTZiM2Y0MDBjMmYyY2E5NCIsImV4cCI6MTY0MTgzNjIwNSwiaWF0IjoxNjQxODMyNjA1LCJqdGkiOiJkNDY5MWI4My1iZGIwLTQ3MWQtOTllZi1jOTQzZGQxMjA5ZTEiLCJ1c2VybmFtZSI6ImFkbWluIn0.HQCZ-EffKDgjb7jHmyHjfr1vNP6u4xBImcgS82AQK-4whsgIwyGu1xeJujT7zLRFdt-Ojz3pvRCiUhvXRaCIpSXnhR_s6otFJCsmZE_oJcFwVLWa7_7heWN4XAxjwUH37_2wiGRay4HFwkh85O4I_VW_NkrUfoLFeWpksBFX8edcbBLUC3iBtMG0wp-ogQdfebKxnpwaEy2o8iR38RhcAJDnOa4XCVGGDTTUZTsEID5yE_M_XRMdYkDA6g0FY3mbsj0VWI8fPELhHgjR39enLDOsykMefmquMkm7mqhufMslC5gls5MlLjmw53xzsAwmQe8NRYD9uuMrGEVFE0Ep3Q"}

var clientHealthString = `{
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
		"starttime" : 1641830700,
		"endtime" : 1641831000
	  }, {
		"scoreCategory" : {
		  "scoreCategory" : "CLIENT_TYPE",
		  "value" : "WIRED"
		},
		"scoreValue" : 99,
		"clientCount" : 1047,
		"clientUniqueCount" : 1047,
		"starttime" : 1641830700,
		"endtime" : 1641831000,
		"scoreList" : [ {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "POOR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 14,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 14,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "DHCP"
			},
			"scoreValue" : -1,
			"clientCount" : 14,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "FAIR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "GOOD"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 1033,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 1033,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "IDLE"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NODATA"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NEW"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000
		} ]
	  }, {
		"scoreCategory" : {
		  "scoreCategory" : "CLIENT_TYPE",
		  "value" : "WIRELESS"
		},
		"scoreValue" : 79,
		"clientCount" : 118,
		"clientUniqueCount" : 118,
		"starttime" : 1641830700,
		"endtime" : 1641831000,
		"scoreList" : [ {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "POOR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 9,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 9,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "ASSOCIATION"
			},
			"scoreValue" : -1,
			"clientCount" : 7,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  }, {
			"scoreCategory" : {
			  "scoreCategory" : "rootCause",
			  "value" : "AAA"
			},
			"scoreValue" : -1,
			"clientCount" : 2,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "FAIR"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 15,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 15,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "GOOD"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 92,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 92,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "IDLE"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 2,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 2,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NODATA"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 0,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000
		}, {
		  "scoreCategory" : {
			"scoreCategory" : "SCORE_TYPE",
			"value" : "NEW"
		  },
		  "scoreValue" : -1,
		  "clientCount" : 1,
		  "clientUniqueCount" : 0,
		  "starttime" : 1641830700,
		  "endtime" : 1641831000,
		  "scoreList" : [ {
			"scoreCategory" : {
			  "scoreCategory" : "deviceType",
			  "value" : "ALL"
			},
			"scoreValue" : -1,
			"clientCount" : 1,
			"clientUniqueCount" : null,
			"starttime" : 1641830700,
			"endtime" : 1641831000
		  } ]
		} ]
	  } ]
	} ]
  }`

var clientHealthFixture map[string]interface{}

var networkHealthString = `{
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
	  "timeinMillis" : 1641837000
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

var networkHealthFixture map[string]interface{}
