package ts3

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultPort is the default TeamSpeak 3 ServerQuery port.
	DefaultPort = 10011

	// MaxParseTokenSize is the maximum buffer size used to parse the
	// server responses.
	// It's relatively large to enable us to deal with the typical responses
	// to commands such as serversnapshotcreate.
	MaxParseTokenSize = 10 << 20

	// connectHeader is the header used as the prefix to responses.
	connectHeader = "TS3"

	// startBufSize is the initial size of allocation for the parse buffer.
	startBufSize = 4096
)

var (
	respTrailerRe = regexp.MustCompile(`^error id=(\d+) msg=([^ ]+)(.*)`)

	// DefaultTimeout is the default read / write / dial timeout for Clients.
	DefaultTimeout = time.Second * 10
)

// Client is a TeamSpeak 3 ServerQuery client.
type Client struct {
	conn       net.Conn
	timeout    time.Duration
	scanner    *bufio.Scanner
	buf        []byte
	maxBufSize int

	Server *ServerMethods
}

// Timeout sets read / write / dial timeout for a TeamSpeak 3 Client.
func Timeout(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout
		return nil
	}
}

// Buffer sets the initial buffer used to parse responses from
// the server and the maximum size of buffer that may be allocated.
// The maximum parsable token size is the larger of max and cap(buf).
// If max <= cap(buf), scanning will use this buffer only and do no
// allocation.
//
// By default, parsing uses an internal buffer and sets the maximum
// token size to MaxParseTokenSize.
func Buffer(buf []byte, max int) func(*Client) error {
	return func(c *Client) error {
		c.buf = buf
		c.maxBufSize = max
		return nil
	}
}

// NewClient returns a new TeamSpeak 3 client connected to addr.
func NewClient(addr string, options ...func(c *Client) error) (*Client, error) {
	if !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%v:%v", addr, DefaultPort)
	}

	c := &Client{
		timeout:    DefaultTimeout,
		buf:        make([]byte, startBufSize),
		maxBufSize: MaxParseTokenSize,
	}
	for _, f := range options {
		if f == nil {
			return nil, ErrNilOption
		}
		if err := f(c); err != nil {
			return nil, err
		}
	}

	// Wire up command groups
	c.Server = &ServerMethods{Client: c}

	var err error
	if c.conn, err = net.DialTimeout("tcp", addr, c.timeout); err != nil {
		return nil, err
	}

	c.scanner = bufio.NewScanner(bufio.NewReader(c.conn))
	c.scanner.Buffer(c.buf, c.maxBufSize)
	c.scanner.Split(ScanLines)

	if err := c.setDeadline(); err != nil {
		return nil, err
	}

	// Reader the connection header
	if !c.scanner.Scan() {
		return nil, c.scanErr()
	}

	if l := c.scanner.Text(); l != connectHeader {
		return nil, fmt.Errorf("invalid connection header %q", l)
	}

	// Slurp the banner
	if !c.scanner.Scan() {
		return nil, c.scanErr()
	}

	return c, nil
}

// setDeadline updates the deadline on the connection based on the clients configured timeout.
func (c *Client) setDeadline() error {
	return c.conn.SetDeadline(time.Now().Add(c.timeout))
}

// Exec executes cmd on the server and returns the response.
func (c *Client) Exec(cmd string) ([]string, error) {
	return c.ExecCmd(NewCmd(cmd))
}

// ExecCmd executes cmd on the server and returns the response.
func (c *Client) ExecCmd(cmd *Cmd) ([]string, error) {
	if err := c.setDeadline(); err != nil {
		return nil, err
	}

	if _, err := c.conn.Write([]byte(cmd.String())); err != nil {
		return nil, err
	}

	if err := c.setDeadline(); err != nil {
		return nil, err
	}

	lines := make([]string, 0, 10)
	for c.scanner.Scan() {
		l := c.scanner.Text()
		if l == "error id=0 msg=ok" {
			if cmd.response != nil {
				if err := DecodeResponse(lines, cmd.response); err != nil {
					return nil, err
				}
			}
			return lines, nil
		} else if matches := respTrailerRe.FindStringSubmatch(l); len(matches) == 4 {
			return nil, NewError(matches)
		} else {
			lines = append(lines, l)
		}
	}

	return nil, c.scanErr()
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	_, err := c.Exec("quit")
	err2 := c.conn.Close()
	if err != nil {
		return err
	}

	return err2
}

// scanError returns the error from the scanner if non-nil,
// io.ErrUnexpectedEOF otherwise.
func (c *Client) scanErr() error {
	if err := c.scanner.Err(); err != nil {
		return err
	}
	return io.ErrUnexpectedEOF
}
