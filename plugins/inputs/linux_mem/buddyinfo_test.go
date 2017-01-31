// +build linux

package linux_mem

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var buddyinfoSample = `Node 0, zone      DMA      1      1      1      0      3      0      1      0      1      1      3 
Node 0, zone    DMA32    611  19989  10727   1841    246    103     12      0      0      0      1 
Node 0, zone   Normal   2735  87214  50555   6889    563     34      1      2      3      1      0 
`

func TestBuddyinfoGather(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tf.Name())
	defer func(p string) { buddyinfoPath = p }(buddyinfoPath)
	buddyinfoPath = tf.Name()

	_, err = tf.Write([]byte(buddyinfoSample))
	require.NoError(t, err)

	bi := Buddyinfo{}

	var acc testutil.Accumulator
	require.NoError(t, bi.Gather(&acc))

	buddyinfo0DMA := map[string]interface{}{
		"4k":    uint64(1),
		"8k":    uint64(1),
		"16k":   uint64(1),
		"32k":   uint64(0),
		"64k":   uint64(3),
		"128k":  uint64(0),
		"256k":  uint64(1),
		"512k":  uint64(0),
		"1024k": uint64(1),
		"2048k": uint64(1),
		"4096k": uint64(3),
	}
	buddyinfo0DMATags := map[string]string{
		"node": "0",
		"zone": "DMA",
	}
	acc.AssertContainsTaggedFields(t, "buddyinfo", buddyinfo0DMA, buddyinfo0DMATags)

	buddyinfo0DMA32 := map[string]interface{}{
		"4k":    uint64(611),
		"8k":    uint64(19989),
		"16k":   uint64(10727),
		"32k":   uint64(1841),
		"64k":   uint64(246),
		"128k":  uint64(103),
		"256k":  uint64(12),
		"512k":  uint64(0),
		"1024k": uint64(0),
		"2048k": uint64(0),
		"4096k": uint64(1),
	}
	buddyinfo0DMA32Tags := map[string]string{
		"node": "0",
		"zone": "DMA32",
	}
	acc.AssertContainsTaggedFields(t, "buddyinfo", buddyinfo0DMA32, buddyinfo0DMA32Tags)

	buddyinfo0Normal := map[string]interface{}{
		"4k":    uint64(2735),
		"8k":    uint64(87214),
		"16k":   uint64(50555),
		"32k":   uint64(6889),
		"64k":   uint64(563),
		"128k":  uint64(34),
		"256k":  uint64(1),
		"512k":  uint64(2),
		"1024k": uint64(3),
		"2048k": uint64(1),
		"4096k": uint64(0),
	}
	buddyinfo0NormalTags := map[string]string{
		"node": "0",
		"zone": "Normal",
	}
	acc.AssertContainsTaggedFields(t, "buddyinfo", buddyinfo0Normal, buddyinfo0NormalTags)
}
