package socket

import (
	"errors"
	"fmt"
	"github.com/alitto/pond"
	"github.com/influxdata/telegraf/config"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type packetListener struct {
	Encoding             string
	MaxDecompressionSize int64
	SocketMode           string
	ReadBufferSize       int
	Log                  telegraf.Logger

	conn      net.PacketConn
	decoders  []internal.ContentDecoder
	path      string
	wg        sync.WaitGroup
	parsePool *pond.WorkerPool
}

func newPacketListener(encoding string, maxDecompressionSize config.Size, maxWorkers int) *packetListener {
	return &packetListener{
		Encoding:             encoding,
		MaxDecompressionSize: int64(maxDecompressionSize),
		parsePool:            pond.New(maxWorkers, 0, pond.MinWorkers(maxWorkers/2+1)),
	}
}

func (l *packetListener) listenData(onData CallbackData, onError CallbackError) {
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
		for {
			n, src, err := l.conn.ReadFrom(buf)
			receiveTime := time.Now()
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					if onError != nil {
						onError(err)
					}
				}
				break
			}

			d := make([]byte, n)
			copy(d, buf[:n])
			decoderIdx := int(l.parsePool.SubmittedTasks()) % len(l.decoders)
			decoder := l.decoders[decoderIdx]
			l.parsePool.Submit(func() {
				body, err := decoder.Decode(d)
				if err != nil && onError != nil {
					onError(fmt.Errorf("unable to decode incoming packet: %w", err))
				}

				if l.path != "" {
					src = &net.UnixAddr{Name: l.path, Net: "unixgram"}
				}

				onData(src, body, receiveTime)
			})
		}
	}()
}

func (l *packetListener) listenConnection(onConnection CallbackConnection, onError CallbackError) {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		defer l.conn.Close()

		buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
		for {
			// Wait for packets and read them
			n, src, err := l.conn.ReadFrom(buf)
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					if onError != nil {
						onError(err)
					}
				}
				break
			}

			d := make([]byte, n)
			copy(d, buf[:n])
			decoderIdx := int(l.parsePool.SubmittedTasks()) % len(l.decoders)
			decoder := l.decoders[decoderIdx]
			l.parsePool.Submit(func() {
				// Decode the contents depending on the given encoding
				body, err := decoder.Decode(d[:n])
				if err != nil && onError != nil {
					onError(fmt.Errorf("unable to decode incoming packet: %w", err))
				}

				// Workaround to provide remote endpoints for Unix-type sockets
				if l.path != "" {
					src = &net.UnixAddr{Name: l.path, Net: "unixgram"}
				}

				// Create a pipe and notify the caller via Callback that new data is
				// available. Afterwards write the data. Please note: Write() will
				// blocks until all data is consumed!
				reader, writer := io.Pipe()
				go onConnection(src, reader)
				if _, err := writer.Write(body); err != nil && onError != nil {
					onError(err)
				}
				writer.Close()
			})
		}
	}()
}

func (l *packetListener) setupUnixgram(u *url.URL, socketMode string) error {
	l.path = filepath.FromSlash(u.Path)
	if runtime.GOOS == "windows" && strings.Contains(l.path, ":") {
		l.path = strings.TrimPrefix(l.path, `\`)
	}
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing socket failed: %w", err)
	}

	conn, err := net.ListenPacket(u.Scheme, l.path)
	if err != nil {
		return fmt.Errorf("listening (unixgram) failed: %w", err)
	}
	l.conn = conn

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

	err = l.setupDecoder()

	return err
}

func (l *packetListener) setupUDP(u *url.URL, ifname string, bufferSize int) error {
	var conn *net.UDPConn

	addr, err := net.ResolveUDPAddr(u.Scheme, u.Host)
	if err != nil {
		return fmt.Errorf("resolving UDP address failed: %w", err)
	}
	if addr.IP.IsMulticast() {
		var iface *net.Interface
		if ifname != "" {
			var err error
			iface, err = net.InterfaceByName(ifname)
			if err != nil {
				return fmt.Errorf("resolving address of %q failed: %w", ifname, err)
			}
		}
		conn, err = net.ListenMulticastUDP(u.Scheme, iface, addr)
		if err != nil {
			return fmt.Errorf("listening (udp multicast) failed: %w", err)
		}
	} else {
		conn, err = net.ListenUDP(u.Scheme, addr)
		if err != nil {
			return fmt.Errorf("listening (udp) failed: %w", err)
		}
	}

	if bufferSize > 0 {
		if err := conn.SetReadBuffer(bufferSize); err != nil {
			l.Log.Warnf("Setting read buffer on %s socket failed: %v", u.Scheme, err)
		}
	}

	l.conn = conn
	err = l.setupDecoder()

	return err
}

func (l *packetListener) setupIP(u *url.URL) error {
	conn, err := net.ListenPacket(u.Scheme, u.Host)
	if err != nil {
		return fmt.Errorf("listening (ip) failed: %w", err)
	}
	l.conn = conn
	err = l.setupDecoder()

	return err
}

func (l *packetListener) setupDecoder() error {
	// Create a decoder for the given encoding
	var options []internal.DecodingOption
	if l.MaxDecompressionSize > 0 {
		options = append(options, internal.WithMaxDecompressionSize(l.MaxDecompressionSize))
	}

	l.decoders = make([]internal.ContentDecoder, 0, l.parsePool.MaxWorkers())
	for range l.parsePool.MaxWorkers() {
		decoder, err := internal.NewContentDecoder(l.Encoding, options...)
		if err != nil {
			return fmt.Errorf("creating decoder failed: %w", err)
		}

		l.decoders = append(l.decoders, decoder)
	}

	return nil
}

func (l *packetListener) address() net.Addr {
	return l.conn.LocalAddr()
}

func (l *packetListener) close() error {
	if err := l.conn.Close(); err != nil {
		return err
	}
	l.wg.Wait()

	if l.path != "" {
		fn := filepath.FromSlash(l.path)
		if runtime.GOOS == "windows" && strings.Contains(fn, ":") {
			fn = strings.TrimPrefix(fn, `\`)
		}
		if err := os.Remove(fn); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Ignore file-not-exists errors when removing the socket
			return err
		}
	}

	l.parsePool.StopAndWait()

	return nil
}
