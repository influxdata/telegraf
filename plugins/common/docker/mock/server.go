package mock

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

const apiVersion = "1.54"

type Logs struct {
	Content     string
	Multiplexed bool
}

type Server struct {
	List    []container.Summary
	Inspect map[string]container.InspectResponse
	Logs    map[string]Logs

	server *httptest.Server
}

func (s *Server) Start(t *testing.T) string {
	t.Helper()

	s.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response []byte

		parts := strings.Split(r.URL.Path, "/")
		switch {
		case r.URL.Path == "/_ping":
			// Ping response
			var err error
			response, err = json.Marshal(&client.PingResult{
				APIVersion: apiVersion,
				OSType:     "linux/amd64",
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal ping response: %v", err)
				return
			}
		case r.URL.Path == "/v"+apiVersion+"/containers/json":
			// List response
			var err error
			response, err = json.Marshal(s.List)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal list response: %v", err)
				return
			}
		case strings.HasPrefix(r.URL.Path, "/v"+apiVersion+"/containers/") &&
			len(parts) == 5 &&
			strings.HasSuffix(r.URL.Path, "/json"):
			// Inspect response
			id := parts[3]
			data, found := s.Inspect[id]
			if !found {
				w.WriteHeader(http.StatusNotFound)
				t.Errorf("inspect response for %q not found", id)
				return
			}
			var err error
			response, err = json.Marshal(data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal inspect response: %v", err)
				return
			}
		case strings.HasPrefix(r.URL.Path, "/v"+apiVersion+"/containers/") &&
			len(parts) == 5 &&
			strings.HasSuffix(r.URL.Path, "/logs"):
			// Logs response
			id := parts[3]
			data, found := s.Logs[id]
			if !found {
				w.WriteHeader(http.StatusNotFound)
				t.Errorf("log response for %q not found", id)
				return
			}
			if data.Multiplexed {
				// Emulate a multiplexed writer
				var buf bytes.Buffer
				header := [8]byte{0: 1}
				binary.BigEndian.PutUint32(header[4:], uint32(len(data.Content)))
				if _, err := buf.Write(header[:]); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("writing log multiplex header failed: %v", err)
					return
				}
				if _, err := buf.WriteString(data.Content); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("writing log multiplex content failed: %v", err)
					return
				}
				response = buf.Bytes()
			} else {
				response = []byte(data.Content)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			t.Errorf("unhandled url: %q (len: %d)", r.URL.Path, len(parts))
			return
		}

		if _, err := w.Write(response); err != nil {
			t.Errorf("failed to write response: %v", err)
			return
		}
	}))

	return s.server.URL
}

func (s *Server) Close() {
	if s.server != nil {
		s.server.Close()
	}
}
