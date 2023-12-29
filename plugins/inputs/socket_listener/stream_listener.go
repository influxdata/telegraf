package socket_listener

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/mdlayher/vsock"
)

type hasSetReadBuffer interface {
	SetReadBuffer(bytes int) error
}

type streamListener struct {
	Encoding        string
	ReadBufferSize  int
	MaxConnections  int
	ReadTimeout     config.Duration
	KeepAlivePeriod *config.Duration
	Splitter        bufio.SplitFunc
	Parser          telegraf.Parser
	Log             telegraf.Logger

	listener    net.Listener
	connections map[net.Conn]struct{}
	path        string

	wg sync.WaitGroup
	sync.Mutex
}

func (l *streamListener) setupTCP(u *url.URL, tlsCfg *tls.Config) error {
	var err error
	if tlsCfg == nil {
		l.listener, err = net.Listen(u.Scheme, u.Host)
	} else {
		l.listener, err = tls.Listen(u.Scheme, u.Host, tlsCfg)
	}
	return err
}

func (l *streamListener) setupUnix(u *url.URL, tlsCfg *tls.Config, socketMode string) error {
	err := os.Remove(u.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing socket failed: %w", err)
	}

	if tlsCfg == nil {
		l.listener, err = net.Listen(u.Scheme, u.Path)
	} else {
		l.listener, err = tls.Listen(u.Scheme, u.Path, tlsCfg)
	}
	if err != nil {
		return err
	}
	l.path = u.Path

	// Set permissions on socket
	if socketMode != "" {
		// Convert from octal in string to int
		i, err := strconv.ParseUint(socketMode, 8, 32)
		if err != nil {
			return fmt.Errorf("converting socket mode failed: %w", err)
		}

		perm := os.FileMode(uint32(i))
		if err := os.Chmod(u.Path, perm); err != nil {
			return fmt.Errorf("changing socket permissions failed: %w", err)
		}
	}
	return nil
}

func (l *streamListener) setupVsock(u *url.URL) error {
	var err error

	addrTuple := strings.SplitN(u.String(), ":", 2)

	// Check address string for containing two tokens
	if len(addrTuple) < 2 {
		return fmt.Errorf("CID and/or port number missing")
	}
	// Parse CID and port number from address string both being 32-bit
	// source: https://man7.org/linux/man-pages/man7/vsock.7.html
	cid, _ := strconv.ParseUint(addrTuple[0], 10, 32)
	if (cid >= uint64(math.Pow(2, 32))-1) && (cid <= 0) {
		return fmt.Errorf("CID %d is out of range", cid)
	}
	port, _ := strconv.ParseUint(addrTuple[1], 10, 32)
	if (port >= uint64(math.Pow(2, 32))-1) && (port <= 0) {
		return fmt.Errorf("Port number %d is out of range", port)
	}

	l.listener, err = vsock.Listen(uint32(port), nil)
	return err
}

func (l *streamListener) setupConnection(conn net.Conn) error {
	if c, ok := conn.(*tls.Conn); ok {
		conn = c.NetConn()
	}

	addr := conn.RemoteAddr().String()
	l.Lock()
	if l.MaxConnections > 0 && len(l.connections) >= l.MaxConnections {
		l.Unlock()
		// Ignore the returned error as we cannot do anything about it anyway
		_ = conn.Close()
		return fmt.Errorf("unable to accept connection from %q: too many connections", addr)
	}
	l.connections[conn] = struct{}{}
	l.Unlock()

	if l.ReadBufferSize > 0 {
		if rb, ok := conn.(hasSetReadBuffer); ok {
			if err := rb.SetReadBuffer(l.ReadBufferSize); err != nil {
				l.Log.Warnf("Setting read buffer on socket failed: %v", err)
			}
		} else {
			l.Log.Warn("Cannot set read buffer on socket of this type")
		}
	}

	// Set keep alive handlings
	if l.KeepAlivePeriod != nil {
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			l.Log.Warnf("connection not a TCP connection (%T)", conn)
		}
		if *l.KeepAlivePeriod == 0 {
			if err := tcpConn.SetKeepAlive(false); err != nil {
				l.Log.Warnf("Cannot set keep-alive: %v", err)
			}
		} else {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				l.Log.Warnf("Cannot set keep-alive: %v", err)
			}
			err := tcpConn.SetKeepAlivePeriod(time.Duration(*l.KeepAlivePeriod))
			if err != nil {
				l.Log.Warnf("Cannot set keep-alive period: %v", err)
			}
		}
	}

	return nil
}

func (l *streamListener) closeConnection(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.EPIPE) {
		l.Log.Warnf("Cannot close connection to %q: %v", addr, err)
	}
	delete(l.connections, conn)
}

func (l *streamListener) addr() net.Addr {
	return l.listener.Addr()
}

func (l *streamListener) close() error {
	if err := l.listener.Close(); err != nil {
		return err
	}

	l.Lock()
	for conn := range l.connections {
		l.closeConnection(conn)
	}
	l.Unlock()
	l.wg.Wait()

	if l.path != "" {
		err := os.Remove(l.path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			// Ignore file-not-exists errors when removing the socket
			return err
		}
	}
	return nil
}

func (l *streamListener) listen(acc telegraf.Accumulator) {
	l.connections = make(map[net.Conn]struct{})

	l.wg.Add(1)
	defer l.wg.Done()

	var wg sync.WaitGroup
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				acc.AddError(err)
			}
			break
		}

		if err := l.setupConnection(conn); err != nil {
			acc.AddError(err)
			continue
		}

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			if err := l.read(acc, c); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
					acc.AddError(err)
				}
			}
			l.Lock()
			l.closeConnection(conn)
			l.Unlock()
		}(conn)
	}
	wg.Wait()
}

func (l *streamListener) read(acc telegraf.Accumulator, conn net.Conn) error {
	decoder, err := internal.NewStreamContentDecoder(l.Encoding, conn)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}

	timeout := time.Duration(l.ReadTimeout)

	scanner := bufio.NewScanner(decoder)
	scanner.Split(l.Splitter)
	for {
		// Set the read deadline, if any, then start reading. The read
		// will accept the deadline and return if no or insufficient data
		// arrived in time. We need to set the deadline in every cycle as
		// it is an ABSOLUTE time and not a timeout.
		if timeout > 0 {
			deadline := time.Now().Add(timeout)
			if err := conn.SetReadDeadline(deadline); err != nil {
				return fmt.Errorf("setting read deadline failed: %w", err)
			}
		}
		if !scanner.Scan() {
			// Exit if no data arrived e.g. due to timeout or closed connection
			break
		}

		data := scanner.Bytes()
		metrics, err := l.Parser.Parse(data)
		if err != nil {
			acc.AddError(fmt.Errorf("parsing error: %w", err))
			l.Log.Debugf("invalid data for parser: %v", data)
			continue
		}
		for _, m := range metrics {
			acc.AddMetric(m)
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			// Ignore the timeout and silently close the connection
			l.Log.Debug(err)
			return nil
		}
		if errors.Is(err, net.ErrClosed) {
			// Ignore the connection closing of the remote side
			return nil
		}
		return err
	}
	return nil
}
