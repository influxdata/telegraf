package httpjson

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/influxdb/telegraf/plugins"
	"net/http"
	"sync"
)

type HttpJson struct {
	Servers      []string
	Measurements map[string]string
	Method       string
	Foo          string
	client       *http.Client
}

var sampleConfig = `
# stats url endpoint
servers = ["http://localhost:5000"]

# a name for server(s)
foo = "mycluster"

# HTTP method (GET or POST)
method = "GET"

# Map of key transforms # TODO describe
[httpjson.measurements]
stats_measurements_measurement = "my_measurement"
`

func (h *HttpJson) SampleConfig() string {
	return sampleConfig
}

func (h *HttpJson) Description() string {
	return "Read flattened metrics from one or more JSON HTTP endpoints"
}

func (h *HttpJson) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	var outerr error

	for _, server := range h.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			outerr = h.gatherServer(server, acc)
		}(server)
	}

	wg.Wait()

	return outerr
}

func (h *HttpJson) gatherServer(url string, acc plugins.Accumulator) error {
	r, err := h.client.Get(url)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("httpjson: server '%s' responded with status-code %d, expected %d", r.StatusCode, http.StatusOK)
	}

	response, err := simplejson.NewFromReader(r.Body)

	if err != nil {
		return err
	}

	tags := map[string]string{
		"server": url,
	}

	return parseResponse(acc, h.Foo, tags, response.Interface(), h.Measurements)
}

func parseResponse(acc plugins.Accumulator, prefix string, tags map[string]string, v interface{}, measurements map[string]string) error {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			if err := parseResponse(acc, prefix+"_"+k, tags, v, measurements); err != nil {
				return err
			}
		}
	case json.Number:
		if transform, ok := measurements[prefix]; ok {
			prefix = transform
		}
		acc.Add(prefix, t, tags)
	case bool, string, []interface{}:
		// ignored types
		return nil
	default:
		return fmt.Errorf("httpjson: got unexpected type %T with value %v (%s)", t, v, prefix)
	}
	return nil
}

func init() {
	plugins.Add("httpjson", func() plugins.Plugin {
		return &HttpJson{client: http.DefaultClient}
	})
}
