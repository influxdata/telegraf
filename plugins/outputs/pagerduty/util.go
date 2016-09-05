package pagerduty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"net/http"
)

const EventEndPoint = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"

type Event struct {
	Type        string        `json:"event_type"`
	ServiceKey  string        `json:"service_key"`
	Description string        `json:"description,omitempty"`
	Client      string        `json:"client,omitempty"`
	ClientURL   string        `json:"client_url,omitempty"`
	Details     interface{}   `json:"details,omitempty"`
	Contexts    []interface{} `json:"contexts,omitempty"`
	IncidentKey string        `json:"incident_key"`
}

type Response struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	IncidentKey string `json:"incident_key"`
}

func createEvent(e Event) (*Response, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", EventEndPoint, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status Code: %d", resp.StatusCode)
	}
	var r Response
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func evalBoolExpr(expr string) (bool, error) {
	fs := token.NewFileSet()
	tv, err := types.Eval(fs, nil, token.NoPos, expr)
	if err != nil {
		return false, err
	}
	return tv.Value.String() == "true", nil
}
