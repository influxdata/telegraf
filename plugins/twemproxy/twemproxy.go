package twemproxy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdb/telegraf/plugins"
)

type Twemproxy struct {
	Instances []TwemproxyInstance
}

type TwemproxyInstance struct {
	Addr  string
	Pools []string
}

var sampleConfig = `
  [[twemproxy.instances]]
    # Twemproxy stats address and port (no scheme)
    addr = "10.16.29.1:22222"
    # Monitor pool name
    pools = ["redis_pool", "mc_pool"]
`

func (t *Twemproxy) SampleConfig() string {
	return sampleConfig
}

func (t *Twemproxy) Description() string {
	return "Read Twemproxy stats data"
}

// Gather data from all Twemproxy instances
func (t *Twemproxy) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(t.Instances))
	for _, inst := range t.Instances {
		wg.Add(1)
		go func(inst TwemproxyInstance) {
			defer wg.Done()
			if err := inst.Gather(acc); err != nil {
				errorChan <- err
			}
		}(inst)
	}
	wg.Wait()

	close(errorChan)
	errs := []string{}
	for err := range errorChan {
		errs = append(errs, err.Error())
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}

// Gather data from one Twemproxy
func (ti *TwemproxyInstance) Gather(
	acc plugins.Accumulator,
) error {
	conn, err := net.DialTimeout("tcp", ti.Addr, 1*time.Second)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}

	var stats map[string]interface{}
	if err = json.Unmarshal(body, &stats); err != nil {
		return errors.New("Error decoding JSON response")
	}

	tags := make(map[string]string)
	tags["twemproxy"] = ti.Addr
	ti.processStat(acc, tags, stats)

	return nil
}

// Process Twemproxy server stats
func (ti *TwemproxyInstance) processStat(
	acc plugins.Accumulator,
	tags map[string]string,
	data map[string]interface{},
) {
	if source, ok := data["source"]; ok {
		if val, ok := source.(string); ok {
			tags["source"] = val
		}
	}

	metrics := []string{"total_connections", "curr_connections", "timestamp"}
	for _, m := range metrics {
		if value, ok := data[m]; ok {
			if val, ok := value.(float64); ok {
				acc.Add(m, val, tags)
			}
		}
	}

	for _, pool := range ti.Pools {
		if poolStat, ok := data[pool]; ok {
			if data, ok := poolStat.(map[string]interface{}); ok {
				poolTags := copyTags(tags)
				poolTags["pool"] = pool
				ti.processPool(acc, poolTags, pool+"_", data)
			}
		}
	}
}

// Process pool data in Twemproxy stats
func (ti *TwemproxyInstance) processPool(
	acc plugins.Accumulator,
	tags map[string]string,
	prefix string,
	data map[string]interface{},
) {
	serverTags := make(map[string]map[string]string)

	for key, value := range data {
		switch key {
		case "client_connections", "forward_error", "client_err", "server_ejects", "fragments", "client_eof":
			if val, ok := value.(float64); ok {
				acc.Add(prefix+key, val, tags)
			}
		default:
			if data, ok := value.(map[string]interface{}); ok {
				if _, ok := serverTags[key]; !ok {
					serverTags[key] = copyTags(tags)
					serverTags[key]["server"] = key
				}
				ti.processServer(acc, serverTags[key], prefix, data)
			}
		}
	}
}

// Process backend server(redis/memcached) stats
func (ti *TwemproxyInstance) processServer(
	acc plugins.Accumulator,
	tags map[string]string,
	prefix string,
	data map[string]interface{},
) {
	for key, value := range data {
		switch key {
		default:
			if val, ok := value.(float64); ok {
				acc.Add(prefix+key, val, tags)
			}
		}
	}
}

// Tags is not expected to be mutated after passing to Add.
func copyTags(tags map[string]string) map[string]string {
	newTags := make(map[string]string)
	for k, v := range tags {
		newTags[k] = v
	}
	return newTags
}

func init() {
	plugins.Add("twemproxy", func() plugins.Plugin {
		return &Twemproxy{}
	})
}
