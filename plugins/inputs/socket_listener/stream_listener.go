package socket_listener

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
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
	connections map[string]net.Conn
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

func (l *streamListener) setupConnection(conn net.Conn) error {
	if l.ReadBufferSize > 0 {
		if rb, ok := conn.(hasSetReadBuffer); ok {
			if err := rb.SetReadBuffer(l.ReadBufferSize); err != nil {
				l.Log.Warnf("Setting read buffer on socket failed: %v", err)
			}
		} else {
			l.Log.Warn("Cannot set read buffer on socket of this type")
		}
	}

	addr := conn.RemoteAddr().String()
	if l.MaxConnections > 0 && len(l.connections) >= l.MaxConnections {
		// Ignore the returned error as we cannot do anything about it anyway
		_ = conn.Close()
		l.Log.Infof("unable to accept connection from %q: too many connections", addr)
		return nil
	}

	// Set keep alive handlings
	if l.KeepAlivePeriod != nil {
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return fmt.Errorf("connection not a TCP connection (%T)", conn)
		}
		if *l.KeepAlivePeriod == 0 {
			if err := tcpConn.SetKeepAlive(false); err != nil {
				return fmt.Errorf("cannot set keep-alive: %w", err)
			}
		} else {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				return fmt.Errorf("cannot set keep-alive: %w", err)
			}
			err := tcpConn.SetKeepAlivePeriod(time.Duration(*l.KeepAlivePeriod))
			if err != nil {
				return fmt.Errorf("cannot set keep-alive period: %w", err)
			}
		}
	}

	// Store the connection mapped to its address
	l.Lock()
	defer l.Unlock()
	l.connections[addr] = conn

	return nil
}

func (l *streamListener) closeConnection(conn net.Conn) {
	l.Lock()
	defer l.Unlock()
	addr := conn.RemoteAddr().String()
	if err := conn.Close(); err != nil {
		l.Log.Errorf("Cannot close connection to %q: %v", addr, err)
	}
	delete(l.connections, addr)
}

func (l *streamListener) addr() net.Addr {
	return l.listener.Addr()
}

func (l *streamListener) close() error {
	if err := l.listener.Close(); err != nil {
		return err
	}

	for _, conn := range l.connections {
		l.closeConnection(conn)
	}
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
	l.connections = make(map[string]net.Conn)

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
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.read(acc, conn); err != nil {
				acc.AddError(err)
			}
		}()
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
