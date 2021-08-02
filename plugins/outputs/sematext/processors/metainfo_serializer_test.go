package processors

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializeMetainfo(t *testing.T) {
	mInfo := &MetricMetainfo{
		name:        "cpu.user",
		namespace:   "os",
		label:       "os.cpu.user",
		semType:     Gauge,
		numericType: Long,
		token:       "token",
		host:        "host001",
	}

	body := string(serializeMetainfo(mInfo))

	assert.Equal(t, "os,label=os.cpu.user,numericType=long,os.host=host001,token=token,type=gauge cpu.user", body)
}
