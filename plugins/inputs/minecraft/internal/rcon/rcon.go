// Package rcon implements the communication protocol for communicating
// with RCON servers. Tested and working with Valve game servers.
package rcon

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	PacketPaddingSize uint8 = 2 // Size of Packet's padding.
	PacketHeaderSize  uint8 = 8 // Size of Packet's header.
)

const (
	TerminationSequence = "\x00" // Null empty ASCII string suffix.
)

// Packet type constants.
// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Packet_Type
const (
	Exec          int32 = 2
	Auth          int32 = 3
	AuthResponse  int32 = 2
	ResponseValue int32 = 0
)

// Rcon package errors.
var (
	ErrInvalidWrite        = errors.New("Failed to write the payload correctly to remote connection.")
	ErrInvalidRead         = errors.New("Failed to read the response correctly from remote connection.")
	ErrInvalidChallenge    = errors.New("Server failed to mirror request challenge.")
	ErrUnauthorizedRequest = errors.New("Client not authorized to remote server.")
	ErrFailedAuthorization = errors.New("Failed to authorize to the remote server.")
)

type Client struct {
	Host       string   // The IP address of the remote server.
	Port       int      // The Port the remote server's listening on.
	Authorized bool     // Has the client been authorized by the server?
	Connection net.Conn // The TCP connection to the server.
}

type Header struct {
	Size      int32 // The size of the payload.
	Challenge int32 // The challenge ths server should mirror.
	Type      int32 // The type of request being sent.
}

type Packet struct {
	Header Header // Packet header.
	Body   string // Body of packet.
}

// Compile converts a packets header and body into its appropriate
// byte array payload, returning an error if the binary packages
// Write method fails to write the header bytes in their little
// endian byte order.
func (p Packet) Compile() (payload []byte, err error) {
	var size int32 = p.Header.Size
	var buffer bytes.Buffer
	var padding [PacketPaddingSize]byte

	if err = binary.Write(&buffer, binary.LittleEndian, &size); nil != err {
		return
	} else if err = binary.Write(&buffer, binary.LittleEndian, &p.Header.Challenge); nil != err {
		return
	} else if err = binary.Write(&buffer, binary.LittleEndian, &p.Header.Type); nil != err {
		return
	}

	buffer.WriteString(p.Body)
	buffer.Write(padding[:])

	return buffer.Bytes(), nil
}

// NewPacket returns a pointer to a new Packet type.
func NewPacket(challenge, typ int32, body string) (packet *Packet) {
	size := int32(len([]byte(body)) + int(PacketHeaderSize+PacketPaddingSize))
	return &Packet{Header{size, challenge, typ}, body}
}

// Authorize calls Send with the appropriate command type and the provided
// password.  The response packet is returned if authorization is successful
// or a potential error.
func (c *Client) Authorize(password string) (response *Packet, err error) {
	if response, err = c.Send(Auth, password); nil == err {
		if response.Header.Type == AuthResponse {
			c.Authorized = true
		} else {
			err = ErrFailedAuthorization
			response = nil
			return
		}
	}

	return
}

// Execute calls Send with the appropriate command type and the provided
// command.  The response packet is returned if the command executed successfully
// or a potential error.
func (c *Client) Execute(command string) (response *Packet, err error) {
	return c.Send(Exec, command)
}

// Sends accepts the commands type and its string to execute to the clients server,
// creating a packet with a random challenge id for the server to mirror,
// and compiling its payload bytes in the appropriate order. The response is
// decompiled from its bytes into a Packet type for return. An error is returned
// if send fails.
func (c *Client) Send(typ int32, command string) (response *Packet, err error) {
	if typ != Auth && !c.Authorized {
		err = ErrUnauthorizedRequest
		return
	}

	// Create a random challenge for the server to mirror in its response.
	var challenge int32
	binary.Read(rand.Reader, binary.LittleEndian, &challenge)

	// Create the packet from the challenge, typ and command
	// and compile it to its byte payload
	packet := NewPacket(challenge, typ, command)
	payload, err := packet.Compile()

	var n int

	if nil != err {
		return
	} else if n, err = c.Connection.Write(payload); nil != err {
		return
	} else if n != len(payload) {
		err = ErrInvalidWrite
		return
	}

	var header Header

	if err = binary.Read(c.Connection, binary.LittleEndian, &header.Size); nil != err {
		return
	} else if err = binary.Read(c.Connection, binary.LittleEndian, &header.Challenge); nil != err {
		return
	} else if err = binary.Read(c.Connection, binary.LittleEndian, &header.Type); nil != err {
		return
	}

	if packet.Header.Type == Auth && header.Type == ResponseValue {
		// Discard, empty SERVERDATA_RESPONSE_VALUE from authorization.
		c.Connection.Read(make([]byte, header.Size-int32(PacketHeaderSize)))

		// Reread the packet header.
		if err = binary.Read(c.Connection, binary.LittleEndian, &header.Size); nil != err {
			return
		} else if err = binary.Read(c.Connection, binary.LittleEndian, &header.Challenge); nil != err {
			return
		} else if err = binary.Read(c.Connection, binary.LittleEndian, &header.Type); nil != err {
			return
		}
	}

	if header.Challenge != packet.Header.Challenge {
		err = ErrInvalidChallenge
		return
	}

	body := make([]byte, header.Size-int32(PacketHeaderSize))
	n, err = c.Connection.Read(body)

	for n < len(body) {
		var nBytes int
		nBytes, err = c.Connection.Read(body[n:])
		if err != nil {
			return
		}
		n += nBytes
	}

	if nil != err {
		return
	} else if n != len(body) {
		err = ErrInvalidRead
		return
	}

	response = new(Packet)
	response.Header = header
	response.Body = strings.TrimRight(string(body), TerminationSequence)

	return
}

// NewClient creates a new Client type, creating the connection
// to the server specified by the host and port arguments. If
// the connection fails, an error is returned.
func NewClient(host string, port int) (client *Client, err error) {
	client = new(Client)
	client.Host = host
	client.Port = port
	client.Connection, err = net.Dial("tcp", fmt.Sprintf("%v:%v", client.Host, client.Port))
	return
}
