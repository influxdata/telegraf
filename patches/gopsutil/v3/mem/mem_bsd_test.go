//go:build freebsd || openbsd
// +build freebsd openbsd

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const validFreeBSD = `Device:       1kB-blocks      Used:
/dev/gpt/swapfs    1048576          1234
/dev/md0         1048576          666
`

const validOpenBSD = `Device       1K-blocks      Used	Avail	Capacity	Priority
/dev/wd0b    655025          1234	653791	1%	0
`

const invalid = `Device:       512-blocks      Used:
/dev/gpt/swapfs    1048576          1234
/dev/md0         1048576          666
`

func TestParseSwapctlOutput_FreeBSD(t *testing.T) {
	assert := assert.New(t)
	stats, err := parseSwapctlOutput(validFreeBSD)
	assert.NoError(err)

	assert.Equal(*stats[0], SwapDevice{
		Name:      "/dev/gpt/swapfs",
		UsedBytes: 1263616,
		FreeBytes: 1072478208,
	})

	assert.Equal(*stats[1], SwapDevice{
		Name:      "/dev/md0",
		UsedBytes: 681984,
		FreeBytes: 1073059840,
	})
}

func TestParseSwapctlOutput_OpenBSD(t *testing.T) {
	assert := assert.New(t)
	stats, err := parseSwapctlOutput(validOpenBSD)
	assert.NoError(err)

	assert.Equal(*stats[0], SwapDevice{
		Name:      "/dev/wd0b",
		UsedBytes: 1234 * 1024,
		FreeBytes: 653791 * 1024,
	})
}

func TestParseSwapctlOutput_Invalid(t *testing.T) {
	_, err := parseSwapctlOutput(invalid)
	assert.Error(t, err)
}

func TestParseSwapctlOutput_Empty(t *testing.T) {
	_, err := parseSwapctlOutput("")
	assert.Error(t, err)
}
