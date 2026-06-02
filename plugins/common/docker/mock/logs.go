package mock

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"testing"
)

type Logs struct {
	Content     string
	Multiplexed bool
}

func WriteLog(ctx context.Context, t *testing.T, w http.ResponseWriter, msgs <-chan *Logs) {
	t.Helper()

	// We need to flush all content to really transmit it over the wrie and
	// avoid keeping things in the buffer
	f := w.(interface{ Flush() })

	// Write the header first so the client knows stuff succeeds
	w.WriteHeader(http.StatusOK)
	f.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			logLine := []byte(msg.Content)
			if _, err := w.Write(logLine); err != nil {
				t.Logf("writing log line failed: %v", err)
			}
			f.Flush()
			continue
		}
	}
}

func (l *Logs) multiplex() ([]byte, error) {
	// Emulate a multiplexed writer
	var buf bytes.Buffer
	header := [8]byte{0: 1}
	binary.BigEndian.PutUint32(header[4:], uint32(len(l.Content)))
	if _, err := buf.Write(header[:]); err != nil {
		return nil, fmt.Errorf("writing log multiplex header failed: %w", err)
	}
	if _, err := buf.WriteString(l.Content); err != nil {
		return nil, fmt.Errorf("writing log multiplex content failed: %w", err)
	}

	return buf.Bytes(), nil
}
