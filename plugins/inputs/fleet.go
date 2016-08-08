/*
* @Author: Jim Weber
* @Date:   2016-05-18 22:07:31
* @Last Modified by:   Jim Weber
* @Last Modified time: 2016-08-07 20:20:26
 */

package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// FleetStates struct to hold all the data for a machine state
type FleetStates struct {
	States []struct {
		SystemdActiveState string `json:"systemdActiveState"`
		MachineID          string `json:"machineID"`
		Hash               string `json:"hash"`
		SystemdSubState    string `json:"systemdSubState"`
		Name               string `json:"name"`
		SystemdLoadState   string `json:"systemdLoadState"`
	}
}

// Fleet struct to hold fleet hosts
type Fleet struct {
	Hosts []string `toml:"hosts"`
}

// Description - Method to provide description of plugin
func (f *Fleet) Description() string {
	return "Fleetd Plugin to glather information about container states in fleet cluster"
}

// SampleConfig output sample config for this plugin
func (*Fleet) SampleConfig() string {
	return `
# Description
[[inputs.fleet]]
## Works with Fleet HTTP API
## Multiple Hosts from which to read Fleet stats:
	host = ["http://localhost:49153/fleet/v1/state"]
`
}

// Gather method to gather stats for telegraf input
func (f *Fleet) Gather(accumulator telegraf.Accumulator) error {
	errorChannel := make(chan error, len(f.Hosts))
	var wg sync.WaitGroup
	for _, u := range f.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err := f.fetchAndReturnData(accumulator, host); err != nil {
				errorChannel <- fmt.Errorf("[host=%s]: %s", host, err)
			}
		}(u)
	}

	wg.Wait()
	close(errorChannel)

	// If there weren't any errors, we can return nil now.
	if len(errorChannel) == 0 {
		return nil
	}

	// There were errors, so join them all together as one big error.
	errorStrings := make([]string, 0, len(errorChannel))
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	return errors.New(strings.Join(errorStrings, "\n"))
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func (f *Fleet) fetchAndReturnData(accumulator telegraf.Accumulator, host string) error {
	_, error := client.Get(host)
	if error != nil {
		return error
	}

	fleetStates := getInstanceStates(host, nil)
	containerCounts := getContainerCount(fleetStates)
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for k, v := range containerCounts {
		fields[k] = v
	}

	// create tags for each host if needed
	tags["server"] = host

	accumulator.AddFields("fleet", fields, tags)
	return nil
}

func getInstanceStates(host string, params map[string]string) FleetStates {

	response, err := http.Get(host)
	fleetStates := FleetStates{}

	if err != nil {
		fmt.Printf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(contents, &fleetStates); err != nil {
			panic(err)
		}

	}

	return fleetStates
}

func getContainerCount(fleetUnits FleetStates) map[string]int {
	containerCount := make(map[string]int)
	for _, fleetUnit := range fleetUnits.States {
		shortNameParts := strings.Split(fleetUnit.Name, "@")
		shortName := shortNameParts[0]
		if fleetUnit.SystemdSubState == "running" {
			containerCount[shortName]++
		}
	}

	return containerCount
}

func init() {
	inputs.Add("fleet", func() telegraf.Input {
		return &Fleet{}
	})
}
