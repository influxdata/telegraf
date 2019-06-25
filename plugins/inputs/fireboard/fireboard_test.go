package fireboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestFireboard(t *testing.T) {
	// Create a test server with the const response JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	// Parse the URL of the test server, used to verify the expected host
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)

	// Create a new Riak instance with our given test server
	fireboard := NewFireboard()
	fireboard.AuthToken = []string{ts.URL}

	// Create a test accumulator
	acc := &testutil.Accumulator{}

	// Gather data from the test server
	err = fireboard.Gather(acc)
	require.NoError(t, err)

	// Expect the correct values for all known keys
	expectFields := map[string]interface{}{
		"cpu_avg1":  int64(504),
		"cpu_avg15": int64(294),
		"cpu_avg5":  int64(325),
	}

	// Expect the correct values for all tags
	expectTags := map[string]string{
		"title": "telegraf-FireBoard",
		"uuid":  "5c597f9b-9914-4586-a35c-5d0c973fb542",
	}

	acc.AssertContainsTaggedFields(t, "riak", expectFields, expectTags)
}

var response = `
{
	"id": 16480,
	"title": "ronnoco-FireBoard",
	"owner": {
	  "username": "lance21",
	  "email": "lance@ronnoco.net",
	  "first_name": "Lance",
	  "last_name": "O'Connor",
	  "id": 15814,
	  "userprofile": {
		"company": "",
		"alert_sms": "4082507483",
		"alert_emails": "lance@ronnoco.net",
		"notification_tone": "default",
		"user": 15814,
		"picture": "/media/profile_images/default-profile.png",
		"last_templog": "2019-06-25T06:06:40Z",
		"commercial_user": false
	  }
	},
	"created": "2019-03-23T16:48:32.152010Z",
	"uuid": "5c597f9b-9914-4586-a35c-5d0c973fb542",
	"hardware_id": "GG7J97JF2",
	"latest_temps": [
	  {
		"temp": 79.9,
		"channel": 1,
		"degreetype": 2,
		"created": "2019-06-25T06:07:10Z"
	  },
	  {
		"temp": 80.9,
		"channel": 6,
		"degreetype": 2,
		"created": "2019-06-25T06:07:10Z"
	  }
	],
	"device_log": {
	  "deviceID": "5c597f9b-9914-4586-a35c-5d0c973fb542",
	  "cpuUsage": "19%",
	  "onboardTemp": 84.551974522293,
	  "mode": "Managed",
	  "boardID": "GG7J97JF2",
	  "frequency": "2.4 GHz",
	  "vBatt": 3.894,
	  "versionNode": "5.0.0",
	  "signallevel": -38,
	  "vBattPer": 0.776284064923077,
	  "versionUtils": "6.1.3",
	  "publicIP": "98.234.104.118",
	  "tempFilter": "true",
	  "commercialMode": "false",
	  "bleSignalLevel": 0,
	  "versionJava": "7.6.5",
	  "versionEspHal": "HAL: V1R2;AVR: 0.0.14;",
	  "internalIP": "10.100.252.138",
	  "date": "2019-06-25 06:04:40 UTC",
	  "vBattPerRaw": 0.685,
	  "bleClientMAC": "",
	  "auxPort": "",
	  "linkquality": "100/100",
	  "macAP": "92:3b:ad:31:0b:b2",
	  "contrast": "7",
	  "memUsage": "2.6M/4.2M",
	  "diskUsage": "1.4M/16.0M",
	  "txpower": 78,
	  "macNIC": "30:ae:a4:c6:4c:90",
	  "model": "FBX11D",
	  "ssid": "ronnoco-n",
	  "version": "0.3.1",
	  "uptime": "0:05",
	  "band": "802.11bgn",
	  "versionImage": "201703260554",
	  "drivesettings": "{\"p\":4,\"s\":1,\"d\":7,\"ms\":100,\"f\":0,\"l\":1}"
	},
	"last_templog": "2019-06-25T06:06:40Z",
	"channels": [
	  {
		"id": 2694431,
		"channel_label": "Channel 1",
		"channel": 1,
		"current_temp": 79.9,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": 80,
		"range_max_temp": 80.4,
		"range_min_temp": 79.8,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": {
		  "temp": 79.9,
		  "degreetype": 2,
		  "created": "2019-06-25T06:07:10Z"
		}
	  },
	  {
		"id": 2694432,
		"channel_label": "Channel 2",
		"channel": 2,
		"current_temp": null,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": null,
		"range_max_temp": null,
		"range_min_temp": null,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": null
	  },
	  {
		"id": 2694433,
		"channel_label": "Channel 3",
		"channel": 3,
		"current_temp": null,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": null,
		"range_max_temp": null,
		"range_min_temp": null,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": null
	  },
	  {
		"id": 2694434,
		"channel_label": "Channel 4",
		"channel": 4,
		"current_temp": null,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": null,
		"range_max_temp": null,
		"range_min_temp": null,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": null
	  },
	  {
		"id": 2694435,
		"channel_label": "Channel 5",
		"channel": 5,
		"current_temp": null,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": null,
		"range_max_temp": null,
		"range_min_temp": null,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": null
	  },
	  {
		"id": 2694436,
		"channel_label": "Channel 6",
		"channel": 6,
		"current_temp": 80.9,
		"created": "2019-06-25T05:59:37Z",
		"enabled": true,
		"notify_email": true,
		"notify_sms": true,
		"temp_max": null,
		"temp_min": null,
		"minutes_buffer": null,
		"range_average_temp": 80.8,
		"range_max_temp": 80.9,
		"range_min_temp": 80.7,
		"alerts": [],
		"sessionid": 451804,
		"last_templog": {
		  "temp": 80.9,
		  "degreetype": 2,
		  "created": "2019-06-25T06:07:10Z"
		}
	  }
	],
	"model": "FBX11E",
	"channel_count": 6,
	"degreetype": 2
  }
`
