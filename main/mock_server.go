package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error: ", err)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	defer c.Close()
	for {
		fmt.Println("\n(0) GET_PLUGIN")
		fmt.Println("(1) START_PLUGIN")
		fmt.Println("(2) STOP_PLUGIN")
		fmt.Println("(3) UPDATE_PLUGIN")
		fmt.Println("(4) GET_RUNNING_PLUGINS")
		fmt.Println("(5) GET_ALL_PLUGINS")
		fmt.Println("(6) GET_PLUGIN_SCHEMA")
		fmt.Print("\nOperation: ")
		operation, _ := reader.ReadString('\n')
		fmt.Print("Plugin name: ")
		plugin, _ := reader.ReadString('\n')
		fmt.Print("Plugin type: ")
		pluginType, _ := reader.ReadString('\n')
		fmt.Print("Plugin config: ")
		pluginConfig, _ := reader.ReadString('\n')
		plugin = strings.Replace(plugin, "\n", "", -1)
		pluginType = strings.Replace(pluginType, "\n", "", -1)
		operation = strings.Replace(operation, "\n", "", -1)
		pluginConfig = strings.Replace(pluginConfig, "\n", "", -1)

		switch operation {
		case "0":
			operation = "GET_PLUGIN"
		case "1":
			operation = "START_PLUGIN"
		case "2":
			operation = "STOP_PLUGIN"
		case "3":
			operation = "UPDATE_PLUGIN"
		case "4":
			operation = "GET_RUNNING_PLUGINS"
		case "5":
			operation = "GET_ALL_PLUGINS"
		case "6":
			operation = "GET_PLUGIN_SCHEMA"
		default:
			operation = ""
		}

		uid, _ := uuid.NewRandom()
		var config map[string]interface{}
		_ = json.Unmarshal([]byte(pluginConfig), &config)
		var m = map[string]interface{}{
			"Operation": operation,
			"Uuid":      uid.String(),
			"Plugin": map[string]interface{}{
				"Name":   plugin,
				"Type":   pluginType,
				"Config": config,
			},
		}
		err = c.WriteJSON(m)
		if err != nil {
			log.Println("write:", err)
			break
		}
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}

func convertStringToJSON(s string) ([]byte, error) {
	req := &Request{}
	json.Unmarshal([]byte(s), req)
	return json.Marshal(req)
}

type Request struct {
	Operation string
	Uuid      string
	Plugin    map[string]string
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	log.Println("listening on localhost:8080/echo...")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
