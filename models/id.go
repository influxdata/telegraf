package models

import (
	"math/rand"
	"sync/atomic"
)

var (
	globalPluginIDIncrement uint32
	globalInstanceID        = rand.Uint32() // set a new instance ID on every app load.
)

// NextPluginID generates a globally unique plugin ID for use referencing the plugin within the lifetime of Telegraf.
func NextPluginID() uint64 {
	num := atomic.AddUint32(&globalPluginIDIncrement, 1)
	return uint64(globalInstanceID)<<32 + uint64(num)
}
