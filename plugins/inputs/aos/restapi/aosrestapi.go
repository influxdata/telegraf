/*
The MIT License (MIT)

Copyright 2014-present, Apstra, Inc. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package aosrestapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

type AosToken struct {
	Token string
}

type aosSystemList struct {
	Items []aosSystem `json:"items"`
}

type aosBlueprintList struct {
	Items []aosBlueprint `json:"items"`
}

type aosSystem struct {
	DeviceKey  string              `json:"device_key"`
	Facts      aosSystemFacts      `json:"facts"`
	Id         string              `json:"id"`
	Status     aosSystemStatus     `json:"status"`
	UserConfig aosSystemUserConfig `json:"user_config"`
	Blueprint  aosBlueprintSystemNode
}

type aosSystemFacts struct {
	AosHclModel  string `json:"aos_hcl_model"`
	AosServer    string `json:"aos_server"`
	AosVersion   string `json:"aos_version"`
	HwModel      string `json:"hw_model"`
	HwVersion    string `json:"hw_version"`
	MgmtIfName   string `json:"mgmt_ifname"`
	MgmtIpAddr   string `json:"mgmt_ipaddr"`
	MgmtMacAddr  string `json:"mgmt_macaddr"`
	OsArch       string `json:"os_arch"`
	OsFamily     string `json:"os_family"`
	OsVersion    string `json:"os_version"`
	SerialNumber string `json:"serial_number"`
	Vendor       string `json:"vendor"`
}

type aosSystemStatus struct {
	AgentStartTime  string `json:"agent_start_time"`
	CommState       string `json:"comm_state"`
	BlueprintActive bool   `json:"blueprint_active"`
	BlueprintId     string `json:"blueprint_id"`
	DeviceStartTime string `json:"device_start_time"`
	DomainName      string `json:"domain_name"`
	ErrorMessage    string `json:"error_message"`
	Fqdn            string `json:"fqdn"`
	Hostname        string `json:"hostname"`
	IsAcknowledged  bool   `json:"is_acknowledged"`
	PoolId          string `json:"pool_id"`
	State           string `json:"state"`
}

type aosSystemUserConfig struct {
	AdminState  string `json:"admin_state"`
	AosHclModel string `json:"aos_hcl_model"`
	Location    string `json:"location"`
}

type aosBlueprint struct {
	Id                    string `json:"id"`
	ReferenceArchitecture string `json:"reference_architecture"`
	Name                  string `json:"label"`
	Systems               map[string]aosBlueprintSystemNode
}

type aosVersion struct {
	Major   string `json:"major"`
	Minor   string `json:"minor"`
	Build   string `json:"build"`
	Version string `json:"version"`
}

// --------------------------------------------------------------------------
// Datastructure return by Query Engine Query for System Node in Blueprint
// --------------------------------------------------------------------------
type aosBlueprintSystem struct {
	Nodes []aosBlueprintSystemNode
}

type aosBlueprintSystemNodeList struct {
	Items []aosBlueprintSystemNodeItem `json:"items"`
}

type aosBlueprintSystemNodeItem struct {
	System aosBlueprintSystemNode `json:"system_node"`
}

type aosBlueprintSystemNode struct {
	Hostname   string `json:"hostname"`
	Name       string `json:"label"`
	Id         string `json:"id"`
	Role       string `json:"role"`
	DeployMode string `json:"deploy_mode"`
	Position   int    `json:"position"`
	SystemId   string `json:"system_id"`
	Type       string `json:"system_type"`
}

type AosServerApi struct {
	Address  string
	Port     int
	User     string
	Password string
	Protocol string

	Token      string
	AosVersion aosVersion

	sync.RWMutex      // following fields are protected by this lock
	Blueprints        map[string]aosBlueprint
	Systems           map[string]aosSystem
	StreamingSessions []string
}

type apiResponseId struct {
	Id string
}

func NewAosServerApi(address string, port int, user string, password string, protocol string) *AosServerApi {

	//TODO add check for protocol can only be http or https
	api := AosServerApi{Address: address, Port: port, User: user, Password: password, Protocol: protocol}

	// initialize Maps
	api.Blueprints = make(map[string]aosBlueprint, 20)
	api.Systems = make(map[string]aosSystem, 1000)

	return &api
}

func (api *AosServerApi) Login() (err error) {

	url := fmt.Sprintf("%v://%v:%v/api/user/login", api.Protocol, api.Address, api.Port)

	var jsonStr = []byte(fmt.Sprintf(`{ "username": "%v", "password": "%v" }`, api.User, api.Password))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("Status Code is not %v got %v", 201, resp.Status))
	}

	token := AosToken{}
	json.NewDecoder(resp.Body).Decode(&token)

	api.Token = token.Token

	return nil
}

func (api *AosServerApi) httpRequest(httpType string, address string, payload []byte, respData interface{}, expectedCode int) error {

	url := fmt.Sprintf("%v://%v:%v/api/%v", api.Protocol, api.Address, api.Port, address)

	req, err := http.NewRequest(httpType, url, bytes.NewBuffer(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AUTHTOKEN", api.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if Return code is what we are expecting
	if resp.StatusCode != expectedCode {
		return errors.New(fmt.Sprintf("Status Code is not %v got %v", expectedCode, resp.Status))
	}

	// Read Body and Unmarshal JSON
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, respData)
	if err != nil {
		return err
	}

	return nil
}

func (api *AosServerApi) StopStreaming() error {

	for _, id := range api.StreamingSessions {

		url := fmt.Sprintf("%v://%v:%v/api/streaming-config/%v", api.Protocol, api.Address, api.Port, id)

		req, err := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("AUTHTOKEN", api.Token)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 202 {
			return errors.New(resp.Status)
		}
	}

	return nil
}

func (api *AosServerApi) StartStreaming(streamingType string, address string, port uint16, sequencing bool) error {

	url := fmt.Sprintf("%v://%v:%v/api/streaming-config", api.Protocol, api.Address, api.Port)

	var jsonStr []byte

	if sequencing {
		jsonStr = []byte(fmt.Sprintf(`{
                    "streaming_type": "%v",
                    "hostname": "%v",
                    "protocol": "protoBufOverTcp",
                    "port": %v,
                    "sequencing_mode": "sequenced"
                }`, streamingType, address, port))
	} else {
		jsonStr = []byte(fmt.Sprintf(`{
                    "streaming_type": "%v",
                    "hostname": "%v",
                    "protocol": "protoBufOverTcp",
                    "port": %v
                }`, streamingType, address, port))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AUTHTOKEN", api.Token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		return errors.New(fmt.Sprintf("%v : Endpoint already exist", resp.Status))
	} else if resp.StatusCode != 201 {
		panic(resp.Status)
	}

	stramingResp := apiResponseId{}

	json.NewDecoder(resp.Body).Decode(&stramingResp)

	api.Lock()
	defer api.Unlock()

	api.StreamingSessions = append(api.StreamingSessions, stramingResp.Id)

	return nil
}

func (api *AosServerApi) GetBlueprints() error {

	blueprintList := aosBlueprintList{}
	err := api.httpRequest("GET", "blueprints", nil, &blueprintList, 200)
	if err != nil {
		return err
	}

	api.Lock()
	defer api.Unlock()

	// Save all items in the API Object by ID
	for i := 0; i < len(blueprintList.Items); i++ {

		id := blueprintList.Items[i].Id
		fmt.Printf("GetBlueprints() - Id %v \n", id)

		// Create Blueprint from JSON value returned by API
		// Initialize list of map of Systems to be queried later by Query Engine
		//   Not possible to assign a field inside a struct directly in if part of a map
		//   Must save same stuct temporarely
		b := blueprintList.Items[i]
		b.Systems = make(map[string]aosBlueprintSystemNode, 1000)
		api.Blueprints[id] = b

		// Get list of system in the blueprint with Separate Query
		tmp, systemErr := api.GetSystemsInBlueprint(id)
		if systemErr != nil {
			fmt.Printf("Issue while trying to GetSystemsInBlueprint  %s\n", systemErr)
			continue
		}

		for y := 0; y < len(tmp.Items); y++ {
			systemId := tmp.Items[y].System.Name
			api.Blueprints[id].Systems[systemId] = tmp.Items[y].System
		}
	}

	return nil
}

func (api *AosServerApi) GetSystemsInBlueprint(blueprintId string) (*aosBlueprintSystemNodeList, error) {

	systemBpList := aosBlueprintSystemNodeList{}
	var jsonStr = []byte(fmt.Sprintf(`{ "query": "match(node('system', name='system_node'))" }`))

	err := api.httpRequest("POST", "blueprints/"+blueprintId+"/qe", jsonStr, &systemBpList, 200)
	if err != nil {
		return nil, err
	}

	return &systemBpList, nil
}

func (api *AosServerApi) GetSystems() error {

	systemList := aosSystemList{}

	err := api.httpRequest("GET", "systems", nil, &systemList, 200)
	if err != nil {
		return err
	}

	api.Lock()
	defer api.Unlock()

	for _, system := range systemList.Items {

		id := system.Id

		api.Systems[id] = system
		s := api.Systems[id]

		// If Blueprint is defined, search for the System.Node information and copy it inside system
		if system.Status.BlueprintId != "" {

			if bp, ok := api.Blueprints[system.Status.BlueprintId]; ok {
				var found bool

				for _, node := range bp.Systems {
					if node.SystemId == id {
						s.Blueprint = node
						api.Systems[id] = s
						found = true
						break
					}
				}
				if found == false {
					return errors.New(fmt.Sprintf(
						"System %v has Blueprint ID defined (%v) but was not able to find it in System.Nodes",
						id, system.Status.BlueprintId))
				}
			} else {
				return errors.New(fmt.Sprintf(
					"System %v has Blueprint ID defined (%v) but no blueprint with this id exist in map",
					id, system.Status.BlueprintId))
			}
		}
	}
	return nil
}

func (api *AosServerApi) GetSystemByKey(deviceKey string) *aosSystem {

	api.RLock()
	defer api.RUnlock()

	system, ok := api.Systems[deviceKey]
	if ok {
		return &system
	}

	return nil
}

func (api *AosServerApi) GetBlueprintById(blueprintId string) *aosBlueprint {

	api.RLock()
	defer api.RUnlock()

	blueprint, ok := api.Blueprints[blueprintId]
	if ok {
		return &blueprint
	}

	return nil
}

func (api *AosServerApi) GetVersion() error {
	aosVersion := aosVersion{}
	err := api.httpRequest("GET", "version", nil, &aosVersion, 200)
	if err != nil {
		return err
	}
	api.AosVersion = aosVersion
	return nil
}
