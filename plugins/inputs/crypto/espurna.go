package crypto

import (
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const espurnaName = "espurna"

var espurnaHTTPHeader = http.Header{"Accept": []string{"application/json"}}

// espurna firmware
// https://github.com/xoseperez/espurna/wiki/RESTAPI
// curl -s --connect-timeout 1 -m 1 -H "Accept: application/json" http://172.33.22.101/apis?apikey=3BCD588C7A371AAE | jq .
type espurna struct {
	serverBase
	APIKeys []string `toml:"apiKeys"`
}

var espurnaSampleConf = `
  interval = "1m"
  ## sensors' addresses and names
  servers  = ["rig1;localhost:3333"]
  names    = ["Rig1"]
  apiKeys  = ["3BCD588C7A371AAE"]
`

// Description of espurnaPow
func (*espurna) Description() string {
	return "Read espurna sensor's status"
}

// SampleConfig of espurnaPow
func (*espurna) SampleConfig() string {
	return espurnaSampleConf
}

func (e *espurna) getURL(i int, path string, reply interface{}) error {
	return getResponse("http://"+e.Servers[i]+path+"?apikey="+e.APIKeys[i], espurnaHTTPHeader, nil, &reply)
}

func (e *espurna) serverGather(acc telegraf.Accumulator, i int, tags map[string]string) error {
	apis := make(map[string]string)
	if err := e.getURL(i, "/apis", &apis); err != nil {
		return err
	}
	fields := make(map[string]interface{})
	for _, path := range apis {
		response := make(map[string]float64)
		if err := e.getURL(i, path, &response); err != nil {
			acc.AddError(err)
			continue
		}
		for key, value := range response {
			if strings.Contains(key, "/") {
				// name := strings.Split(key, "/")[0]
				name := strings.Replace(key, "/", "_", -1)
				fields[name] = value
			} else {
				fields[key] = value
			}
		}
	}
	acc.AddFields(espurnaName, fields, tags)
	return nil
}

// Gather of espurnaPow
func (e *espurna) Gather(acc telegraf.Accumulator) error {
	return e.minerGather(acc, e)
}

func init() {
	inputs.Add(espurnaName, func() telegraf.Input { return &espurna{} })
}
