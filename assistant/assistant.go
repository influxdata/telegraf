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

type plugin struct {
	Name   string
	Type   string
	Config map[string]interface{}
}

type requestType string

const (
	GET_PLUGIN          = requestType("GET_PLUGIN")
	ADD_PLUGIN          = requestType("ADD_PLUGIN")
	UPDATE_PLUGIN       = requestType("UPDATE_PLUGIN")
	DELETE_PLUGIN       = requestType("DELETE_PLUGIN")
	GET_ALL_PLUGINS     = requestType("GET_ALL_PLUGINS")
	GET_RUNNING_PLUGINS = requestType("GET_RUNNING_PLUGINS")
	SUCCESS             = "SUCCESS"
	FAILURE             = "FAILURE"
)

type request struct {
	Operation requestType
	UUID      string
	Plugin    plugin
}

type response struct {
	Status string
	UUID   string
	Data   interface{}
}

// Run starts the assistant listening to the server and handles and interrupts or closed connections
func (assistant *Assistant) Run(ctx context.Context) error {
	var config = assistant.config
	var addr = flag.String("addr", config.Host, "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: config.Path}

	header := http.Header{}

	if v, exists := os.LookupEnv("INFLUX_TOKEN"); exists {
		header.Add("Authorization", "Token "+v)
	} else {
		return fmt.Errorf("influx authorization token not found, please set in env")
	}

	// creates a new websocket connection
	log.Printf("D! [assistant] Attempting connection to [%s]", config.Host)
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	for err != nil { // on error, retry connection again
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

	defer assistant.connection.Close()
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
			// TODO add error handling for different types of errors
			// common error that we see now is trying to read from a closed connection
			log.Printf("E! [assistant] error while reading from server: %s", err)
			// retry connection
		}
		var res response
		switch req.Operation {
		case GET_PLUGIN:
			res = assistant.getPlugin(req)
		case ADD_PLUGIN:
			// epic 2
			res = response{SUCCESS, req.UUID, fmt.Sprintf("%s plugin added.", req.Plugin.Name)}
		case UPDATE_PLUGIN:
			data := "TODO fetch plugin config"
			res = response{SUCCESS, req.UUID, data}
		case DELETE_PLUGIN:
			// epic 2
			res = response{SUCCESS, req.UUID, fmt.Sprintf("%s plugin deleted.", req.Plugin.Name)}
		case GET_RUNNING_PLUGINS:
			res, err = assistant.getRunningPlugins()
		case GET_ALL_PLUGINS:
			// epic 2
			data := "TODO fetch all available plugins"
			res = response{SUCCESS, req.UUID, data}
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

func (assistant *Assistant) getPlugin(req request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var res response
	var data interface{} // Pointer to the plugin.
	var err error

	switch req.Plugin.Type {
	case "INPUT":
		data, err = assistant.agent.GetInputPlugin(req.Plugin.Name)
	case "OUTPUT":
		data, err = assistant.agent.GetOutputPlugin(req.Plugin.Name)
	case "AGGREGATOR":
		data, err = assistant.agent.GetAggregatorPlugin(req.Plugin.Name)
	case "PROCESSOR":
		data, err = assistant.agent.GetProcessorPlugin(req.Plugin.Name)
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

func (assistant *Assistant) updatePlugin(req request) response {
	fmt.Print("D! [assistant] Received request: ", req.Operation, " for plugin ", req.Plugin.Name, "\n")

	var res response
	var data interface{} // Pointer to the plugin.
	var err error

	if req.Plugin.Config == nil {
		res = response{FAILURE, req.UUID, "no config specified!"}
		return res
	}
	fmt.Printf("Starting checking type of plugin\n")

	switch req.Plugin.Type {
	case "INPUT":
		data, err = assistant.agent.UpdateInputPlugin(req.Plugin.Name, req.Plugin.Config)
	case "OUTPUT":
		data, err = assistant.agent.UpdateOutputPlugin(req.Plugin.Name, req.Plugin.Config)
	case "AGGREGATOR":
		// // TODO
		// data, err = assistant.agent.UpdateAggregatorPlugin(req.Plugin.Name, &req.Plugin.Config)
	case "PROCESSOR":
		// // TODO
		// data, err = assistant.agent.UpdateProcessorPlugin(req.Plugin.Name, &req.Plugin.Config)
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

// TODO Implement after merge
func (assistant *Assistant) getRunningPlugins() (response, error) {
	// inputs := assistant.agent.GetInputPlugins()
	// outputs := assistant.agent.GetOutputPlugins()
	// data := map[string][]string{
	// 	"outputs": outputs,
	// 	"inputs":  inputs,
	// }
	// res = response{SUCCESS, req.UUID, data}
	// return res
	return response{}, nil
}

// TODO Implement after merge
func (assistant *Assistant) getAllPlugins() (response, error) {
	// inputs := assistant.agent.GetAllInputPlugins()
	// outputs := assistant.agent.GetAllOutputPlugins()
	// data := map[string][]string{
	// 	"outputs": outputs,
	// 	"inputs":  inputs,
	// }
	// res = response{SUCCESS, req.UUID, data}
	return response{}, nil
}
