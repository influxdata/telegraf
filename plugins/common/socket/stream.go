package socket

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alitto/pond"
	"github.com/mdlayher/vsock"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

type hasSetReadBuffer interface {
	SetReadBuffer(bytes int) error
}

type streamListener struct {
	AllowedSources  []net.IP
	Encoding        string
	ReadBufferSize  int
	MaxConnections  uint64
	ReadTimeout     config.Duration
	KeepAlivePeriod *config.Duration
	Splitter        bufio.SplitFunc
	Log             telegraf.Logger

	listener    net.Listener
	connections uint64
	path        string
	cancel      context.CancelFunc
	parsePool   *pond.WorkerPool

	wg sync.WaitGroup
	sync.Mutex
}

func newStreamListener(conf Config, splitter bufio.SplitFunc, log telegraf.Logger) *streamListener {
	return &streamListener{
		AllowedSources:  conf.AllowedSources,
		ReadBufferSize:  int(conf.ReadBufferSize),
		ReadTimeout:     conf.ReadTimeout,
		KeepAlivePeriod: conf.KeepAlivePeriod,
		MaxConnections:  conf.MaxConnections,
		Encoding:        conf.ContentEncoding,
		Splitter:        splitter,
		Log:             log,

		parsePool: pond.New(
			conf.MaxParallelParsers,
			0,
			pond.MinWorkers(conf.MaxParallelParsers/2+1)),
	}
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
	l.path = filepath.FromSlash(u.Path)
	if runtime.GOOS == "windows" && strings.Contains(l.path, ":") {
		l.path = strings.TrimPrefix(l.path, `\`)
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing socket failed: %w", err)
	}

	var err error
	if tlsCfg == nil {
		l.listener, err = net.Listen(u.Scheme, l.path)
	} else {
		l.listener, err = tls.Listen(u.Scheme, l.path, tlsCfg)
	}
	if err != nil {
		return err
	}

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
		return errors.New("port and/or CID number missing")
	}
	// Parse CID and port number from address string both being 32-bit
	// source: https://man7.org/linux/man-pages/man7/vsock.7.html
	cid, err := strconv.ParseUint(addrTuple[0], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse CID %s: %w", addrTuple[0], err)
	}
	if (cid >= uint64(math.Pow(2, 32))-1) && (cid <= 0) {
		return fmt.Errorf("value of CID %d is out of range", cid)
	}
	port, err := strconv.ParseUint(addrTuple[1], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse port number %s: %w", addrTuple[1], err)
	}
	if (port >= uint64(math.Pow(2, 32))-1) && (port <= 0) {
		return fmt.Errorf("port number %d is out of range", port)
	}

	l.listener, err = vsock.Listen(uint32(port), nil)
	return err
}

func (l *streamListener) setupConnection(conn net.Conn) error {
	addr := conn.RemoteAddr().String()
	l.Lock()
	if l.MaxConnections > 0 && l.connections >= l.MaxConnections {
		l.Unlock()
		// Ignore the returned error as we cannot do anything about it anyway
		_ = conn.Close()
		return fmt.Errorf("unable to accept connection from %q: too many connections", addr)
	}
	l.connections++
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
		if c, ok := conn.(*tls.Conn); ok {
			conn = c.NetConn()
		}
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
	// Fallback to enforce blocked reads on connections to end immediately
	//nolint:errcheck // Ignore errors as this is a fallback only
	conn.SetReadDeadline(time.Now())

	addr := conn.RemoteAddr().String()
	if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.EPIPE) {
		l.Log.Warnf("Cannot close connection to %q: %v", addr, err)
	} else {
		l.Lock()
		l.connections--
		l.Unlock()
	}
}

func (l *streamListener) address() net.Addr {
	return l.listener.Addr()
}

func (l *streamListener) close() error {
	if l.listener != nil {
		// Continue even if we cannot close the listener in order to at least
		// close all active connections
		if err := l.listener.Close(); err != nil {
			l.Log.Errorf("Cannot close listener: %v", err)
		}
	}

	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
	l.wg.Wait()

	if l.path != "" {
		fn := filepath.FromSlash(l.path)
		if runtime.GOOS == "windows" && strings.Contains(fn, ":") {
			fn = strings.TrimPrefix(fn, `\`)
		}
		// Ignore file-not-exists errors when removing the socket
		if err := os.Remove(fn); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	l.parsePool.StopAndWait()

	return nil
}

func (l *streamListener) listenData(onData CallbackData, onError CallbackError) {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		for {
			conn, err := l.listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) && onError != nil {
					onError(err)
				}
				break
			}

			if allowed, err := isSourceAllowed(l.AllowedSources, conn.RemoteAddr()); err != nil {
				if onError != nil {
					onError(err)
				}
				if err := conn.Close(); err != nil {
					onError(fmt.Errorf("closing connection from %q failed: %w", conn.RemoteAddr(), err))
				}
				continue
			} else if !allowed {
				if err = conn.Close(); err != nil {
					onError(fmt.Errorf("closing connection from %q failed: %w", conn.RemoteAddr(), err))
				}
				continue
			}

			if err := l.setupConnection(conn); err != nil && onError != nil {
				onError(err)
				continue
			}

			l.wg.Add(1)
			go l.handleReaderConn(ctx, conn, onData, onError)
		}
	}()
}

func (l *streamListener) handleReaderConn(ctx context.Context, conn net.Conn, onData CallbackData, onError CallbackError) {
	defer l.wg.Done()

	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer l.closeConnection(conn)
	stopFunc := context.AfterFunc(localCtx, func() { l.closeConnection(conn) })
	defer stopFunc()

	reader := l.read
	if l.Splitter == nil {
		reader = l.readAll
	}
	if err := reader(conn, onData); err != nil {
		if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
			if onError != nil {
				onError(err)
			}
		}
	}
}

func (l *streamListener) listenConnection(onConnection CallbackConnection, onError CallbackError) {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		for {
			conn, err := l.listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) && onError != nil {
					onError(err)
				}
				break
			}

			if allowed, err := isSourceAllowed(l.AllowedSources, conn.RemoteAddr()); err != nil {
				if onError != nil {
					onError(err)
				}
				if err = conn.Close(); err != nil {
					onError(fmt.Errorf("closing connection from %q failed: %w", conn.RemoteAddr(), err))
				}
				continue
			} else if !allowed {
				if err = conn.Close(); err != nil {
					onError(fmt.Errorf("closing connection from %q failed: %w", conn.RemoteAddr(), err))
				}
				l.Log.Debugf("Received message from blocked IP: %s", conn.RemoteAddr())
				continue
			}

			if err := l.setupConnection(conn); err != nil && onError != nil {
				onError(err)
				continue
			}

			l.wg.Add(1)
			go func(c net.Conn) {
				if err := l.handleConnection(ctx, c, onConnection); err != nil {
					if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) {
						if onError != nil {
							onError(err)
						}
					}
				}
			}(conn)
		}
	}()
}

func (l *streamListener) read(conn net.Conn, onData CallbackData) error {
	decoder, err := internal.NewStreamContentDecoder(l.Encoding, conn)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}

	timeout := time.Duration(l.ReadTimeout)

	scanner := bufio.NewScanner(decoder)
	if l.ReadBufferSize > bufio.MaxScanTokenSize {
		scanner.Buffer(make([]byte, l.ReadBufferSize), l.ReadBufferSize)
	}
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

		receiveTime := time.Now()
		src := conn.RemoteAddr()
		if l.path != "" {
			src = &net.UnixAddr{Name: l.path, Net: "unix"}
		}

		data := scanner.Bytes()
		d := make([]byte, len(data))
		copy(d, data)
		l.parsePool.Submit(func() {
			onData(src, d, receiveTime)
		})
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

func (l *streamListener) readAll(conn net.Conn, onData CallbackData) error {
	src := conn.RemoteAddr()
	if l.path != "" {
		src = &net.UnixAddr{Name: l.path, Net: "unix"}
	}

	decoder, err := internal.NewStreamContentDecoder(l.Encoding, conn)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}

	timeout := time.Duration(l.ReadTimeout)
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
	buf, err := io.ReadAll(decoder)
	if err != nil {
		return fmt.Errorf("read on %s failed: %w", src, err)
	}

	receiveTime := time.Now()
	l.parsePool.Submit(func() {
		onData(src, buf, receiveTime)
	})

	return nil
}

func (l *streamListener) handleConnection(ctx context.Context, conn net.Conn, onConnection CallbackConnection) error {
	defer l.wg.Done()

	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer l.closeConnection(conn)
	stopFunc := context.AfterFunc(localCtx, func() { l.closeConnection(conn) })
	defer stopFunc()

	// Prepare the data decoder for the connection
	decoder, err := internal.NewStreamContentDecoder(l.Encoding, conn)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}

	// Get the remote address
	src := conn.RemoteAddr()
	if l.path != "" {
		src = &net.UnixAddr{Name: l.path, Net: "unix"}
	}

	// Create a pipe and feed it to the callback
	reader, writer := io.Pipe()
	defer writer.Close()
	go onConnection(src, reader)

	timeout := time.Duration(l.ReadTimeout)
	buf := make([]byte, 4096) // 4kb
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

		// Copy the data
		n, err := decoder.Read(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				if !errors.Is(err, os.ErrDeadlineExceeded) && errors.Is(err, net.ErrClosed) {
					writer.CloseWithError(err)
				}
			}
			return nil
		}
		if _, err := writer.Write(buf[:n]); err != nil {
			return err
		}
	}
}
