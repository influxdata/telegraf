package mailchimp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestMailChimpGatherReports(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprintln(w, sampleReports)
				require.NoError(t, err)
			},
		))
	defer ts.Close()

	u, err := url.ParseRequestURI(ts.URL)
	require.NoError(t, err)

	api := &ChimpAPI{
		url:   u,
		debug: true,
		log:   testutil.Logger{},
	}
	m := MailChimp{
		api: api,
	}

	var acc testutil.Accumulator
	err = m.Gather(&acc)
	require.NoError(t, err)

	tags := make(map[string]string)
	tags["id"] = "42694e9e57"
	tags["campaign_title"] = "Freddie's Jokes Vol. 1"

	fields := map[string]interface{}{
		"emails_sent":              200,
		"abuse_reports":            0,
		"unsubscribed":             2,
		"hard_bounces":             0,
		"soft_bounces":             2,
		"syntax_errors":            0,
		"forwards_count":           0,
		"forwards_opens":           0,
		"opens_total":              186,
		"unique_opens":             100,
		"clicks_total":             42,
		"unique_clicks":            400,
		"unique_subscriber_clicks": 42,
		"facebook_recipient_likes": 5,
		"facebook_unique_likes":    8,
		"facebook_likes":           42,
		"open_rate":                float64(42),
		"click_rate":               float64(42),
		"industry_open_rate":       float64(0.17076777144396),
		"industry_click_rate":      float64(0.027431311866951),
		"industry_bounce_rate":     float64(0.0063767751251474),
		"industry_unopen_rate":     float64(0.82285545343089),
		"industry_unsub_rate":      float64(0.001436957032815),
		"industry_abuse_rate":      float64(0.00021111996110887),
		"list_stats_sub_rate":      float64(10),
		"list_stats_unsub_rate":    float64(20),
		"list_stats_open_rate":     float64(42),
		"list_stats_click_rate":    float64(42),
		"industry_type":            "Social Networks and Online Communities",
	}
	acc.AssertContainsTaggedFields(t, "mailchimp", fields, tags)
}

func TestMailChimpGatherReport(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprintln(w, sampleReport)
				require.NoError(t, err)
			},
		))
	defer ts.Close()

	u, err := url.ParseRequestURI(ts.URL)
	require.NoError(t, err)

	api := &ChimpAPI{
		url:   u,
		debug: true,
		log:   testutil.Logger{},
	}
	m := MailChimp{
		api:        api,
		CampaignID: "test",
	}

	var acc testutil.Accumulator
	err = m.Gather(&acc)
	require.NoError(t, err)

	tags := make(map[string]string)
	tags["id"] = "42694e9e57"
	tags["campaign_title"] = "Freddie's Jokes Vol. 1"

	fields := map[string]interface{}{
		"emails_sent":              int(200),
		"abuse_reports":            int(0),
		"unsubscribed":             int(2),
		"hard_bounces":             int(0),
		"soft_bounces":             int(2),
		"syntax_errors":            int(0),
		"forwards_count":           int(0),
		"forwards_opens":           int(0),
		"opens_total":              int(186),
		"unique_opens":             int(100),
		"clicks_total":             int(42),
		"unique_clicks":            int(400),
		"unique_subscriber_clicks": int(42),
		"facebook_recipient_likes": int(5),
		"facebook_unique_likes":    int(8),
		"facebook_likes":           int(42),
		"open_rate":                float64(42),
		"click_rate":               float64(42),
		"industry_open_rate":       float64(0.17076777144396),
		"industry_click_rate":      float64(0.027431311866951),
		"industry_bounce_rate":     float64(0.0063767751251474),
		"industry_unopen_rate":     float64(0.82285545343089),
		"industry_unsub_rate":      float64(0.001436957032815),
		"industry_abuse_rate":      float64(0.00021111996110887),
		"list_stats_sub_rate":      float64(10),
		"list_stats_unsub_rate":    float64(20),
		"list_stats_open_rate":     float64(42),
		"list_stats_click_rate":    float64(42),
		"industry_type":            "Social Networks and Online Communities",
	}
	acc.AssertContainsTaggedFields(t, "mailchimp", fields, tags)
}

func TestMailChimpGatherError(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprintln(w, sampleError)
				require.NoError(t, err)
			},
		))
	defer ts.Close()

	u, err := url.ParseRequestURI(ts.URL)
	require.NoError(t, err)

	api := &ChimpAPI{
		url:   u,
		debug: true,
		log:   testutil.Logger{},
	}
	m := MailChimp{
		api:        api,
		CampaignID: "test",
	}

	var acc testutil.Accumulator
	err = m.Gather(&acc)
	require.Error(t, err)
}

var sampleReports = `
{
  "reports": [
    {
      "id": "42694e9e57",
      "campaign_title": "Freddie's Jokes Vol. 1",
      "type": "regular",
      "emails_sent": 200,
      "abuse_reports": 0,
      "unsubscribed": 2,
      "send_time": "2015-09-15T19:05:51+00:00",
      "bounces": {
        "hard_bounces": 0,
        "soft_bounces": 2,
        "syntax_errors": 0
      },
      "forwards": {
        "forwards_count": 0,
        "forwards_opens": 0
      },
      "opens": {
        "opens_total": 186,
        "unique_opens": 100,
        "open_rate": 42,
        "last_open": "2015-09-15T19:15:47+00:00"
      },
      "clicks": {
        "clicks_total": 42,
        "unique_clicks": 400,
        "unique_subscriber_clicks": 42,
        "click_rate": 42,
        "last_click": "2015-09-15T19:15:47+00:00"
      },
      "facebook_likes": {
        "recipient_likes": 5,
        "unique_likes": 8,
        "facebook_likes": 42
      },
      "industry_stats": {
        "type": "Social Networks and Online Communities",
        "open_rate": 0.17076777144396,
        "click_rate": 0.027431311866951,
        "bounce_rate": 0.0063767751251474,
        "unopen_rate": 0.82285545343089,
        "unsub_rate": 0.001436957032815,
        "abuse_rate": 0.00021111996110887
      },
      "list_stats": {
        "sub_rate": 10,
        "unsub_rate": 20,
        "open_rate": 42,
        "click_rate": 42
      },
      "timeseries": [
        {
          "timestamp": "2015-09-15T19:00:00+00:00",
          "emails_sent": 198,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-15T20:00:00+00:00",
          "emails_sent": 2,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-15T21:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-15T22:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-15T23:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T00:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T01:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T02:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T03:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T04:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T05:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T06:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T07:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T08:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T09:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T10:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T11:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T12:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T13:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T14:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T15:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T16:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T17:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        },
        {
          "timestamp": "2015-09-16T18:00:00+00:00",
          "emails_sent": 0,
          "unique_opens": 0,
          "recipients_clicks": 0
        }
      ],
      "share_report": {
        "share_url": "http://usX.vip-reports.net/reports/summary?u=xxxx&id=xxxx",
        "share_password": "freddielikesjokes"
      },
      "delivery_status": {
        "enabled": false
      },
      "_links": [
        {
          "rel": "parent",
          "href": "https://usX.api.mailchimp.com/3.0/reports",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Collection.json",
          "schema": "https://api.mailchimp.com/schema/3.0/CollectionLinks/Reports.json"
        },
        {
          "rel": "self",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Instance.json"
        },
        {
          "rel": "campaign",
          "href": "https://usX.api.mailchimp.com/3.0/campaigns/42694e9e57",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Campaigns/Instance.json"
        },
        {
          "rel": "sub-reports",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/sub-reports",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Sub/Collection.json"
        },
        {
          "rel": "abuse-reports",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/abuse-reports",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Abuse/Collection.json"
        },
        {
          "rel": "advice",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/advice",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Advice/Collection.json"
        },
        {
          "rel": "click-details",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/click-details",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/ClickDetails/Collection.json"
        },
        {
          "rel": "domain-performance",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/domain-performance",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/DomainPerformance/Collection.json"
        },
        {
          "rel": "eepurl",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/eepurl",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Eepurl/Collection.json"
        },
        {
          "rel": "email-activity",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/email-activity",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/EmailActivity/Collection.json"
        },
        {
          "rel": "locations",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/locations",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Locations/Collection.json"
        },
        {
          "rel": "sent-to",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/sent-to",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/SentTo/Collection.json"
        },
        {
          "rel": "unsubscribed",
          "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/unsubscribed",
          "method": "GET",
          "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Unsubs/Collection.json"
        }
      ]
    }
  ],
  "_links": [
    {
      "rel": "parent",
      "href": "https://usX.api.mailchimp.com/3.0/",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Root.json"
    },
    {
      "rel": "self",
      "href": "https://usX.api.mailchimp.com/3.0/reports",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Collection.json",
      "schema": "https://api.mailchimp.com/schema/3.0/CollectionLinks/Reports.json"
    }
  ],
  "total_items": 1
}
`

var sampleReport = `
{
  "id": "42694e9e57",
  "campaign_title": "Freddie's Jokes Vol. 1",
  "type": "regular",
  "emails_sent": 200,
  "abuse_reports": 0,
  "unsubscribed": 2,
  "send_time": "2015-09-15T19:05:51+00:00",
  "bounces": {
    "hard_bounces": 0,
    "soft_bounces": 2,
    "syntax_errors": 0
  },
  "forwards": {
    "forwards_count": 0,
    "forwards_opens": 0
  },
  "opens": {
    "opens_total": 186,
    "unique_opens": 100,
    "open_rate": 42,
    "last_open": "2015-09-15T19:15:47+00:00"
  },
  "clicks": {
    "clicks_total": 42,
    "unique_clicks": 400,
    "unique_subscriber_clicks": 42,
    "click_rate": 42,
    "last_click": "2015-09-15T19:15:47+00:00"
  },
  "facebook_likes": {
    "recipient_likes": 5,
    "unique_likes": 8,
    "facebook_likes": 42
  },
  "industry_stats": {
    "type": "Social Networks and Online Communities",
    "open_rate": 0.17076777144396,
    "click_rate": 0.027431311866951,
    "bounce_rate": 0.0063767751251474,
    "unopen_rate": 0.82285545343089,
    "unsub_rate": 0.001436957032815,
    "abuse_rate": 0.00021111996110887
  },
  "list_stats": {
    "sub_rate": 10,
    "unsub_rate": 20,
    "open_rate": 42,
    "click_rate": 42
  },
  "timeseries": [
    {
      "timestamp": "2015-09-15T19:00:00+00:00",
      "emails_sent": 198,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-15T20:00:00+00:00",
      "emails_sent": 2,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-15T21:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-15T22:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-15T23:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T00:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T01:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T02:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T03:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T04:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T05:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T06:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T07:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T08:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T09:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T10:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T11:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T12:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T13:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T14:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T15:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T16:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T17:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    },
    {
      "timestamp": "2015-09-16T18:00:00+00:00",
      "emails_sent": 0,
      "unique_opens": 0,
      "recipients_clicks": 0
    }
  ],
  "share_report": {
    "share_url": "http://usX.vip-reports.net/reports/summary?u=xxxx&id=xxxx",
    "share_password": "freddielikesjokes"
  },
  "delivery_status": {
    "enabled": false
  },
  "_links": [
    {
      "rel": "parent",
      "href": "https://usX.api.mailchimp.com/3.0/reports",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Collection.json",
      "schema": "https://api.mailchimp.com/schema/3.0/CollectionLinks/Reports.json"
    },
    {
      "rel": "self",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Instance.json"
    },
    {
      "rel": "campaign",
      "href": "https://usX.api.mailchimp.com/3.0/campaigns/42694e9e57",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Campaigns/Instance.json"
    },
    {
      "rel": "sub-reports",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/sub-reports",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Sub/Collection.json"
    },
    {
      "rel": "abuse-reports",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/abuse-reports",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Abuse/Collection.json"
    },
    {
      "rel": "advice",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/advice",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Advice/Collection.json"
    },
    {
      "rel": "click-details",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/click-details",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/ClickDetails/Collection.json"
    },
    {
      "rel": "domain-performance",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/domain-performance",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/DomainPerformance/Collection.json"
    },
    {
      "rel": "eepurl",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/eepurl",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Eepurl/Collection.json"
    },
    {
      "rel": "email-activity",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/email-activity",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/EmailActivity/Collection.json"
    },
    {
      "rel": "locations",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/locations",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Locations/Collection.json"
    },
    {
      "rel": "sent-to",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/sent-to",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/SentTo/Collection.json"
    },
    {
      "rel": "unsubscribed",
      "href": "https://usX.api.mailchimp.com/3.0/reports/42694e9e57/unsubscribed",
      "method": "GET",
      "targetSchema": "https://api.mailchimp.com/schema/3.0/Reports/Unsubs/Collection.json"
    }
  ]
}
`

var sampleError = `
{
    "type": "http://developer.mailchimp.com/documentation/mailchimp/guides/error-glossary/",
    "title": "API Key Invalid",
    "status": 401,
    "detail": "Your API key may be invalid, or you've attempted to access the wrong datacenter.",
    "instance": ""
}
`
