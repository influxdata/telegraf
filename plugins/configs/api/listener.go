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
	m.HandleFunc("/status", s.status).Methods("GET")
	m.HandleFunc("/plugins/create", s.createPlugin).Methods("POST")
	m.HandleFunc("/plugins/{id:[0-9a-f]+}/status", s.pluginStatus).Methods("GET")
	m.HandleFunc("/plugins/list", s.listPlugins).Methods("GET")
	m.HandleFunc("/plugins/running", s.runningPlugins).Methods("GET")
	m.HandleFunc("/plugins/{id:[0-9a-f]+}", s.deleteOrUpdatePlugin).Methods("DELETE", "PUT")
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
	id, err := s.api.CreatePlugin(cfg, "")
	if err != nil {
		s.Log.Error("error creating plugin %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("W! error writing to connection: %v", err)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("W! error writing to connection: %v", err)
		return
	}
}

func (s *ConfigAPIService) pluginStatus(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	state := s.api.GetPluginStatus(models.PluginID(id))
	w.Header().Set("Content-Type", "application/json")
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

func (s *ConfigAPIService) deleteOrUpdatePlugin(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "DELETE":
		s.deletePlugin(w, req)
	case "PUT":
		s.updatePlugin(w, req)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *ConfigAPIService) deletePlugin(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	id := mux.Vars(req)["id"]
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.api.DeletePlugin(models.PluginID(id)); err != nil {
		s.Log.Error("error deleting plugin %v", err)
		// TODO: improve error handling? Would like to see status based on error type
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (s *ConfigAPIService) updatePlugin(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
