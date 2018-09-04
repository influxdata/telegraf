package gorethink

import (
	"encoding/binary"
	"golang.org/x/net/context"
)

// Write 'data' to conn
func (c *Connection) writeData(data []byte) error {
	_, err := c.Conn.Write(data[:])

	return err
}

func (c *Connection) read(buf []byte, length int) (total int, err error) {
	var n int
	for total < length {
		if n, err = c.Conn.Read(buf[total:length]); err != nil {
			break
		}
		total += n
	}

	return total, err
}

func (c *Connection) writeQuery(token int64, q []byte) error {
	pos := 0
	dataLen := 8 + 4 + len(q)
	data := make([]byte, dataLen)

	// Send the protocol version to the server as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint64(data[pos:], uint64(token))
	pos += 8

	// Send the length of the auth key to the server as a 4-byte little-endian-encoded integer
	binary.LittleEndian.PutUint32(data[pos:], uint32(len(q)))
	pos += 4

	// Send the auth key as an ASCII string
	pos += copy(data[pos:], q)

	return c.writeData(data)
}

func (c *Connection) contextFromConnectionOpts() context.Context {
	sum := c.opts.ReadTimeout + c.opts.WriteTimeout
	if c.opts.ReadTimeout == 0 || c.opts.WriteTimeout == 0 {
		return context.Background()
	}
	ctx, _ := context.WithTimeout(context.Background(), sum)
	return ctx
}
