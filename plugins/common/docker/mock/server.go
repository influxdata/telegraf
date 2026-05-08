package mock

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
)

type Logs struct {
	Content     string
	Multiplexed bool
}

type Server struct {
	Info     system.Info
	List     []container.Summary
	Disks    container.DiskUsage
	RawDisks []byte // Raw response as the disks response format depends on the API version
	Inspect  map[string]container.InspectResponse
	Stats    map[string]container.StatsResponse
	Logs     map[string]Logs

	Services []swarm.Service
	Tasks    []swarm.Task
	Nodes    []swarm.Node

	APIVersion string

	ListParams map[string]string

	server *httptest.Server
}

func NewServerFromFiles(path string) (*Server, error) {
	var s Server

	// Read info
	if _, err := os.Stat(filepath.Join(path, "info.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "info.json"))
		if err != nil {
			return nil, fmt.Errorf("reading info failed: %w", err)
		}
		if err := json.Unmarshal(buf, &s.Info); err != nil {
			return nil, fmt.Errorf("parsing info failed: %w", err)
		}
	}

	// Read container list
	if _, err := os.Stat(filepath.Join(path, "list.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "list.json"))
		if err != nil {
			return nil, fmt.Errorf("reading container list failed: %w", err)
		}
		if err := json.Unmarshal(buf, &s.List); err != nil {
			return nil, fmt.Errorf("parsing container list failed: %w", err)
		}
	}

	// Read container statistics data
	matches, err := filepath.Glob(filepath.Join(path, "stats_*.json"))
	if err != nil {
		return nil, fmt.Errorf("matching stats failed: %w", err)
	}
	s.Stats = make(map[string]container.StatsResponse, len(matches))
	for _, fn := range matches {
		buf, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("reading stats %q failed: %w", fn, err)
		}
		var stats container.StatsResponse
		if err := json.Unmarshal(buf, &stats); err != nil {
			return nil, fmt.Errorf("parsing stats %q failed: %w", fn, err)
		}
		id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(fn), "stats_"), ".json")
		s.Stats[id] = stats
	}

	// Read container inspection data
	matches, err = filepath.Glob(filepath.Join(path, "inspect_*.json"))
	if err != nil {
		return nil, fmt.Errorf("matching stats failed: %w", err)
	}
	s.Inspect = make(map[string]container.InspectResponse, len(matches))
	for _, fn := range matches {
		buf, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("reading inspection data failed: %w", err)
		}
		var r container.InspectResponse
		if err := json.Unmarshal(buf, &r); err != nil {
			return nil, fmt.Errorf("parsing inspection data failed: %w", err)
		}
		s.Inspect[r.ID] = r
	}

	// Read service data
	if _, err := os.Stat(filepath.Join(path, "services.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "services.json"))
		if err != nil {
			return nil, fmt.Errorf("reading services failed: %w", err)
		}
		if err := json.Unmarshal(buf, &s.Services); err != nil {
			return nil, fmt.Errorf("parsing services failed: %w", err)
		}
	}

	// Read task data
	if _, err := os.Stat(filepath.Join(path, "tasks.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "tasks.json"))
		if err != nil {
			return nil, fmt.Errorf("reading tasks failed: %w", err)
		}
		if err := json.Unmarshal(buf, &s.Tasks); err != nil {
			return nil, fmt.Errorf("parsing tasks failed: %w", err)
		}
	}
	// Read node data
	if _, err := os.Stat(filepath.Join(path, "nodes.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "nodes.json"))
		if err != nil {
			return nil, fmt.Errorf("reading nodes failed: %w", err)
		}
		if err := json.Unmarshal(buf, &s.Nodes); err != nil {
			return nil, fmt.Errorf("parsing nodes failed: %w", err)
		}
	}

	// Read disk usage
	if _, err := os.Stat(filepath.Join(path, "disk.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "disk.json"))
		if err != nil {
			return nil, fmt.Errorf("reading disk failed: %w", err)
		}
		s.RawDisks = make([]byte, len(buf))
		copy(s.RawDisks, buf)

		if err := json.Unmarshal(buf, &s.Disks); err != nil {
			return nil, fmt.Errorf("parsing disk failed: %w", err)
		}
	}

	return &s, nil
}

func (s *Server) Start(t *testing.T) string {
	t.Helper()

	if s.APIVersion == "" {
		s.APIVersion = "1.54"
	}

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if agent := r.Header.Get("User-Agent"); agent != "engine-api-cli-1.0" {
			w.WriteHeader(http.StatusInternalServerError)
			t.Errorf("invalid user-agent %q", agent)
			return
		}

		var response []byte
		parts := strings.Split(r.URL.Path, "/")
		switch {
		case r.URL.Path == "/_ping":
			// Ping response
			var err error
			response, err = json.Marshal(&client.PingResult{
				APIVersion: s.APIVersion,
				OSType:     "linux/amd64",
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal ping response: %v", err)
				return
			}
		case r.URL.Path == "/v"+s.APIVersion+"/info":
			// Info response
			var err error
			response, err = json.Marshal(s.Info)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal info response: %v", err)
				return
			}
		case r.URL.Path == "/v"+s.APIVersion+"/containers/json":
			q := r.URL.Query()
			for k, v := range s.ListParams {
				if q.Get(k) != v {
					w.WriteHeader(http.StatusBadRequest)
					t.Errorf("invalid list parameter for %q: %q (expected %q)", k, q.Get(k), v)
					return
				}
			}

			// List response
			var err error
			response, err = json.Marshal(s.List)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal list response: %v", err)
				return
			}
		case strings.HasPrefix(r.URL.Path, "/v"+s.APIVersion+"/containers/") &&
			len(parts) == 5 && strings.HasSuffix(r.URL.Path, "/json"):
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
		case strings.HasPrefix(r.URL.Path, "/v"+s.APIVersion+"/containers/") &&
			len(parts) == 5 && strings.HasSuffix(r.URL.Path, "/logs"):
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
		case strings.HasPrefix(r.URL.Path, "/v"+s.APIVersion+"/containers/") &&
			len(parts) == 5 && strings.HasSuffix(r.URL.Path, "/stats"):
			// Statistics response
			id := parts[3]
			data, found := s.Stats[id]
			if !found {
				w.WriteHeader(http.StatusNotFound)
				t.Errorf("stats response for %q not found", id)
				return
			}
			var err error
			response, err = json.Marshal(data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal stats response: %v", err)
				return
			}
		case r.URL.Path == "/v"+s.APIVersion+"/system/df":
			// Disk usage response
			version, err := semver.NewVersion(s.APIVersion)
			if err != nil || version.Major() < 1 || (version.Major() == 1 && version.Minor() < 52) {
				// For old versions we simply send the raw disk data
				response = s.RawDisks
			} else {
				// New versions use the correct response format
				response, err = json.Marshal(s.Disks)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("failed to marshal disk response: %v", err)
					return
				}
			}
		case r.URL.Path == "/v"+s.APIVersion+"/services":
			// Swarm services response
			var err error
			response, err = json.Marshal(s.Services)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal services response: %v", err)
				return
			}
		case r.URL.Path == "/v"+s.APIVersion+"/tasks":
			// Swarm tasks response
			var err error
			response, err = json.Marshal(s.Tasks)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal tasks response: %v", err)
				return
			}
		case r.URL.Path == "/v"+s.APIVersion+"/nodes":
			// Swarm nodes response
			var err error
			response, err = json.Marshal(s.Nodes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Errorf("failed to marshal nodes response: %v", err)
				return
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
