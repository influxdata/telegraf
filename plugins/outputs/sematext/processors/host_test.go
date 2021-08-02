package processors

import (
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAdjustHostname(t *testing.T) {
	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	adjustHostname(m, "abc")

	newVal, set := m.GetTag(sematextHostTag)

	assert.Equal(t, true, set)
	assert.Equal(t, "abc", newVal)
}

func TestLoadHostname(t *testing.T) {
	assert.Equal(t, "somehost001", onlyHost(loadHostname("./testdata/resolved-hostname")))
	assert.Equal(t, "", onlyHost(loadHostname("./testdata/doesnt-exist")))
	assert.Equal(t, "", onlyHost(loadHostname("/baddir")))
	assert.Equal(t, "somehost001", onlyHost(loadHostname("./testdata/resolved-hostname-multiline")))
	assert.Equal(t, "somehost001", onlyHost(loadHostname("./testdata/resolved-hostname-multiline2")))
}

func onlyHost(host string, _ error) string {
	return host
}

func TestHostProcess(t *testing.T) {
	h := &Host{
		hostname: "abc",
	}

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	err := h.Process(m)
	assert.Nil(t, err)

	newVal, set := m.GetTag(sematextHostTag)

	assert.Equal(t, true, set)
	assert.Equal(t, "abc", newVal)
}
