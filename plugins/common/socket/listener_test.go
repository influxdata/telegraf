package socket

import (
	"bytes"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// the following are ai-generated tests for manually chosen properties

// newLoopbackListener starts a PacketStreamListener over a real loopback UDP
// socket and returns it together with the address peers should send to.
func newLoopbackListener(t *testing.T) (*PacketStreamListener, net.Addr) {
	t.Helper()
	ln, err := Listen(func() (net.PacketConn, error) {
		return net.ListenPacket("udp", "127.0.0.1:0")
	})
	require.NoError(t, err)
	return ln, ln.Addr()
}

// dialLoopback opens a connected UDP socket towards addr. Each call gets a
// distinct temporary source port, so the listener sees it as a separate peer.
func dialLoopback(t *testing.T, addr net.Addr) *net.UDPConn {
	t.Helper()
	c, err := net.Dial("udp", addr.String())
	require.NoError(t, err)
	return c.(*net.UDPConn)
}

// readN reads exactly n bytes from c, guarding against a hang with a deadline.
func readN(t *testing.T, c net.Conn, n int) []byte {
	t.Helper()
	require.NoError(t, c.SetReadDeadline(time.Now().Add(3*time.Second)))
	buf := make([]byte, n)
	_, err := io.ReadFull(c, buf)
	require.NoError(t, err)
	require.NoError(t, c.SetReadDeadline(time.Time{}))
	return buf
}

// acceptStreamsByRemote accepts n streams and keys them by their peer address.
func acceptStreamsByRemote(t *testing.T, ln *PacketStreamListener, n int) map[string]net.Conn {
	t.Helper()
	out := make(map[string]net.Conn, n)
	for i := 0; i < n; i++ {
		c, err := ln.Accept()
		require.NoError(t, err)
		out[c.RemoteAddr().String()] = c
	}
	return out
}

func randomBytes(r *rand.Rand, n int) []byte {
	b := make([]byte, n)
	r.Read(b)
	return b
}

func randomPackets(r *rand.Rand, count int) [][]byte {
	packets := make([][]byte, count)
	for i := range packets {
		packets[i] = randomBytes(r, r.Intn(64)+1)
	}
	return packets
}

// Property: for any interleaving of packets from two distinct peers, each peer
// is surfaced as its own stream and each stream reconstructs exactly the byte
// sequence that peer sent, in order.
func TestPropertyDemuxAlternatingHosts(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))

	const iterations = 40
	for iter := 0; iter < iterations; iter++ {
		func() {
			ln, addr := newLoopbackListener(t)
			defer ln.Close()
			clientA := dialLoopback(t, addr)
			defer clientA.Close()
			clientB := dialLoopback(t, addr)
			defer clientB.Close()

			packetsA := randomPackets(r, r.Intn(6)+1)
			packetsB := randomPackets(r, r.Intn(6)+1)

			// Interleave the two peers' packets in a random order.
			ia, ib := 0, 0
			for ia < len(packetsA) || ib < len(packetsB) {
				if ia < len(packetsA) && (ib >= len(packetsB) || r.Intn(2) == 0) {
					_, err := clientA.Write(packetsA[ia])
					require.NoError(t, err)
					ia++
				} else {
					_, err := clientB.Write(packetsB[ib])
					require.NoError(t, err)
					ib++
				}
			}

			streams := acceptStreamsByRemote(t, ln, 2)
			streamA := streams[clientA.LocalAddr().String()]
			streamB := streams[clientB.LocalAddr().String()]
			require.NotNil(t, streamA, "no stream for peer A")
			require.NotNil(t, streamB, "no stream for peer B")

			wantA := bytes.Join(packetsA, nil)
			wantB := bytes.Join(packetsB, nil)
			require.Equal(t, wantA, readN(t, streamA, len(wantA)))
			require.Equal(t, wantB, readN(t, streamB, len(wantB)))
		}()
	}
}

// Property: multiple packets from a single peer are concatenated into one
// stream, preserving the bytes across arbitrary packet boundaries.
func TestPropertyStreamAcrossPackets(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))

	const iterations = 40
	for iter := 0; iter < iterations; iter++ {
		func() {
			ln, addr := newLoopbackListener(t)
			defer ln.Close()
			client := dialLoopback(t, addr)
			defer client.Close()

			packets := make([][]byte, r.Intn(20)+1)
			for i := range packets {
				packets[i] = randomBytes(r, r.Intn(200)+1)
				_, err := client.Write(packets[i])
				require.NoError(t, err)
			}

			c, err := ln.Accept()
			require.NoError(t, err)

			want := bytes.Join(packets, nil)
			require.Equal(t, want, readN(t, c, len(want)))
		}()
	}
}

// Property: a stream write larger than a single UDP packet is split across
// multiple packets transparently and reassembled intact by the peer.
func TestPropertyMessageLargerThanPacket(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))

	const iterations = 20
	for iter := 0; iter < iterations; iter++ {
		func() {
			ln, addr := newLoopbackListener(t)
			defer ln.Close()
			client := dialLoopback(t, addr)
			defer client.Close()

			// Establish the stream so the listener has a peer to write back to.
			_, err := client.Write([]byte("open"))
			require.NoError(t, err)
			c, err := ln.Accept()
			require.NoError(t, err)
			readN(t, c, len("open"))

			// Build a message that spans two or three UDP packets.
			spans := r.Intn(2) + 2
			size := maxPayload*(spans-1) + r.Intn(maxPayload) + 1
			msg := randomBytes(r, size)

			// Read on the peer concurrently so the listener's writes are not
			// dropped by a full receive buffer.
			got := make([]byte, size)
			readErr := make(chan error, 1)
			go func() {
				if err := client.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
					readErr <- err
					return
				}
				_, err := io.ReadFull(client, got)
				readErr <- err
			}()

			n, err := c.Write(msg)
			require.NoError(t, err)
			require.Equal(t, size, n)

			require.NoError(t, <-readErr)
			require.Equal(t, msg, got)
		}()
	}
}

// Property: closing a stream deregisters its peer, so a later packet from the
// same source address starts a fresh stream rather than resurrecting the
// closed one.
func TestPropertyClosedStreamReopens(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))

	const iterations = 20
	for iter := 0; iter < iterations; iter++ {
		func() {
			ln, addr := newLoopbackListener(t)
			defer ln.Close()
			client := dialLoopback(t, addr)
			defer client.Close()

			first := randomBytes(r, r.Intn(64)+1)
			_, err := client.Write(first)
			require.NoError(t, err)

			s1, err := ln.Accept()
			require.NoError(t, err)
			require.Equal(t, first, readN(t, s1, len(first)))
			require.NoError(t, s1.Close())

			// A new packet from the same source port must yield a new stream
			// carrying the new bytes, not the closed one.
			second := randomBytes(r, r.Intn(64)+1)
			_, err = client.Write(second)
			require.NoError(t, err)

			s2, err := ln.Accept()
			require.NoError(t, err)
			require.Equal(t, s1.RemoteAddr().String(), s2.RemoteAddr().String())
			require.NotSame(t, s1.(*conn), s2.(*conn))
			require.Equal(t, second, readN(t, s2, len(second)))
		}()
	}
}

// Property: closing the listener terminates every live stream (reads report
// net.ErrClosed once buffered bytes drain) and stops Accept.
func TestPropertyClosedListenerClosesStreams(t *testing.T) {
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	r := rand.New(rand.NewSource(seed))

	const iterations = 20
	for iter := 0; iter < iterations; iter++ {
		func() {
			ln, addr := newLoopbackListener(t)

			n := r.Intn(4) + 1
			clients := make([]*net.UDPConn, n)
			payloads := make([][]byte, n)
			for i := range clients {
				clients[i] = dialLoopback(t, addr)
				defer clients[i].Close()
				payloads[i] = randomBytes(r, r.Intn(32)+1)
				_, err := clients[i].Write(payloads[i])
				require.NoError(t, err)
			}

			// Accept every peer and drain its initial payload so the buffers
			// are empty before the listener closes.
			byRemote := acceptStreamsByRemote(t, ln, n)
			streams := make([]net.Conn, n)
			for i := range clients {
				s := byRemote[clients[i].LocalAddr().String()]
				require.NotNil(t, s, "no stream for peer %d", i)
				require.Equal(t, payloads[i], readN(t, s, len(payloads[i])))
				streams[i] = s
			}

			require.NoError(t, ln.Close())

			// Every accepted stream now reports closed on read.
			for i, s := range streams {
				require.NoError(t, s.SetReadDeadline(time.Now().Add(3*time.Second)))
				_, err := s.Read(make([]byte, 1))
				require.ErrorIsf(t, err, net.ErrClosed, "stream %d", i)
			}

			// Accept no longer yields streams.
			_, err := ln.Accept()
			require.ErrorIs(t, err, net.ErrClosed)
		}()
	}
}
