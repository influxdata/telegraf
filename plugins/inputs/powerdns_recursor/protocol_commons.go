package powerdns_recursor

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"unsafe"
)

func parseResponse(metrics string) map[string]interface{} {
	values := make(map[string]interface{})

	s := strings.Split(metrics, "\n")

	for _, metric := range s[:len(s)-1] {
		m := strings.Split(metric, "\t")

		if len(m) < 2 {
			continue
		}

		i, err := strconv.ParseInt(m[1], 10, 64)

		if err != nil {
			continue
		}

		values[m[0]] = i
	}

	return values
}

// This below is generally unsafe but necessary in this case
// since the powerdns protocol encoding is host dependent.
// The C implementation uses size_t as the size type for the
// command length. The size and endianness of size_t change
// depending on the platform the program is being run on.
// At the time of writing, the Go type `uint` has the same
// behavior, where its size and endianness are platform
// dependent. This means that we can do an unsafe cast to
// grab the data representing the length, and know that in
// most cases it'll be what we need. In all other cases,
// we still handle all error cases gracefully, so the plugin
// will just fail to gather data.

const UIntSizeInBytes = strconv.IntSize / 8

func getEndianness() binary.ByteOrder {
	buf := make([]byte, 2)
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0x0001)

	if buf[0] == 1 {
		return binary.LittleEndian
	}

	return binary.BigEndian
}

func writeNativeUIntToConn(conn net.Conn, value uint) (int, error) {
	intData := make([]byte, UIntSizeInBytes)

	if UIntSizeInBytes == 4 {
		getEndianness().PutUint32(intData, uint32(value))
	} else if UIntSizeInBytes == 8 {
		getEndianness().PutUint64(intData, uint64(value))
	} else {
		return 0, fmt.Errorf("unsupported system configuration")
	}

	return conn.Write(intData)
}

func readNativeUIntFromConn(conn net.Conn) (uint, error) {
	intData := make([]byte, UIntSizeInBytes)

	n, err := conn.Read(intData)

	if err != nil {
		return 0, err
	}

	if n != UIntSizeInBytes {
		return 0, fmt.Errorf("did not read enough data for native uint: read '%v' bytes, expected '%v'", n, UIntSizeInBytes)
	}

	if UIntSizeInBytes == 4 {
		return uint(getEndianness().Uint32(intData)), nil
	} else if UIntSizeInBytes == 8 {
		return uint(getEndianness().Uint64(intData)), nil
	} else {
		return 0, fmt.Errorf("unsupported system configuration")
	}
}
