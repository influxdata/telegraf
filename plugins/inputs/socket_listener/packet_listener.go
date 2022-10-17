package socket_listener

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type packetListener struct {
	Encoding       string
	SocketMode     string
	ReadBufferSize int
	Parser         telegraf.Parser
	Log            telegraf.Logger

	conn    net.PacketConn
	decoder internal.ContentDecoder
	path    string
}

func (l *packetListener) listen(acc telegraf.Accumulator) {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := l.conn.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				acc.AddError(err)
			}
			break
		}

		body, err := l.decoder.Decode(buf[:n])
		if err != nil {
			acc.AddError(fmt.Errorf("unable to decode incoming packet: %w", err))
		}

		metrics, err := l.Parser.Parse(body)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse incoming packet: %w", err))
			// TODO rate limit
			continue
		}
		for _, m := range metrics {
			acc.AddMetric(m)
		}
	}
}

func (l *packetListener) setupUnixgram(u *url.URL, socketMode string) error {
	err := os.Remove(u.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing socket failed: %w", err)
	}

	conn, err := net.ListenPacket(u.Scheme, u.Path)
	if err != nil {
		return fmt.Errorf("listening (unixgram) failed: %w", err)
	}
	l.path = u.Path
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

	// Create a decoder for the given encoding
	decoder, err := internal.NewContentDecoder(l.Encoding)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}
	l.decoder = decoder

	return nil
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

	// Create a decoder for the given encoding
	decoder, err := internal.NewContentDecoder(l.Encoding)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}
	l.decoder = decoder

	return nil
}

func (l *packetListener) setupIP(u *url.URL) error {
	conn, err := net.ListenPacket(u.Scheme, u.Host)
	if err != nil {
		return fmt.Errorf("listening (ip) failed: %w", err)
	}
	l.conn = conn

	// Create a decoder for the given encoding
	decoder, err := internal.NewContentDecoder(l.Encoding)
	if err != nil {
		return fmt.Errorf("creating decoder failed: %w", err)
	}
	l.decoder = decoder

	return nil
}

func (l *packetListener) addr() net.Addr {
	return l.conn.LocalAddr()
}

func (l *packetListener) close() error {
	if err := l.conn.Close(); err != nil {
		return err
	}

	if l.path != "" {
		err := os.Remove(l.path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			// Ignore file-not-exists errors when removing the socket
			return err
		}
	}

	return nil
}
