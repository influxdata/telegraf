//go:build arm64be
// +build arm64be

package binary

import "encoding/binary"

var hostEndianess = binary.BigEndian
