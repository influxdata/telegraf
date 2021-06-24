// +build !windows

package varnish_cache

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func newTestServer() *VarnishCache {
	server := &VarnishCache{
		Log: testutil.Logger{},
	}
	return server
}

func TestVarnishCachePlugin(t *testing.T) {
	server := newTestServer()
	require.Equal(t, "A plugin to collect stats from Varnish HTTP Cache", server.Description())
	acc := &testutil.Accumulator{}

	require.Equal(t, 0, len(acc.Metrics))

	measurement := extractMeasurement("VBE.boot.default.fail")
	measurement2, field, tags := createMetric("VBE.boot.default.fail")
	require.Equal(t, measurement2, measurement)
	require.Equal(t, "varnish_vbe", measurement)
	require.Equal(t, "fail", field)
	require.Equal(t, map[string]string{"backend": "default"}, tags)
}

func TestParseVarnishNames(t *testing.T) {
	type testConfig struct {
		vName       string
		measurement string
		tags        map[string]string
		field       string
	}

	for _, c := range []testConfig{
		{
			vName:       "VBE.boot.default.fail",
			measurement: "varnish_vbe",
			tags:        map[string]string{"backend": "default"},
			field:       "fail",
		},
		{
			vName:       "MEMPOOL.req1.allocs",
			measurement: "varnish_mempool",
			tags:        map[string]string{"id": "req1"},
			field:       "allocs",
		},
		{
			vName:       "SMF.s0.c_bytes",
			measurement: "varnish_smf",
			tags:        map[string]string{"id": "s0"},
			field:       "c_bytes",
		},
		{
			vName:       "VBE.reload_20210622_153544_23757.server1.happy",
			measurement: "varnish_vbe",
			tags:        map[string]string{"backend": "server1"},
			field:       "happy",
		},
		{
			vName:       "XXX.YYY.XXX",
			measurement: "varnish_xxx",
			tags:        map[string]string{"id": "yyy"},
			field:       "xxx",
		},
	} {
		measurement, field, tags := createMetric(c.vName)
		require.Equal(t, c.measurement, measurement, c.vName)
		require.Equal(t, c.field, field)
		require.Equal(t, c.tags, tags)
	}
}

func getJson(fileName string, v interface{}) error {
	f, _ := os.Open("test_data/" + fileName)
	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("invalid json file " + fileName)
	}
	err := f.Close()
	return err
}
func TestReloadPrefix(t *testing.T) {
	server := newTestServer()
	require.Equal(t, "A plugin to collect stats from Varnish HTTP Cache", server.Description())
	acc := &testutil.Accumulator{}

	require.Equal(t, 0, len(acc.Metrics))

	type testConfig struct {
		jsonFile           string
		activeReloadPrefix string
	}

	for _, c := range []testConfig{
		{jsonFile: "varnish6.2.1_reload.json", activeReloadPrefix: "VBE.reload_20210623_170621_31083"},
		{jsonFile: "varnish6.6.json", activeReloadPrefix: ""},
		{jsonFile: "varnish4_4.json", activeReloadPrefix: ""},
	} {
		var rootJSON map[string]interface{}
		e := getJson(c.jsonFile, &rootJSON)
		require.NoError(t, e)
		countersJSON, e := getCountersJSON(rootJSON)
		require.NoError(t, e)

		recentPrefix := findActiveReloadPrefix(countersJSON)
		require.Equal(t, c.activeReloadPrefix, recentPrefix)

		err := server.processJSON(acc, rootJSON)
		require.NoError(t, err)

		for _, m := range acc.Metrics {
			require.NotEmpty(t, m.Fields)
			require.True(t, strings.HasPrefix(m.Measurement, "varnish_"))
			require.NotContains(t, "reload_", m.Measurement)
			for field := range m.Fields {
				require.NotContains(t, "reload_", field)
			}
			for tag := range m.Tags {
				require.NotContains(t, "reload_", tag)
			}
		}
	}
}
func TestVBEMetricParsing(t *testing.T) {
	type testConfig struct {
		vName       string
		backend     string
		server      string
		field       string
		measurement string
		id          string
	}
	for _, c := range []testConfig{
		//old varnish 4.x
		{vName: "VBE.aa_b.c.-d:d(10.100.0.108,,8080).happy", backend: "aa_b.c.-d:d", server: "10.100.0.108:8080", field: "happy", measurement: "vbe"},
		{vName: "VBE.root:aa2_b.c-d:e.happy", backend: "aa2_b.c-d:e", server: "", field: "happy", measurement: "vbe"},
		{vName: "VBE.34f022e6-55f3-4167-b213-95fac1921a0e.aa1_x.y-z:8080.happy", backend: "aa1_x.y-z:8080", server: "34f022e6-55f3-4167-b213-95fac1921a0e", field: "happy", measurement: "vbe"},
		{vName: "VBE.root:34f022e6-55f3-4167-b213-95fac1921a0e.aa1_x.y-z:w.happy", backend: "aa1_x.y-z:w", server: "34f022e6-55f3-4167-b213-95fac1921a0e", field: "happy", measurement: "vbe"},
		{vName: "VBE.boot.default.happy", backend: "default", server: "", field: "happy", measurement: "vbe"},
		// varnish reload old
		{vName: "VBE.reload_2021-01-29T100458.default.happy", backend: "default", server: "", field: "happy", measurement: "vbe"}, // varnish_reload_vcl in 4
		// varnish reload 6+
		{vName: "VBE.reload_20210622_153544_23757.default.happy", backend: "default", server: "", field: "happy", measurement: "vbe"},                                              // varnishreload in 6+
		{vName: "VBE.root:34f022e6-55f3-4167-b213-95fac1921a0e.x_y_z.happy", backend: "x_y_z", server: "34f022e6-55f3-4167-b213-95fac1921a0e", field: "happy", measurement: "vbe"}, // varnishreload in 6+
		{vName: "VBE.34f022e6-55f3-4167-b213-95fac1921a0e.default.happy", backend: "default", server: "34f022e6-55f3-4167-b213-95fac1921a0e", field: "happy", measurement: "vbe"},
		{vName: "LCK.vbe.creat", backend: "", server: "", field: "creat", id: "vbe", measurement: "lck"},
		{vName: "LCK.tcp_pool.dbg_busy", backend: "", server: "", field: "dbg_busy", id: "tcp_pool", measurement: "lck"},
	} {
		measurement, field, tags := createMetric(c.vName)
		require.Equal(t, "varnish_"+c.measurement, measurement)
		require.Equal(t, c.field, field)
		require.Equal(t, c.backend, tags["backend"])
		require.Equal(t, c.id, tags["id"])
		require.Equal(t, c.server, tags["server"])
	}
}
