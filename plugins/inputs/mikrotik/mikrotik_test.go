package mikrotik

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var username = config.NewSecret([]byte("testOfTheMikrotik"))

func generateHandler() func(http.ResponseWriter, *http.Request) {
	crackers := map[string]string{}
	for k, v := range modules {
		crackers[v] = k
	}
	crackers["/rest/system/routerboard"] = "system_routerboard"
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := "./testData/" + crackers[r.URL.Path] + ".json"
		data, err := os.ReadFile(filePath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			_, err = w.Write(data)
			if err != nil {
				panic(err)
			}
		}
	}
}

func TestUptimeConverter(t *testing.T) {
	values := map[string]time.Duration{
		"1d0h0m0s":  time.Duration(24) * time.Hour,
		"0d0m0s":    0,
		"1s0m0s":    0,
		"0d0a0s":    0,
		"2m29s":     time.Duration(29)*time.Second + time.Duration(2)*time.Minute,
		"23m35s":    time.Duration(35)*time.Second + time.Duration(23)*time.Minute,
		"1d0h52m0s": time.Duration(24)*time.Hour + time.Duration(52)*time.Minute,
		"12s":       time.Duration(12) * time.Second,
		"121d12h12m0s": (time.Duration(24*121) * time.Hour) +
			(time.Duration(12) * time.Hour) +
			(time.Duration(12) * time.Minute) +
			(time.Duration(0) * time.Second),
	}
	for uptime, result := range values {
		d, err := parseUptimeIntoDuration(uptime)
		require.NoError(t, err)
		require.Equal(t, d, int64(result/time.Second))
	}
}

func TestMikrotikBaseTagsCollection(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:        fakeServer.URL,
		Username:       username,
		Log:            testutil.Logger{},
		IncludeModules: []string{"interface"},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))

	plugin = &Mikrotik{
		Address:        fakeServer.URL,
		Log:            testutil.Logger{},
		Username:       username,
		IncludeModules: []string{"interface"},
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start())
	require.NoError(t, acc.GatherError(plugin.Gather))
	var metric = acc.Metrics[0]
	require.Equal(t, "mikrotik", metric.Measurement)
}

func TestMikrotikCheckCorrectAmountOfMetricsWithOneEmptyAndOneDisabled(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:  fakeServer.URL,
		Log:      testutil.Logger{},
		Username: username,
		IncludeModules: []string{
			"interface",
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start())
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 2)
}

func TestMikrotikAllDataPoints(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:  fakeServer.URL,
		Username: username,
		Log:      testutil.Logger{},
		IncludeModules: []string{
			"interface",
			"interface_wireguard_peers",
			"interface_wireless_registration",
			"ip_dhcp_server_lease",
			"ip_firewall_connection",
			"ip_firewall_filter",
			"ip_firewall_nat",
			"ip_firewall_mangle",
			"ipv6_firewall_connection",
			"ipv6_firewall_filter",
			"ipv6_firewall_nat",
			"ipv6_firewall_mangle",
			"system_script",
			"system_resourses",
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start())
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 16)
}

func TestMikrotikConfigurationErrors(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:  fakeServer.URL,
		Username: username,
		Log:      testutil.Logger{},
		IncludeModules: []string{
			"interfacessss",
		},
	}
	require.True(t, strings.Contains(plugin.Init().Error(), "mikrotik init ->"))
	plugin.IncludeModules = []string{}
	require.NoError(t, plugin.Init())
}

func TestMikrotikDataIsCorrect(t *testing.T) {
	requiredFields := map[string]int64{
		"tx-packet":     56850710,
		"tx-error":      0,
		"fp-rx-packet":  144047662,
		"fp-tx-byte":    0,
		"rx-byte":       194632346152,
		"rx-error":      0,
		"rx-packet":     144047662,
		"fp-rx-byte":    194056155504,
		"tx-queue-drop": 0,
		"rx-drop":       0,
		"tx-byte":       15355309685,
		"tx-drop":       0,
		"fp-tx-packet":  0,
	}

	requiredTags := map[string]string{
		".id":               "*1",
		"architecture-name": "arm",
		"board-name":        "hAP",
		"cpu":               "ARM",
		"current-firmware":  "7.15.3",
		"default-name":      "ether1",
		"disabled":          "false",
		"firmware-type":     "ipq4000L",
		"mac-address":       "00:11:22:33:44:55",
		"model":             "RBD52G-5HacD2HnD",
		"name":              "ether1",
		"platform":          "MikroTik",
		"version":           "7.16 (stable)",
		"running":           "true",
		"serial-number":     "123456789",
		"type":              "ether",
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:  fakeServer.URL,
		Log:      testutil.Logger{},
		Username: username,
		IncludeModules: []string{
			"interface",
		},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start())
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 2)
	fields := acc.Metrics[0].Fields
	for k := range fields {
		require.Equal(t, requiredFields[k], fields[k])
		delete(requiredFields, k)
	}
	require.Empty(t, requiredFields)

	tags := acc.Metrics[0].Tags
	for k := range tags {
		// this is a workaround to not filter tags in
		// getSystemTags because the issue is only in tests
		if _, ok := requiredTags[k]; ok {
			require.Equal(t, requiredTags[k], tags[k])
			delete(requiredTags, k)
		}
	}
	require.Empty(t, requiredTags)
}

func TestMikrotikCommentExclusion(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	defer fakeServer.Close()
	plugin := &Mikrotik{
		Address:  fakeServer.URL,
		Log:      testutil.Logger{},
		Username: username,
		IncludeModules: []string{"interface",
			"interface_wireguard_peers",
			"interface_wireless_registration",
			"ip_dhcp_server_lease",
			"ip_firewall_connection",
			"ip_firewall_filter",
			"ip_firewall_nat",
			"ip_firewall_mangle",
			"ipv6_firewall_connection",
			"ipv6_firewall_filter",
			"ipv6_firewall_nat",
			"ipv6_firewall_mangle",
			"system_script",
			"system_resourses"},
		IgnoreComments: []string{"ignoreThis"},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start())
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.Len(t, acc.Metrics, 15)
}
