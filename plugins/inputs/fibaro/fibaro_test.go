package fibaro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sectionsJSON = `
    [
        {
            "id": 1,
            "name": "Section 1",
            "sortOrder": 1
        },
        {
            "id": 2,
            "name": "Section 2",
            "sortOrder": 2
        },
        {
            "id": 3,
            "name": "Section 3",
            "sortOrder": 3
        }
    ]`

const roomsJSON = `
    [
        {
            "id": 1,
            "name": "Room 1",
            "sectionID": 1,
            "icon": "room_1",
            "sortOrder": 1
        },
        {
            "id": 2,
            "name": "Room 2",
            "sectionID": 2,
            "icon": "room_2",
            "sortOrder": 2
        },
        {
            "id": 3,
            "name": "Room 3",
            "sectionID": 3,
            "icon": "room_3",
            "sortOrder": 3
        },
        {
            "id": 4,
            "name": "Room 4",
            "sectionID": 3,
            "icon": "room_4",
            "sortOrder": 4
        }
    ]`

const devicesJSON = `
    [
        {
            "id": 1,
            "name": "Device 1",
            "roomID": 1,
            "type": "com.fibaro.binarySwitch",
            "enabled": true,
            "properties": {
                "dead": "false",
                "value": "false"
            },
            "sortOrder": 1
        },
        {
            "id": 2,
            "name": "Device 2",
            "roomID": 2,
            "type": "com.fibaro.binarySwitch",
            "enabled": true,
            "properties": {
                "dead": "false",
                "value": "true"
            },
            "sortOrder": 2
        },
        {
            "id": 3,
            "name": "Device 3",
            "roomID": 3,
            "type": "com.fibaro.multilevelSwitch",
            "enabled": true,
            "properties": {
                "dead": "false",
                "value": "67"
            },
            "sortOrder": 3
        },
        {
            "id": 4,
            "name": "Device 4",
            "roomID": 4,
            "type": "com.fibaro.temperatureSensor",
            "enabled": true,
            "properties": {
                "dead": "false",
                "value": "22.80"
            },
            "sortOrder": 4
        },
        {
            "id": 5,
            "name": "Device 5",
            "roomID": 4,
            "type": "com.fibaro.FGRM222",
            "enabled": true,
            "properties": {
                "energy": "4.33",
                "power": "0.7",
                "dead": "false",
                "value": "50",
                "value2": "75"
            },
            "sortOrder": 5
        }
    ]`

// TestUnauthorized validates that 401 (wrong credentials) is managed properly
func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	a := Fibaro{
		URL:      ts.URL,
		Username: "user",
		Password: "pass",
		client:   &http.Client{},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.Error(t, err)
}

// TestJSONSuccess validates that module works OK with valid JSON payloads
func TestJSONSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := ""
		switch r.URL.Path {
		case "/api/sections":
			payload = sectionsJSON
		case "/api/rooms":
			payload = roomsJSON
		case "/api/devices":
			payload = devicesJSON
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, payload)
	}))
	defer ts.Close()

	a := Fibaro{
		URL:      ts.URL,
		Username: "user",
		Password: "pass",
		client:   &http.Client{},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)

	// Gather should add 5 metrics
	assert.Equal(t, uint64(5), acc.NMetrics())

	// Ensure fields / values are correct - Device 1
	tags := map[string]string{"deviceId": "1", "section": "Section 1", "room": "Room 1", "name": "Device 1", "type": "com.fibaro.binarySwitch"}
	fields := map[string]interface{}{"value": float64(0)}
	acc.AssertContainsTaggedFields(t, "fibaro", fields, tags)

	// Ensure fields / values are correct - Device 2
	tags = map[string]string{"deviceId": "2", "section": "Section 2", "room": "Room 2", "name": "Device 2", "type": "com.fibaro.binarySwitch"}
	fields = map[string]interface{}{"value": float64(1)}
	acc.AssertContainsTaggedFields(t, "fibaro", fields, tags)

	// Ensure fields / values are correct - Device 3
	tags = map[string]string{"deviceId": "3", "section": "Section 3", "room": "Room 3", "name": "Device 3", "type": "com.fibaro.multilevelSwitch"}
	fields = map[string]interface{}{"value": float64(67)}
	acc.AssertContainsTaggedFields(t, "fibaro", fields, tags)

	// Ensure fields / values are correct - Device 4
	tags = map[string]string{"deviceId": "4", "section": "Section 3", "room": "Room 4", "name": "Device 4", "type": "com.fibaro.temperatureSensor"}
	fields = map[string]interface{}{"value": float64(22.8)}
	acc.AssertContainsTaggedFields(t, "fibaro", fields, tags)

	// Ensure fields / values are correct - Device 5
	tags = map[string]string{"deviceId": "5", "section": "Section 3", "room": "Room 4", "name": "Device 5", "type": "com.fibaro.FGRM222"}
	fields = map[string]interface{}{"energy": float64(4.33), "power": float64(0.7), "value": float64(50), "value2": float64(75)}
	acc.AssertContainsTaggedFields(t, "fibaro", fields, tags)
}
