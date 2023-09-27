//go:build 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mipsle || mips64le || mips64p32le || ppc64le || riscv || riscv64 || wasm

package internal

import "encoding/binary"

var HostEndianness = binary.LittleEndian
