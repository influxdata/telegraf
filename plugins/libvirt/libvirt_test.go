package libvirt

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestLibvirt(t *testing.T) {
	var acc testutil.Accumulator

	l := &Libvirt{Uri: "test:///default"}
	l.Gather(&acc)

	assert.True(t, acc.HasUIntValue("cpu_time"))
	assert.True(t, acc.HasUIntValue("memory"))
	assert.True(t, acc.HasUIntValue("max_mem"))

	expectedTags := map[string]string{"domain": "test"}
	expectedNumberOfCpus := uint16(2)
	assert.NoError(t, acc.ValidateTaggedValue("nr_virt_cpu", expectedNumberOfCpus, expectedTags))
}
