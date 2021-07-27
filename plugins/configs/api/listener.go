package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log" // nolint:revive
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
)

type ConfigAPIService struct {
	server *http.Server
	api    *api
	Log    telegraf.Logger
}

func newConfigAPIService(server *http.Server, api *api, logger telegraf.Logger) *ConfigAPIService {
	service := &ConfigAPIService{
		server: server,
		api:    api,
		Log:    logger,
	}
	server.Handler = service.mux()
	return service
}

// nolint:revive
func (s *ConfigAPIService) mux() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/status", s.status)
	m.HandleFunc("/plugins/create", s.createPlugin)
	m.HandleFunc("/plugins/{id:[0-9a-f]+}/status", s.pluginStatus)
	m.HandleFunc("/plugins/list", s.listPlugins)
	m.HandleFunc("/plugins/running", s.runningPlugins)
	return m
}

func (s *ConfigAPIService) status(w http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	_, _ = w.Write([]byte("ok"))
}

func (s *ConfigAPIService) createPlugin(w http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	cfg := PluginConfigCreate{}

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&cfg); err != nil {
		s.Log.Error("decode error %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := s.api.CreatePlugin(cfg)
	if err != nil {
		s.Log.Error("error creating plugin %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, _ = w.Write([]byte(fmt.Sprintf(`{"id": "%s"}`, id)))
}

func (s *ConfigAPIService) Start() {
	// if s.server.TLSConfig != nil {
	// 	s.server.ListenAndServeTLS()
	// }
	go func() {
		_ = s.server.ListenAndServe()
	}()
}

func (s *ConfigAPIService) listPlugins(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	typeInfo := s.api.ListPluginTypes()

	bytes, err := json.Marshal(typeInfo)
	if err != nil {
		log.Printf("!E [configapi] error marshalling json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(bytes)
}

func (s *ConfigAPIService) runningPlugins(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	plugins := s.api.ListRunningPlugins()

	bytes, err := json.Marshal(plugins)
	if err != nil {
		log.Printf("!E [configapi] error marshalling json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(bytes)
}

func (s *ConfigAPIService) pluginStatus(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	state := s.api.GetPluginStatus(models.PluginID(id))
	_, err := w.Write([]byte(fmt.Sprintf(`{"status": %q}`, state.String())))
	if err != nil {
		log.Printf("W! error writing to connection: %v", err)
		return
	}
}

func (s *ConfigAPIService) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("W! [configapi] error on shutdown: %s", err)
	}
}
