package containerapp

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal"
)

var conunt int

const (
	interval        = "10s"
	responseTimeout = "5s"
	delimiter       = "/"
)

type HTTPConfig struct {
	NameOverride    string
	IP              string
	Port            int
	Path            string
	Interval        string
	ResponseTimeout string
	Method          string
	Tags            map[string]string
	TagKeys         []string
	Parameters      map[string]string
	Headers         map[string]string
}

type HTTPGather struct {
	id             string
	server         *ContainerApp
	interval       internal.Duration
	cfg            *HTTPConfig
	httpjsonclient *HttpJson
	closeCh        chan bool
}

func getValue(
	name string,
	settings map[string]string,
	values map[string]string,
	defaults map[string]string,
) (string, bool) {
	if envkey, ok := settings[name]; ok {
		if val, ok := values[envkey]; ok {
			return val, true
		}
	}
	if val, ok := defaults[name]; ok {
		return val, true
	}
	return "", false
}

func CreateHTTPGatherConf(
	containerid string,
	settings map[string]string,
	defaults map[string]string,
	values map[string]string,
) (*HTTPConfig, error) {
	clientcfg := &HTTPConfig{}

	if val, ok := getValue("http_port", settings, values, defaults); ok {
		port, err := strconv.Atoi(val)
		if err == nil {
			clientcfg.Port = port
		} else {
			return nil, fmt.Errorf("skip object: %s", containerid)
		}
	}

	if val, ok := getValue("name_override", settings, values, defaults); ok {
		clientcfg.NameOverride = val
	}
	if val, ok := getValue("http_path", settings, values, defaults); ok {
		if pathDelimiter, ok := getValue("http_path_delimiter", settings, values, defaults); ok {
			clientcfg.Path = strings.Replace(val, pathDelimiter, delimiter, -1)
		} else {
			clientcfg.Path = val
		}
	}

	if val, ok := getValue("interval", settings, values, defaults); ok {
		clientcfg.Interval = val
	} else {
		clientcfg.Interval = interval
	}
	if val, ok := getValue("http_response_timeout", settings, values, defaults); ok {
		clientcfg.ResponseTimeout = val
	} else {
		clientcfg.ResponseTimeout = responseTimeout
	}
	if val, ok := getValue("http_method", settings, values, defaults); ok {
		clientcfg.Method = val
	}
	if val, ok := getValue("tag_keys_json", settings, values, defaults); ok {
		err := json.Unmarshal([]byte(val), &clientcfg.TagKeys)
		if err != nil {
			clientcfg.TagKeys = nil
		}

	}
	if val, ok := getValue("custom_tags", settings, values, defaults); ok {
		err := json.Unmarshal([]byte(val), &clientcfg.Tags)
		if err != nil {
			clientcfg.Tags = nil
		}
	}
	if val, ok := getValue("http_parameters", settings, values, defaults); ok {
		err := json.Unmarshal([]byte(val), &clientcfg.Parameters)
		if err != nil {
			clientcfg.Parameters = nil
		}
	}
	if val, ok := getValue("http_headers", settings, values, defaults); ok {
		err := json.Unmarshal([]byte(val), &clientcfg.Headers)
		if err != nil {
			clientcfg.Headers = nil
		}
	}
	return clientcfg, nil
}

func formatURL(cfg *HTTPConfig) string {
	cfg.Path = strings.TrimSpace(cfg.Path)
	if len(cfg.Path) > 0 && cfg.Path[0] == '/' {
		cfg.Path = cfg.Path[1:]
	}
	var serverURL string
	if len(cfg.Path) > 0 {
		serverURL = fmt.Sprintf("http://%s:%d/%s", cfg.IP, cfg.Port, cfg.Path)
	} else {
		serverURL = fmt.Sprintf("http://%s:%d", cfg.IP, cfg.Port)

	}
	return serverURL
}

func formatDuration(id string, name string, duration string) (internal.Duration, error) {
	d := internal.Duration{}
	err := d.UnmarshalTOML([]byte(duration))
	if err != nil {
		err := fmt.Errorf("Container: %s, wrong %s: %s", id, name, err.Error())
		return d, err
	}
	return d, nil
}

func NewHTTPGather(server *ContainerApp, id string, cfg *HTTPConfig) (*HTTPGather, error) {
	closeCh := make(chan bool)
	httpjsonclient := NewHttpJson()

	clean := func() {
		close(closeCh)
		httpjsonclient = nil
		cfg.Tags = nil
		cfg.TagKeys = nil
		cfg.Parameters = nil
		cfg.Headers = nil
		cfg = nil
	}

	if len(cfg.IP) > 0 && cfg.Port > 0 {
		httpjsonclient.Servers = []string{formatURL(cfg)}
	} else {
		err := fmt.Errorf("Container: %s, IP or Port ivalide, values: %s %d", id, cfg.IP, cfg.Port)
		clean()
		return nil, err
	}

	interval, err := formatDuration(id, "Interval", cfg.Interval)
	if err != nil {
		clean()
		return nil, err
	}

	if len(cfg.Path) > 0 {
		httpjsonclient.Name = cfg.NameOverride
	}
	if len(cfg.Method) > 0 {
		httpjsonclient.Method = cfg.Method
	}
	if len(cfg.TagKeys) > 0 {
		httpjsonclient.TagKeys = cfg.TagKeys
	}
	if len(cfg.Parameters) > 0 {
		httpjsonclient.Parameters = cfg.Parameters
	}
	if len(cfg.Headers) > 0 {
		httpjsonclient.Headers = cfg.Headers
	}
	if len(cfg.ResponseTimeout) > 0 {
		d, err := formatDuration(id, "ResponseTimeout", cfg.ResponseTimeout)
		if err != nil {
			clean()
			return nil, err
		}
		httpjsonclient.ResponseTimeout = d
	}

	return &HTTPGather{id, server, interval, cfg, httpjsonclient, closeCh}, nil
}

func (c *HTTPGather) clean() {
	close(c.closeCh)
	c.httpjsonclient = nil
	c.cfg.Tags = nil
	c.cfg.TagKeys = nil
	c.cfg.Parameters = nil
	c.cfg.Headers = nil
	c.cfg = nil

}

func (c *HTTPGather) Close() {
	c.closeCh <- true
}

func (c *HTTPGather) Run() {
	go func() {
		log.Println("Run client:", c.id)
		ticker := time.NewTicker(c.interval.Duration)
		acc := NewAccumulator(c.id, c.server.metricsCh, c.server.errCh)

		for {
			go func() {
				if c.httpjsonclient != nil {
					err := c.httpjsonclient.Gather(acc)
					if err != nil {
						c.server.errCh <- err
					}
				}
			}()
			select {
			case <-c.closeCh:
				ticker.Stop()
				c.clean()
				return
			case <-ticker.C:
				continue
			}
		}
	}()
}
