package assistant

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/internal"
)

/*
Assistant is a client to facilitate communications between Agent and Cloud.
*/
type Assistant struct {
	config *AssistantConfig // Configuration for Assitant's conn to server
	conn   *websocket.Conn  // Active websocket conn
	running bool
	agent  *agent.Agent     // Pointer to agent to issue commands
}

/*
AssistantConfig allows us to configure where to connect to and other params
for the agent.
*/
type AssistantConfig struct {
	Host          string
	Path          string
	RetryInterval int
}

func NewAssistantConfig() *AssistantConfig {
	return &AssistantConfig{
		Host: "localhost:8080",
		Path: "/assistant",
		RetryInterval: 15,
	}
}

// NewAssistant returns an Assistant for the given Config.
func NewAssistant(config *AssistantConfig, agent *agent.Agent) *Assistant {
	return &Assistant{
		config: config,
		agent:  agent,
	}
}

type pluginInfo struct {
	Name     string
	Type     string
	Config   map[string]interface{}
	UniqueId string
}

type requestType string

const (
	GET_PLUGIN          = requestType("GET_PLUGIN")
	GET_PLUGIN_SCHEMA   = requestType("GET_PLUGIN_SCHEMA")
	UPDATE_PLUGIN       = requestType("UPDATE_PLUGIN")
	START_PLUGIN        = requestType("START_PLUGIN")
	STOP_PLUGIN         = requestType("STOP_PLUGIN")
	GET_RUNNING_PLUGINS = requestType("GET_RUNNING_PLUGINS")
	GET_ALL_PLUGINS     = requestType("GET_ALL_PLUGINS")

	SUCCESS = "SUCCESS"
	FAILURE = "FAILURE"
)

type request struct {
	Operation requestType
	UUID      string
	Plugin    pluginInfo
}

type response struct {
	Status string
	UUID   string
	Data   interface{}
}

func (a *Assistant) init(ctx context.Context) error {
	token, exists := os.LookupEnv("INFLUX_TOKEN")
	if !exists {
		return fmt.Errorf("influx authorization token not found, please set in env")
	}

	header := http.Header{}
	header.Add("Authorization", "Token " + token)
	u := url.URL{Scheme: "ws", Host: a.config.Host, Path: a.config.Path}

	log.Printf("D! [assistant] Attempting conn to [%s]", a.config.Host)
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	for err != nil { // on error, retry conn again
		log.Printf("E! [assistant] Failed to connect to [%s] due to: '%s'. Retrying in %ds... ",
			a.config.Host, err, a.config.RetryInterval)

		err = internal.SleepContext(ctx, time.Duration(a.config.RetryInterval)*time.Second)
		if err != nil {
			// Return because context was closed
			return err
		}

		ws, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	}
	a.conn = ws

	log.Printf("D! [assistant] Successfully connected to %s", a.config.Host)
	return nil
}

// Run starts the assistant listening to the server and handles and interrupts or closed connections
func (a *Assistant) Run(ctx context.Context) error {
	err := a.init(ctx)
	if err != nil {
		log.Printf("E! [assistant] connection could not be established: %s", err.Error())
		return err
	}
	a.running = true

	go a.listen(ctx)

	return nil
}

// listenToServer takes requests from the server and puts it in Requests.
func (a *Assistant) listen(ctx context.Context) {
	defer a.conn.Close()

	go a.shutdownOnContext(ctx)

	for {
		var req request
		if err := a.conn.ReadJSON(&req); err != nil {
			if !a.running {
				log.Printf("I! [assistant] listener shutting down...")
				return
			}

			log.Printf("E! [assistant] error while reading from server: %s", err)
			// retrying a new websocket connection
			err := a.init(ctx)
			if err != nil {
				log.Printf("E! [assistant] connection could not be re-established: %s", err)
				return
			}
			err = a.conn.ReadJSON(&req)
			if err != nil {
				log.Printf("E! [assistant] re-established connection but could not read server request: %s", err)
				return
			}
		}
		res := a.handleRequest(ctx, &req)

		if err := a.conn.WriteJSON(res); err != nil {
			log.Printf("E! [assistant] Error while writing to server: %s", err)
			a.conn.WriteJSON(response{FAILURE, req.UUID, "error marshalling config"})
		}
	}
}

func (a *Assistant) shutdownOnContext(ctx context.Context) {
	<-ctx.Done()
	a.running = false
	a.conn.Close()
}

func (a *Assistant) handleRequest(ctx context.Context, req *request) response {
	var resp interface{}
	var err error

	switch req.Operation {
	case GET_PLUGIN:
		resp, err = a.getPlugin(req)
	case GET_PLUGIN_SCHEMA:
		resp, err = a.getSchema(req)
	case START_PLUGIN:
		resp, err = a.startPlugin(ctx, req)
	case STOP_PLUGIN:
		resp, err = a.stopPlugin(req)
	case UPDATE_PLUGIN:
		resp, err = a.updatePlugin(req)
	case GET_RUNNING_PLUGINS:
		resp, err = a.getRunningPlugins(req)
	case GET_ALL_PLUGINS:
		resp, err = a.getAllPlugins(req)
	default:
		err = errors.New("invalid operation")
	}

	if err != nil {
		return response{FAILURE, req.UUID, err.Error()}
	}

	return response{SUCCESS, req.UUID, resp}
}

// getPlugin returns the struct response containing config for a single plugin
func (a *Assistant) getPlugin(req *request) (interface{}, error) {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	resp, err := a.agent.GetRunningPlugin(req.Plugin.UniqueId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type pluginSchema struct {
	Schema map[string]interface{}
	Defaults map[string]interface{}
}

// getSchema returns the struct response containing config schema for a single plugin
func (a *Assistant) getSchema(req *request) (interface{}, error) {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var plugin interface{}
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		plugin, err = a.agent.CreateInput(req.Plugin.Name)
	case "OUTPUT":
		plugin, err = a.agent.CreateOutput(req.Plugin.Name)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		return nil, err
	}

	schema, err := a.agent.GetPluginTypes(plugin)
	if err != nil {
		return nil, err
	}

	defaults, err := a.agent.GetPluginValues(plugin)
	if err != nil {
		return nil, err
	}

	resp := pluginSchema{
		Schema: schema,
		Defaults: defaults,
	}

	return resp, nil
}

// startPlugin starts a single plugin
func (a *Assistant) startPlugin(ctx context.Context, req *request) (interface{}, error) {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")
	var uid string
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		uid, err = a.agent.StartInput(ctx, req.Plugin.Name)
	case "OUTPUT":
		uid, err = a.agent.StartOutput(ctx, req.Plugin.Name)
	default:
		err = fmt.Errorf("invalid plugin type")
	}

	if err != nil {
		return nil, err
	}

	resp := map[string]string{
		"id": uid,
	}
	return resp, nil
}

// updatePlugin updates a plugin with the config specified in request
func (a *Assistant) updatePlugin(req *request) (interface{}, error) {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.UniqueId, "\n")

	if req.Plugin.Config == nil {
		return nil, errors.New("no configuration values provided")
	}

	var data interface{}
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		data, err = a.agent.UpdateInputPlugin(req.Plugin.UniqueId, req.Plugin.Config)
	case "OUTPUT":
		data, err = a.agent.UpdateOutputPlugin(req.Plugin.UniqueId, req.Plugin.Config)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

// stopPlugin stops a single plugin given a request
func (a *Assistant) stopPlugin(req *request) (interface{}, error) {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var err error
	switch req.Plugin.Type {
	case "INPUT":
		err = a.agent.StopInputPlugin(req.Plugin.UniqueId, true)
	case "OUTPUT":
		err = a.agent.StopOutputPlugin(req.Plugin.UniqueId, true)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("%s stopped", req.Plugin.UniqueId), nil
}

type runningPlugins struct {
	Inputs  []map[string]string
	Outputs []map[string]string
}

// getRunningPlugins returns an object with all running plugins
func (assistant *Assistant) getRunningPlugins(req *request) (interface{}, error) {
	runningPlugins := runningPlugins{
		Inputs: assistant.agent.GetRunningInputPlugins(),
		Outputs: assistant.agent.GetRunningOutputPlugins(),
	}
	return runningPlugins, nil
}

type availablePlugins struct {
	Inputs  []string
	Outputs []string
}

// getAllPlugins returns an object with the names of all available plugins
func (assistant *Assistant) getAllPlugins(req *request) (interface{}, error) {
	availablePlugins := availablePlugins{
		Inputs: agent.GetAllInputPlugins(),
		Outputs: agent.GetAllOutputPlugins(),
	}
	return availablePlugins, nil
}
