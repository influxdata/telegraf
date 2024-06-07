//go:build armbe || arm64be || mips || mips64 || mips64p32 || ppc || ppc64 || s390 || s390x || sparc || sparc64

package internal

import "encoding/binary"

var HostEndianness = binary.BigEndian
