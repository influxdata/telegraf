//go:build arm64be
// +build arm64be

package internal

import "encoding/binary"

var HostEndianess = binary.BigEndian
