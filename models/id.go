package models

import (
	"math/rand" // TODO: maybe switch to crypto/rand
	"sync/atomic"
)

var (
	globalPluginIDIncrement uint32 = 0
	globalInstanceID        uint32 = rand.Uint32() // set a new instance ID on every app load.
)

// nextPluginID generates a globally unique plugin ID for use referencing the plugin within the lifetime of Telegraf.
func nextPluginID() uint64 {
	num := atomic.AddUint32(&globalPluginIDIncrement, 1)
	return uint64(globalInstanceID)<<32 + uint64(num)
}
