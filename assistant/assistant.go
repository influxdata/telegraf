package assistant

import (
	"context"
	"flag"
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
	config     *AssistantConfig // Configuration for Assitant's connection to server
	connection *websocket.Conn  // Active websocket connection
	done       chan struct{}    // Channel used to stop server listener
	agent      *agent.Agent     // Pointer to agent to issue commands
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

func (astConfig *AssistantConfig) fillDefaults() {
	if astConfig.Host == "" {
		astConfig.Host = "localhost:8080"
	}
	if astConfig.Path == "" {
		astConfig.Path = "/echo"
	}
	if astConfig.RetryInterval == 0 {
		astConfig.RetryInterval = 15
	}
}

// NewAssistant returns an Assistant for the given Config.
func NewAssistant(config *AssistantConfig, agent *agent.Agent) (*Assistant, error) {
	config.fillDefaults()

	a := &Assistant{
		config: config,
		done:   make(chan struct{}),
		agent:  agent,
	}

	return a, nil
}

// Stop is used to clean up active connection and all channels
func (assistant *Assistant) Stop() {
	assistant.done <- struct{}{}
}

type pluginInfo struct {
	Name   string
	Type   string
	Config map[string]interface{}
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

func (assistant *Assistant) initWebsocketConnection(ctx context.Context) error {
	var config = assistant.config
	u := url.URL{Scheme: "ws", Host: config.Host, Path: config.Path}

	header := http.Header{}

	if v, exists := os.LookupEnv("INFLUX_TOKEN"); exists {
		header.Add("Authorization", "Token "+v)
	} else {
		return fmt.Errorf("influx authorization token not found, please set in env")
	}

	log.Printf("D! [assistant] Attempting connection to [%s]", config.Host)
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	for err != nil { // on error, retry connection again
		// TODO? Do we really want this in here? Consider move outside
		log.Printf("E! [assistant] Failed to connect to [%s], retrying in %ds, "+
			"error was '%s'", config.Host, config.RetryInterval, err)

		sleepErr := internal.SleepContext(ctx, time.Duration(config.RetryInterval)*time.Second)
		if sleepErr != nil {
			return sleepErr
		}

		ws, _, err = websocket.DefaultDialer.Dial(u.String(), header)
	}
	log.Printf("D! [assistant] Successfully connected to %s", config.Host)
	assistant.connection = ws

	return nil
}

// Run starts the assistant listening to the server and handles and interrupts or closed connections
func (assistant *Assistant) Run(ctx context.Context) error {
	defer assistant.connection.Close()

	var config = assistant.config
	var addr = flag.String("addr", config.Host, "http service address")
	config.Host = *addr

	err := assistant.initWebsocketConnection(ctx)
	if err != nil {
		log.Printf("E! [assistant] connection could not be established: %s", err)
		return err
	}

	go assistant.listenToServer(ctx)

	for {
		select {
		case <-assistant.done:
			return nil
		case <-ctx.Done():
			log.Printf("I! [assistant] Hang on, closing connection to server before shutdown")
			return nil
		}
	}
}

// listenToServer takes requests from the server and puts it in Requests.
func (assistant *Assistant) listenToServer(ctx context.Context) {
	defer close(assistant.done)
	for {
		var req request
		err := assistant.connection.ReadJSON(&req)
		if err != nil {
			log.Printf("E! [assistant] error while reading from server: %s", err)
			// retrying a new websocket connection
			err := assistant.initWebsocketConnection(ctx)
			if err != nil {
				log.Printf("E! [assistant] connection could not be re-established: %s", err)
				return
			}
			err = assistant.connection.ReadJSON(&req)
			if err != nil {
				log.Printf("E! [assistant] re-established connection but could not read server request: %s", err)
				return
			}
		}
    res := assistant.handleRequests(&req)
		switch req.Operation {
		case GET_PLUGIN:
			res = assistant.getPlugin(req)
		case GET_PLUGIN_SCHEMA:
			res = assistant.getSchema(req)
		case START_PLUGIN:
			data, err := assistant.startPlugin(req)
			if err != nil || req.Plugin.Config == nil {
				res = data // either add plugin failed, or we init'd with default config only
			} else {
				res = assistant.updatePlugin(req) // update default config with specific config
			}
		case STOP_PLUGIN:
			res = assistant.stopPlugin(req)
		case UPDATE_PLUGIN:
			res = assistant.updatePlugin(req)
		case GET_RUNNING_PLUGINS:
			res = assistant.getRunningPlugins(req)
		case GET_ALL_PLUGINS:
			res = assistant.getAllPlugins(req)
		default:
			// return error response
			res = response{FAILURE, req.UUID, "invalid operation request"}
		}
		err = assistant.connection.WriteJSON(res)
		if err != nil {
			// log error and keep connection open
			// TODO retry write to server, something wrong with error response.
			log.Printf("E! [assistant] Error while writing to server: %s", err)
			assistant.connection.WriteJSON(response{FAILURE, req.UUID, "error marshalling config"})
		}
	}
}

func (assistant *Assistant) handleRequests(req *request) response {
	var res response
	switch req.Operation {
	case GET_PLUGIN:
		res = assistant.getPlugin(req)
	case GET_PLUGIN_SCHEMA:
		res = assistant.getSchema(req)
	case START_PLUGIN:
		res = assistant.startPlugin(req)
	case STOP_PLUGIN:
		res = assistant.stopPlugin(req)
	case UPDATE_PLUGIN:
		res = assistant.updatePlugin(req)
	case GET_RUNNING_PLUGINS:
		res = assistant.getRunningPlugins(req)
	case GET_ALL_PLUGINS:
		res = assistant.getAllPlugins(req)
	default:
		// return error response
		res = response{FAILURE, req.UUID, "invalid operation request"}
	}
	return res
}

// getPlugin returns the struct response containing config for a single plugin
func (assistant *Assistant) getPlugin(req *request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var data interface{}
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		data, err = assistant.agent.GetRunningInputPlugin(req.Plugin.Name)
	case "OUTPUT":
		data, err = assistant.agent.GetRunningOutputPlugin(req.Plugin.Name)
	case "AGGREGATOR":
		data, err = assistant.agent.GetAggregatorPlugin(req.Plugin.Name)
	case "PROCESSOR":
		data, err = assistant.agent.GetProcessorPlugin(req.Plugin.Name)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		return response{FAILURE, req.UUID, err.Error()}
	}

	return response{SUCCESS, req.UUID, data}
}

type schema struct {
	Types    map[string]interface{}
	Defaults interface{}
}

// getSchema returns the struct response containing config schema for a single plugin
func (assistant *Assistant) getSchema(req *request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var plugin interface{}
	var err error
	switch req.Plugin.Type {
	case "INPUT":
		plugin, err = assistant.agent.GetDefaultInputPlugin(req.Plugin.Name)
	case "OUTPUT":
		plugin, err = assistant.agent.GetDefaultOutputPlugin(req.Plugin.Name)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}
	if err != nil {
		return response{FAILURE, req.UUID, err.Error()}
	}

	var types map[string]interface{}
	types, err = assistant.agent.GetPluginTypes(plugin)
	if err != nil {
		return response{FAILURE, req.UUID, err.Error()}
	}
	return response{SUCCESS, req.UUID, schema{types, plugin}}
}

// startPlugin starts a single plugin
func (assistant *Assistant) startPlugin(req *request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var res response
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		err = assistant.agent.StartInput(req.Plugin.Name)
	case "OUTPUT":
		err = assistant.agent.StartOutput(req.Plugin.Name, req.Plugin.Config)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		res = response{FAILURE, req.UUID, err.Error()}
	} else {
		res = response{SUCCESS, req.UUID, fmt.Sprintf("%s plugin added.", req.Plugin.Name)}
	}

	return res
}

// updatePlugin updates a plugin with the config specified in request
func (assistant *Assistant) updatePlugin(req *request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var res response
	var data interface{}
	var err error

	if req.Plugin.Config == nil {
		res = response{FAILURE, req.UUID, "no config specified!"}
		return res
	}

	switch req.Plugin.Type {
	case "INPUT":
		data, err = assistant.agent.UpdateInputPlugin(req.Plugin.Name, req.Plugin.Config)
	case "OUTPUT":
		data, err = assistant.agent.UpdateOutputPlugin(req.Plugin.Name, req.Plugin.Config)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		res = response{FAILURE, req.UUID, err.Error()}
	} else {
		res = response{SUCCESS, req.UUID, data}
	}

	return res
}

// stopPlugin stops a single plugin
func (assistant *Assistant) stopPlugin(req *request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var res response
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		assistant.agent.StopInputPlugin(req.Plugin.Name, true)
	case "OUTPUT":
    assistant.agent.StopOutputPlugin(req.Plugin.Name, true)
	default:
		err = fmt.Errorf("did not provide a valid plugin type")
	}

	if err != nil {
		res = response{FAILURE, req.UUID, err.Error()}
	} else {
		res = response{SUCCESS, req.UUID, fmt.Sprintf("%s plugin deleted.", req.Plugin.Name)}
	}

	return res
}

type pluginsList struct {
	Inputs  []string
	Outputs []string
}

// getRunningPlugins returns a JSON response obj with all running plugins
func (assistant *Assistant) getRunningPlugins(req *request) response {
	inputs := assistant.agent.GetRunningInputPlugins()
	outputs := assistant.agent.GetRunningOutputPlugins()
	data := pluginsList{inputs, outputs}

	res := response{SUCCESS, req.UUID, data}
	return res
}

// getAllPlugins returns a JSON response obj with names of all possible plugins
func (assistant *Assistant) getAllPlugins(req *request) response {
	inputs := agent.GetAllInputPlugins()
	outputs := agent.GetAllOutputPlugins()
	data := pluginsList{inputs, outputs}
	res := response{SUCCESS, req.UUID, data}
	return res
}
