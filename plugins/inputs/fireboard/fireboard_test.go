package fireboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestFireboard(t *testing.T) {
	// Create a test server with the const response JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Telegraf", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))
	defer ts.Close()

	// Parse the URL of the test server, used to verify the expected host
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)

	// Create a new fb instance with our given test server
	fireboard := newFireboard()
	fireboard.AuthToken = "b4bb6e6a7b6231acb9f71b304edb2274693d8849"
	fireboard.URL = u.String()

	// Create a test accumulator
	acc := &testutil.Accumulator{}

	// Gather data from the test server
	err = fireboard.Gather(acc)
	require.NoError(t, err)

	// --- FBX11E device: channel 1 with a matching channel entry ---
	expectFields := map[string]interface{}{
		"temperature": float64(79.9),
	}
	expectTags := map[string]string{
		"title":        "telegraf-FireBoard",
		"uuid":         "b55e766c-b308-49b5-93a4-df89fe31efd0",
		"channel":      strconv.FormatInt(1, 10),
		"scale":        "Fahrenheit",
		"channel_name": "Instant Probe",
	}
	acc.AssertContainsTaggedFields(t, "fireboard", expectFields, expectTags)

	// --- FBXIR device: channel 1 has no matching channel entry → empty channel_name ---
	expectTagsSpark := map[string]string{
		"title":        "spark",
		"uuid":         "cccccccc-cccc-cccc-cccc-cccccccccccc",
		"channel":      strconv.FormatInt(1, 10),
		"scale":        "Fahrenheit",
		"channel_name": "",
	}
	expectFieldsSpark := map[string]interface{}{
		"temperature": float64(225.0),
	}
	acc.AssertContainsTaggedFields(t, "fireboard", expectFieldsSpark, expectTagsSpark)

	// --- fireboard_drive: FBXPD device has last_drivelog with jsonraw ---
	driveTags := map[string]string{
		"title":       "pellet-grill",
		"uuid":        "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"model":       "FBXPD",
		"pel_state":   "RUNNING",
		"pel_profile": "Hickory",
	}
	driveFields := map[string]interface{}{
		"drive_percent":    float64(42.5),
		"setpoint":         float64(350.0),
		"current_temp":     float64(348.2),
		"vbatt":            float64(12.3),
		"power_mode":       int(1),
		"mode_type":        int(2),
		"drive_type":       int(1),
		"tied_channel":     int64(1),
		"fan_speed_target": float64(60.0),
		"lid_paused":       false,
		"pid_output":       float64(42.5),
		"pid_error_sum":    float64(1.2),
		"pid_kd_output":    float64(0.1),
		"pid_kp_output":    float64(40.0),
		"pid_ki_output":    float64(2.4),
		"pid_last_error":   float64(-1.8),
		"pid_kp":           float64(10.0),
		"pid_ki":           float64(2.0),
		"pid_kd":           float64(0.5),
	}
	acc.AssertContainsTaggedFields(t, "fireboard_drive", driveFields, driveTags)

	// fireboard_drive must be emitted exactly once (FBXPD only; FBX11E and FBXIR have null last_drivelog)
	var driveCount int
	for _, m := range acc.Metrics {
		if m.Measurement == "fireboard_drive" {
			driveCount++
		}
	}
	require.Equal(t, 1, driveCount, "expected exactly one fireboard_drive metric (FBXPD only)")

	// --- fireboard_alert: only channels with active latest_temps are emitted ---
	// FBX11E has 4 channels with alerts but only channel 1 in latest_temps,
	// so channels 2-4 must be suppressed (device off / probes disconnected).
	acc.AssertContainsTaggedFields(t, "fireboard_alert",
		map[string]interface{}{"enabled": true, "temp_max": float64(205.0)},
		map[string]string{
			"title": "telegraf-FireBoard", "uuid": "b55e766c-b308-49b5-93a4-df89fe31efd0",
			"channel": "1", "channel_name": "Instant Probe", "scale": "Fahrenheit",
		})
	// Total: exactly 1 (channels 2-4 suppressed; FBXPD has empty alerts; FBXIR has no channels)
	var alertCount int
	for _, m := range acc.Metrics {
		if m.Measurement == "fireboard_alert" {
			alertCount++
		}
	}
	require.Equal(t, 1, alertCount, "expected exactly 1 fireboard_alert metric (only active channels)")

	// --- fireboard_device: FBXPD has device_log; FBX11E and FBXIR do not ---
	acc.AssertContainsTaggedFields(t, "fireboard_device",
		map[string]interface{}{
			"onboard_temp":  float64(95.0),
			"signal_level":  int(-65),
			"vbatt":         float64(12.3),
			"vbatt_percent": float64(0.85),
		},
		map[string]string{
			"title": "pellet-grill",
			"uuid":  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"model": "FBXPD",
		})
	var deviceCount int
	for _, m := range acc.Metrics {
		if m.Measurement == "fireboard_device" {
			deviceCount++
		}
	}
	require.Equal(t, 1, deviceCount, "expected exactly one fireboard_device metric (FBXPD only)")
}

// response is a 3-device anonymized fixture:
//   - telegraf-FireBoard  (FBX11E): channel 1 with label; 4 channels with alerts; no drive/device_log
//   - pellet-grill        (FBXPD):  channels array; last_drivelog + device_log present
//   - spark               (FBXIR):  empty channels array; no drive/device_log
var response = `
[
  {
    "id": 99999,
    "title": "telegraf-FireBoard",
    "created": "2019-03-23T16:48:32.152010Z",
    "uuid": "b55e766c-b308-49b5-93a4-df89fe31efd0",
    "hardware_id": "XXXXXXXXX",
    "model": "FBX11E",
    "channel_count": 6,
    "degreetype": 2,
    "last_templog": "2019-06-25T06:06:40Z",
    "latest_temps": [
      {
        "temp": 79.9,
        "channel": 1,
        "degreetype": 2,
        "created": "2019-06-25T06:07:10Z"
      }
    ],
    "channels": [
      {
        "channel": 1,
        "channel_label": "Instant Probe",
        "degreetype": 2,
        "alerts": [
          {"channel": 1, "temp_max": 205.0, "temp_min": null, "enabled": true}
        ]
      },
      {
        "channel": 2,
        "channel_label": "Meat Probe",
        "degreetype": 2,
        "alerts": [
          {"channel": 2, "temp_max": 165.0, "temp_min": null, "enabled": true}
        ]
      },
      {
        "channel": 3,
        "channel_label": "Ambient",
        "degreetype": 2,
        "alerts": [
          {"channel": 3, "temp_max": 450.0, "temp_min": 200.0, "enabled": false}
        ]
      },
      {
        "channel": 4,
        "channel_label": "Grill Temp",
        "degreetype": 2,
        "alerts": [
          {"channel": 4, "temp_max": null, "temp_min": 180.0, "enabled": true}
        ]
      }
    ],
    "last_drivelog": null,
    "device_log": null
  },
  {
    "id": 88888,
    "title": "pellet-grill",
    "created": "2021-01-01T00:00:00Z",
    "uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "hardware_id": "YYYYYYYYY",
    "model": "FBXPD",
    "channel_count": 4,
    "degreetype": 2,
    "last_templog": "2021-07-04T12:00:00Z",
    "latest_temps": [
      {
        "temp": 350.0,
        "channel": 1,
        "degreetype": 2,
        "created": "2021-07-04T12:00:00Z"
      }
    ],
    "channels": [
      {
        "channel": 1,
        "channel_label": "Grill",
        "degreetype": 2,
        "alerts": []
      }
    ],
    "last_drivelog": {
      "driveper": 42.5,
      "setpoint": 350.0,
      "currenttemp": 348.2,
      "vbatt": 12.3,
      "powermode": 1,
      "modetype": 2,
      "drivetype": 1,
      "tiedchannel": 1,
      "jsonraw": "{\"pelState\":\"RUNNING\",\"fanSpeedTarget\":60.0,\"lidPaused\":false,\"pelProfile\":\"Hickory\",\"pOut\":42.5,\"pErrSum\":1.2,\"pKdOut\":0.1,\"pKpOut\":40.0,\"pKiOut\":2.4,\"pLastErr\":-1.8,\"pKp\":10.0,\"pKi\":2.0,\"pKd\":0.5}"
    },
    "device_log": {
      "onboardTemp": 95.0,
      "signallevel": -65,
      "vBatt": 12.3,
      "vBattPer": 0.85
    }
  },
  {
    "id": 77777,
    "title": "spark",
    "created": "2022-06-01T00:00:00Z",
    "uuid": "cccccccc-cccc-cccc-cccc-cccccccccccc",
    "hardware_id": "ZZZZZZZZZ",
    "model": "FBXIR",
    "channel_count": 1,
    "degreetype": 2,
    "last_templog": "2022-06-01T10:00:00Z",
    "latest_temps": [
      {
        "temp": 225.0,
        "channel": 1,
        "degreetype": 2,
        "created": "2022-06-01T10:00:00Z"
      }
    ],
    "channels": [],
    "last_drivelog": null,
    "device_log": null
  }
]
`
