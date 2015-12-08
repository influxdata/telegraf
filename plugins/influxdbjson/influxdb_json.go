package influxdbjson

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/influxdb/telegraf/plugins"
)

type InfluxDBJSON struct {
	Name string
	URLs []string `toml:"urls"`
}

func (*InfluxDBJSON) Description() string {
	return "Read InfluxDB-formatted JSON metrics from one or more HTTP endpoints"
}

func (*InfluxDBJSON) SampleConfig() string {
	return `
  # Reads InfluxDB-formatted JSON from given URLs. For example,
	# monitoring a URL which responded with a JSON object formatted like this:
	#
  #   {
  #     "(ignored_key)": {
  #       "name": "connections",
  #       "tags": {
  #         "host": "foo"
  #       },
  #       "values": {
  #         "avg_ms": 1.234,
  #       }
  #     }
  #   }
  #
	# with configuration of { name = "server", urls = ["http://127.0.0.1:8086/x"] }
  #
	# Would result in this recorded metric:
	#
  #   influxdbjson_server_connections,influxdbjson_url='http://127.0.0.1:8086/x',host='foo' avg_ms=1.234
  [[plugins.influxdbjson]]
	# Name to use for measurement
	name = "influxdb"

	# Multiple URLs from which to read InfluxDB-formatted JSON
	urls = [
		"http://localhost:8086/debug/vars"
	]
`
}

func (i *InfluxDBJSON) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	errorChannel := make(chan error, len(i.URLs))

	for _, u := range i.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := i.gatherURL(acc, url); err != nil {
				errorChannel <- err
			}
		}(u)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

type point struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (i *InfluxDBJSON) gatherURL(
	acc plugins.Accumulator,
	url string,
) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Can't predict what all is going to be in the response, so decode the top keys one at a time.
	dec := json.NewDecoder(resp.Body)

	// Parse beginning of object
	if t, err := dec.Token(); err != nil {
		return err
	} else if t != json.Delim('{') {
		return errors.New("document root must be a JSON object")
	}

	// Loop through rest of object
	for {
		// Nothing left in this object, we're done
		if !dec.More() {
			break
		}

		// Read in a string key. We don't do anything with the top-level keys, so it's discarded.
		_, err := dec.Token()
		if err != nil {
			return err
		}

		// Attempt to parse a whole object into a point.
		// It might be a non-object, like a string or array.
		// If we fail to decode it into a point, ignore it and move on.
		var p point
		if err := dec.Decode(&p); err != nil {
			continue
		}

		// If the object was a point, but was not fully initialized, ignore it and move on.
		if p.Name == "" || p.Tags == nil || p.Values == nil || len(p.Values) == 0 {
			continue
		}

		// Add a tag to indicate the source of the data.
		p.Tags["influxdbjson_url"] = url

		acc.AddFields(
			i.Name+"_"+p.Name,
			p.Values,
			p.Tags,
		)
	}

	return nil
}

func init() {
	plugins.Add("influxdbjson", func() plugins.Plugin {
		return &InfluxDBJSON{}
	})
}
