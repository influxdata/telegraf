/*
Package network implements collectd's binary network protocol.
*/
package network // import "collectd.org/network"

// Well-known addresses and port.
const (
	DefaultIPv4Address = "239.192.74.66"
	DefaultIPv6Address = "ff18::efc0:4a42"
	DefaultService     = "25826"
)

// Default size of "Buffer". This is based on the maximum bytes that fit into
// an Ethernet frame without fragmentation:
//   <Ethernet frame> - (<IPv6 header> + <UDP header>) = 1500 - (40 + 8) = 1452
const DefaultBufferSize = 1452

// Numeric data source type identifiers.
const (
	dsTypeCounter = 0
	dsTypeGauge   = 1
	dsTypeDerive  = 2
)

// IDs of the various "parts", i.e. subcomponents of a packet.
const (
	typeHost           = 0x0000
	typeTime           = 0x0001
	typeTimeHR         = 0x0008
	typePlugin         = 0x0002
	typePluginInstance = 0x0003
	typeType           = 0x0004
	typeTypeInstance   = 0x0005
	typeValues         = 0x0006
	typeInterval       = 0x0007
	typeIntervalHR     = 0x0009
	typeSignSHA256     = 0x0200
	typeEncryptAES256  = 0x0210
)

// SecurityLevel determines whether data is signed, encrypted or used without
// any protection.
type SecurityLevel int

// Predefined security levels. "None" is used for plain text.
const (
	None SecurityLevel = iota
	Sign
	Encrypt
)
