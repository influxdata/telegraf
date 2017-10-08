package filestat

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGatherNoMd5(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Files = []string{
		dir + "log1.log",
		dir + "log2.log",
		"/non/existant/file",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	var testMetrics = []struct {
		Tags   map[string]string
		Fields map[string]interface{}
	}{
		{
			map[string]string{"file": dir + "log1.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
			},
		},
		{
			map[string]string{"file": dir + "log2.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
			},
		},
		{
			map[string]string{"file": "/non/existant/file"},
			map[string]interface{}{
				"exists": int64(0),
			},
		},
	}

	for _, m := range acc.Metrics {
		foundTags := false
		for _, tm := range testMetrics {
			if reflect.DeepEqual(tm.Tags, m.Tags) {

				for k, v := range tm.Fields {
					if m.Fields[k] != v {
						t.Errorf("\nFailed on %s Field\n\texpected\t%+v\n\treceived\t%+v\n", k, v, m.Fields[k])
					}
				}

				if m.Fields["exists"] == 1 {
					if _, ok := m.Fields["modification_time"].(int64); !ok {
						t.Error("modification_time Field is not set")
					}
				}

				foundTags = true
				break
			}
		}
		if !foundTags {
			t.Errorf("\nFailed, could not find matching tags\n\t%+v", m.Tags)
		}
	}
}

func TestGatherExplicitFiles(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "log1.log",
		dir + "log2.log",
		"/non/existant/file",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	var testMetrics = []struct {
		Tags   map[string]string
		Fields map[string]interface{}
	}{
		{
			map[string]string{"file": dir + "log1.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		{
			map[string]string{"file": dir + "log2.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		{
			map[string]string{"file": "/non/existant/file"},
			map[string]interface{}{
				"exists": int64(0),
			},
		},
	}

	for _, m := range acc.Metrics {
		foundTags := false
		for _, tm := range testMetrics {
			if reflect.DeepEqual(tm.Tags, m.Tags) {

				for k, v := range tm.Fields {
					if m.Fields[k] != v {
						t.Errorf("\nFailed on %s Field\n\texpected\t%+v\n\treceived\t%+v\n", k, v, m.Fields[k])
					}
				}

				if m.Fields["exists"] == 1 {
					if _, ok := m.Fields["modification_time"].(int64); !ok {
						t.Error("modification_time Field is not set")
					}
				}

				foundTags = true
				break
			}
		}
		if !foundTags {
			t.Errorf("\nFailed, could not find matching tags\n\t%+v", m.Tags)
		}
	}
}

func TestGatherGlob(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "*.log",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	var testMetrics = []struct {
		Tags   map[string]string
		Fields map[string]interface{}
	}{
		{
			map[string]string{"file": dir + "log1.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		{
			map[string]string{"file": dir + "log2.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
	}

	for _, m := range acc.Metrics {
		foundTags := false
		for _, tm := range testMetrics {
			if reflect.DeepEqual(tm.Tags, m.Tags) {

				for k, v := range tm.Fields {
					if m.Fields[k] != v {
						t.Errorf("\nFailed on %s Field\n\texpected\t%+v\n\treceived\t%+v\n", k, v, m.Fields[k])
					}
				}

				if m.Fields["exists"] == 1 {
					if _, ok := m.Fields["modification_time"].(int64); !ok {
						t.Error("modification_time Field is not set")
					}
				}

				foundTags = true
				break
			}
		}
		if !foundTags {
			t.Errorf("\nFailed, could not find matching tags\n\t%+v", m.Tags)
		}
	}
}

func TestGatherSuperAsterisk(t *testing.T) {
	dir := getTestdataDir()
	fs := NewFileStat()
	fs.Md5 = true
	fs.Files = []string{
		dir + "**",
	}

	acc := testutil.Accumulator{}
	acc.GatherError(fs.Gather)

	var testMetrics = []struct {
		Tags   map[string]string
		Fields map[string]interface{}
	}{
		{
			map[string]string{"file": dir + "log1.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		{
			map[string]string{"file": dir + "log2.log"},
			map[string]interface{}{
				"size_bytes": int64(0),
				"exists":     int64(1),
				"md5_sum":    "d41d8cd98f00b204e9800998ecf8427e",
			},
		},
		{
			map[string]string{"file": dir + "test.conf"},
			map[string]interface{}{
				"size_bytes": int64(104),
				"exists":     int64(1),
				"md5_sum":    "5a7e9b77fa25e7bb411dbd17cf403c1f",
			},
		},
	}

	for _, m := range acc.Metrics {
		foundTags := false
		for _, tm := range testMetrics {
			if reflect.DeepEqual(tm.Tags, m.Tags) {

				for k, v := range tm.Fields {
					if m.Fields[k] != v {
						t.Errorf("\nFailed on %s Field\n\texpected\t%+v\n\treceived\t%+v\n", k, v, m.Fields[k])
					}
				}

				if m.Fields["exists"] == 1 {
					if _, ok := m.Fields["modification_time"].(int64); !ok {
						t.Error("modification_time Field is not set")
					}
				}

				foundTags = true
				break
			}
		}
		if !foundTags {
			t.Errorf("\nFailed, could not find matching tags\n\t%+v", m.Tags)
		}
	}
}

func TestGetMd5(t *testing.T) {
	dir := getTestdataDir()
	md5, err := getMd5(dir + "test.conf")
	assert.NoError(t, err)
	assert.Equal(t, "5a7e9b77fa25e7bb411dbd17cf403c1f", md5)

	md5, err = getMd5("/tmp/foo/bar/fooooo")
	assert.Error(t, err)
}

func getTestdataDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "filestat_test.go", "testdata/", 1)
}
