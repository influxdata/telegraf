package expvar

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/influxdb/telegraf/plugins"
)

var sampleConfig = `
  # Specify services via an array of tables
  [[plugins.expvar.services]]
    # Name for the service being polled
    name = "influxdb"

    # Multiple URLs from which to read expvars
    urls = [
      "http://localhost:8086/debug/vars"
    ]
`

type Expvar struct {
	Services []Service
}

type Service struct {
	Name string
	URLs []string `toml:"urls"`
}

func (*Expvar) Description() string {
	return "Read InfluxDB-style expvar metrics from one or more HTTP endpoints"
}

func (*Expvar) SampleConfig() string {
	return sampleConfig
}

func (e *Expvar) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	totalURLs := 0
	for _, service := range e.Services {
		totalURLs += len(service.URLs)
	}
	errorChannel := make(chan error, totalURLs)

	for _, service := range e.Services {
		for _, u := range service.URLs {
			wg.Add(1)
			go func(service Service, url string) {
				defer wg.Done()
				if err := e.gatherURL(acc, service, url); err != nil {
					errorChannel <- err
				}
			}(service, u)
		}
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
//     service: the service being queried
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (e *Expvar) gatherURL(
	acc plugins.Accumulator,
	service Service,
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
		return errors.New("expvars must be a JSON object")
	}

	// Loop through rest of object
	for {
		// Nothing left in this object, we're done
		if !dec.More() {
			break
		}

		// Read in a string key. We actually don't care about the top-level keys
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

		if p.Name == "" || p.Tags == nil || p.Values == nil || len(p.Values) == 0 {
			continue
		}

		p.Tags["expvar_url"] = url

		acc.AddFields(
			service.Name+"_"+p.Name,
			p.Values,
			p.Tags,
		)
	}

	return nil
}

func init() {
	plugins.Add("expvar", func() plugins.Plugin {
		return &Expvar{}
	})
}
