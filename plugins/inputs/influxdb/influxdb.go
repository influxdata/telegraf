package influxdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type InfluxDB struct {
	URLs []string `toml:"urls"`
}

func (*InfluxDB) Description() string {
	return "Read InfluxDB-formatted JSON metrics from one or more HTTP endpoints"
}

func (*InfluxDB) SampleConfig() string {
	return `
  ## Works with InfluxDB debug endpoints out of the box,
  ## but other services can use this format too.
  ## See the influxdb plugin's README for more details.

  ## Multiple URLs from which to read InfluxDB-formatted JSON
  urls = [
    "http://localhost:8086/debug/vars"
  ]
`
}

func (i *InfluxDB) Gather(acc telegraf.Accumulator) error {
	errorChannel := make(chan error, len(i.URLs))

	var wg sync.WaitGroup
	for _, u := range i.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := i.gatherURL(acc, url); err != nil {
				errorChannel <- fmt.Errorf("[url=%s]: %s", url, err)
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

type point struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telegraf Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (i *InfluxDB) gatherURL(
	acc telegraf.Accumulator,
	url string,
) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// It would be nice to be able to decode into a map[string]point, but
	// we'll get a decoder error like:
	// `json: cannot unmarshal array into Go value of type influxdb.point`
	// if any of the values aren't objects.
	// To avoid that error, we decode by hand.
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
		p.Tags["url"] = url

		acc.AddFields(
			"influxdb_"+p.Name,
			p.Values,
			p.Tags,
		)
	}

	return nil
}

func init() {
	inputs.Add("influxdb", func() telegraf.Input {
		return &InfluxDB{}
	})
}
