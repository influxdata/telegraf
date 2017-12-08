// Package logtransport provides a transport that logs all of its messages.
package logtransport

import (
	"bytes"
	"io"
	"log"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2/encoding/text"
	"zombiezen.com/go/capnproto2/rpc"
	"zombiezen.com/go/capnproto2/rpc/internal/logutil"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

type transport struct {
	rpc.Transport
	l       *log.Logger
	sendBuf bytes.Buffer
	recvBuf bytes.Buffer
}

// New creates a new logger that proxies messages to and from t and
// logs them to l.  If l is nil, then the log package's default
// logger is used.
func New(l *log.Logger, t rpc.Transport) rpc.Transport {
	return &transport{Transport: t, l: l}
}

func (t *transport) SendMessage(ctx context.Context, msg rpccapnp.Message) error {
	t.sendBuf.Reset()
	t.sendBuf.WriteString("<- ")
	formatMsg(&t.sendBuf, msg)
	logutil.Print(t.l, t.sendBuf.String())
	return t.Transport.SendMessage(ctx, msg)
}

func (t *transport) RecvMessage(ctx context.Context) (rpccapnp.Message, error) {
	msg, err := t.Transport.RecvMessage(ctx)
	if err != nil {
		return msg, err
	}
	t.recvBuf.Reset()
	t.recvBuf.WriteString("-> ")
	formatMsg(&t.recvBuf, msg)
	logutil.Print(t.l, t.recvBuf.String())
	return msg, nil
}

func formatMsg(w io.Writer, m rpccapnp.Message) {
	text.NewEncoder(w).Encode(0x91b79f1f808db032, m.Struct)
}
