package collectd

import (
	"errors"
	"fmt"
	"os"

	"collectd.org/api"
	"collectd.org/network"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	DefaultAuthFile = "/etc/collectd/auth_file"
)

type Parser struct {
	DefaultTags map[string]string `toml:"-"`

	//whether or not to split multi value metric into multiple metrics
	//default value is split
	ParseMultiValue string `toml:"collectd_parse_multivalue"`

	popts         network.ParseOpts
	AuthFile      string   `toml:"collectd_auth_file"`
	SecurityLevel string   `toml:"collectd_security_level"`
	TypesDB       []string `toml:"collectd_typesdb"`

	Log telegraf.Logger `toml:"-"`
}

func (p *Parser) Init() error {
	switch p.SecurityLevel {
	case "none":
		p.popts.SecurityLevel = network.None
	case "sign":
		p.popts.SecurityLevel = network.Sign
	case "encrypt":
		p.popts.SecurityLevel = network.Encrypt
	default:
		p.popts.SecurityLevel = network.None
	}

	if p.AuthFile == "" {
		p.AuthFile = DefaultAuthFile
	}
	p.popts.PasswordLookup = network.NewAuthFile(p.AuthFile)

	for _, path := range p.TypesDB {
		db, err := LoadTypesDB(path)
		if err != nil {
			return err
		}

		if p.popts.TypesDB != nil {
			p.popts.TypesDB.Merge(db)
		} else {
			p.popts.TypesDB = db
		}
	}

	return nil
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	valueLists, err := network.Parse(buf, p.popts)
	if err != nil {
		return nil, fmt.Errorf("collectd parser error: %s", err)
	}

	metrics := []telegraf.Metric{}
	for _, valueList := range valueLists {
		metrics = append(metrics, p.unmarshalValueList(valueList)...)
	}

	if len(p.DefaultTags) > 0 {
		for _, m := range metrics {
			for k, v := range p.DefaultTags {
				// only set the default tag if it doesn't already exist:
				if !m.HasTag(k) {
					m.AddTag(k, v)
				}
			}
		}
	}

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) != 1 {
		return nil, errors.New("line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

// unmarshalValueList translates a ValueList into a Telegraf metric.
func (p *Parser) unmarshalValueList(vl *api.ValueList) []telegraf.Metric {
	timestamp := vl.Time.UTC()

	var metrics []telegraf.Metric

	var multiValue = p.ParseMultiValue
	//set multiValue to default "split" if nothing is specified
	if multiValue == "" {
		multiValue = "split"
	}
	switch multiValue {
	case "split":
		for i := range vl.Values {
			name := fmt.Sprintf("%s_%s", vl.Identifier.Plugin, vl.DSName(i))
			tags := make(map[string]string)
			fields := make(map[string]interface{})

			// Convert interface back to actual type, then to float64
			switch value := vl.Values[i].(type) {
			case api.Gauge:
				fields["value"] = float64(value)
			case api.Derive:
				fields["value"] = float64(value)
			case api.Counter:
				fields["value"] = float64(value)
			}

			if vl.Identifier.Host != "" {
				tags["host"] = vl.Identifier.Host
			}
			if vl.Identifier.PluginInstance != "" {
				tags["instance"] = vl.Identifier.PluginInstance
			}
			if vl.Identifier.Type != "" {
				tags["type"] = vl.Identifier.Type
			}
			if vl.Identifier.TypeInstance != "" {
				tags["type_instance"] = vl.Identifier.TypeInstance
			}

			// Drop invalid points
			m := metric.New(name, tags, fields, timestamp)

			metrics = append(metrics, m)
		}
	case "join":
		name := vl.Identifier.Plugin
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		for i := range vl.Values {
			switch value := vl.Values[i].(type) {
			case api.Gauge:
				fields[vl.DSName(i)] = float64(value)
			case api.Derive:
				fields[vl.DSName(i)] = float64(value)
			case api.Counter:
				fields[vl.DSName(i)] = float64(value)
			}

			if vl.Identifier.Host != "" {
				tags["host"] = vl.Identifier.Host
			}
			if vl.Identifier.PluginInstance != "" {
				tags["instance"] = vl.Identifier.PluginInstance
			}
			if vl.Identifier.Type != "" {
				tags["type"] = vl.Identifier.Type
			}
			if vl.Identifier.TypeInstance != "" {
				tags["type_instance"] = vl.Identifier.TypeInstance
			}
		}

		m := metric.New(name, tags, fields, timestamp)

		metrics = append(metrics, m)
	default:
		p.Log.Info("parse-multi-value config can only be 'split' or 'join'")
	}
	return metrics
}

func LoadTypesDB(path string) (*api.TypesDB, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return api.NewTypesDB(reader)
}

func init() {
	parsers.Add("collectd",
		func(_ string) telegraf.Parser {
			return &Parser{
				AuthFile: DefaultAuthFile,
			}
		})
}

func (p *Parser) InitFromConfig(config *parsers.Config) error {
	p.AuthFile = config.CollectdAuthFile
	p.SecurityLevel = config.CollectdSecurityLevel
	p.TypesDB = config.CollectdTypesDB
	p.ParseMultiValue = config.CollectdSplit

	return p.Init()
}
