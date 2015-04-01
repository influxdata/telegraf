package tivan

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/naoina/toml"
	"github.com/naoina/toml/ast"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalTOML(b []byte) error {
	dur, err := time.ParseDuration(string(b[1 : len(b)-1]))
	if err != nil {
		return err
	}

	d.Duration = dur

	return nil
}

type Config struct {
	URL       string
	Username  string
	Password  string
	Database  string
	UserAgent string
	Tags      map[string]string

	plugins map[string]*ast.Table
}

func (c *Config) Plugins() map[string]*ast.Table {
	return c.plugins
}

func (c *Config) Apply(name string, v interface{}) error {
	if tbl, ok := c.plugins[name]; ok {
		return toml.UnmarshalTable(tbl, v)
	}

	return nil
}

func DefaultConfig() *Config {
	return &Config{}
}

var ErrInvalidConfig = errors.New("invalid configuration")

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		return nil, err
	}

	c := &Config{
		plugins: make(map[string]*ast.Table),
	}

	for name, val := range tbl.Fields {
		subtbl, ok := val.(*ast.Table)
		if !ok {
			return nil, ErrInvalidConfig
		}

		if name == "influxdb" {
			err := toml.UnmarshalTable(subtbl, c)
			if err != nil {
				return nil, err
			}
		} else {
			c.plugins[name] = subtbl
		}
	}

	return c, nil
}

func (c *Config) ListTags() string {
	var tags []string

	for k, v := range c.Tags {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(tags)

	return strings.Join(tags, " ")
}
