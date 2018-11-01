package jira_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/jira"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	js := `
{
  "expand": "schema,names",
  "startAt": 0,
  "maxResults": 5,
  "total": 13544,
  "issues": [
    {
      "expand": "operations,versionedRepresentations,editmeta,changelog,renderedFields",
      "id": "90375",
      "self": "https://jira.local/rest/api/latest/issue/90375",
      "key": "PROJECT-20001622",
      "fields": {
        "priority": {
          "self": "https://jira.local/rest/api/2/priority/6",
          "iconUrl": "https://jira.local/images/icons/priorities/critical.svg",
          "name": "Express",
          "id": "6"
        },
        "customfield_10510": {
          "self": "https://jira.local/rest/api/2/customFieldOption/10583",
          "value": "DevTeam",
          "id": "10583"
        }
      }
    },
    {
      "expand": "operations,versionedRepresentations,editmeta,changelog,renderedFields",
      "id": "89394",
      "self": "https://jira.local/rest/api/latest/issue/89394",
      "key": "PROJECT-20001551",
      "fields": {
        "priority": {
          "self": "https://jira.local/rest/api/2/priority/1",
          "iconUrl": "https://jira.local/images/icons/priorities/major.svg",
          "name": "High",
          "id": "1"
        },
        "customfield_10510": {
          "self": "https://jira.local/rest/api/2/customFieldOption/10583",
          "value": "DevTeam",
          "id": "10583"
        }
      }
    },
    {
      "expand": "operations,versionedRepresentations,editmeta,changelog,renderedFields",
      "id": "88845",
      "self": "https://jira.local/rest/api/latest/issue/88845",
      "key": "PROJECT-20001492",
      "fields": {
        "priority": {
          "self": "https://jira.local/rest/api/2/priority/6",
          "iconUrl": "https://jira.local/images/icons/priorities/critical.svg",
          "name": "Express",
          "id": "6"
        },
        "customfield_10510": {
          "self": "https://jira.local/rest/api/2/customFieldOption/10583",
          "value": "TestingTeam",
          "id": "10583"
        }
      }
    },
    {
      "expand": "operations,versionedRepresentations,editmeta,changelog,renderedFields",
      "id": "88541",
      "self": "https://jira.local/rest/api/latest/issue/88541",
      "key": "PROJECT-20001460",
      "fields": {
        "priority": {
          "self": "https://jira.local/rest/api/2/priority/1",
          "iconUrl": "https://jira.local/images/icons/priorities/major.svg",
          "name": "Normal",
          "id": "1"
        },
        "customfield_10510": {
          "self": "https://jira.local/rest/api/2/customFieldOption/10583",
          "value": "TestingTeam",
          "id": "10583"
        }
      }
    },
    {
      "expand": "operations,versionedRepresentations,editmeta,changelog,renderedFields",
      "id": "88410",
      "self": "https://jira.local/rest/api/latest/issue/88410",
      "key": "PROJECT-20001448",
      "fields": {
        "priority": {
          "self": "https://jira.local/rest/api/2/priority/1",
          "iconUrl": "https://jira.local/images/icons/priorities/major.svg",
          "name": "Normal",
          "id": "1"
        },
        "customfield_10510": {
          "self": "https://jira.local/rest/api/2/customFieldOption/10583",
          "value": "DevTeam",
          "id": "10583"
        }
      }
    }
  ]
}
`
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/latest/search" {
			_, _ = w.Write([]byte(js))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &jira.Jira{
		Servers: []string{fakeServer.URL},
		Fields:  []string{"priority"},
		Tags:    []string{"customfield_10510"},
		Jql:     []jira.Jql{{Key: "new", Value: "Jql"}, {Key: "total", Value: "Jql"}},
	}

	var acc testutil.Accumulator
	acc.SetDebug(true)
	require.NoError(t, acc.GatherError(plugin.Gather))
}
