package netmonk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	verifyEndpoint = "/public/controller/server/%s/verify"
)

// Agent common to authenticate netmonk agent.
type Agent struct {
	NetmonkHost      string `toml:"netmonk_host"`
	NetmonkServerID  string `toml:"netmonk_server_id"`
	NetmonkServerKey string `toml:"netmonk_server_key"`
}

type CustomerCredentials struct {
	ClientID      string        `json:"client_id"`
	MessageBroker MessageBroker `json:"message_broker"`
	Auth          Auth          `json:"auth"`
	SASL          SASL          `json:"sasl"`
	TLS           TLS           `json:"tls"`
}

type MessageBroker struct {
	Type      string   `json:"type"`
	Addresses []string `json:"address"`
}

type Auth struct {
	IsEnabled bool   `json:"is_enabled"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type SASL struct {
	IsEnabled bool   `json:"is_enabled"`
	Mechanism string `json:"mechanism"`
}

type TLS struct {
	IsEnabled bool   `json:"is_enabled"`
	CA        string `json:"ca"`
	Access    string `json:"access"`
	Key       string `json:"key"`
}

func NewAgent(host, serverid, serverkey string) *Agent {
	return &Agent{
		NetmonkHost:      host,
		NetmonkServerID:  serverid,
		NetmonkServerKey: serverkey,
	}
}

// Verify netmonk agent.
func (n *Agent) Verify() (*CustomerCredentials, error) {
	postBody, _ := json.Marshal(map[string]string{
		"key": n.NetmonkServerKey,
	})
	reqBody := bytes.NewBuffer(postBody)

	endpoint := fmt.Sprintf(verifyEndpoint, n.NetmonkServerID)
	resp, err := http.Post(fmt.Sprintf("%s%s", n.NetmonkHost, endpoint), "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		cc := CustomerCredentials{}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		err = json.Unmarshal(body, &cc)
		if err != nil {
			return nil, err
		}

		return &cc, nil
	}

	return nil, fmt.Errorf("failed to verify agent")
}
