package collectd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"collectd.org/api"
	"collectd.org/network"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const (
	DefaultAuthFile = "/etc/collectd/auth_file"
)

type CollectdParser struct {
	// DefaultTags will be added to every parsed metric
	DefaultTags map[string]string

	//whether or not to split multi value metric into multiple metrics
	//default value is split
	ParseMultiValue string
	popts           network.ParseOpts
}

func (p *CollectdParser) SetParseOpts(popts *network.ParseOpts) {
	p.popts = *popts
}

func NewCollectdParser(
	authFile string,
	securityLevel string,
	typesDB []string,
	split string,
) (*CollectdParser, error) {
	popts := network.ParseOpts{}

	switch securityLevel {
	case "none":
		popts.SecurityLevel = network.None
	case "sign":
		popts.SecurityLevel = network.Sign
	case "encrypt":
		popts.SecurityLevel = network.Encrypt
	default:
		popts.SecurityLevel = network.None
	}

	if authFile == "" {
		authFile = DefaultAuthFile
	}
	popts.PasswordLookup = network.NewAuthFile(authFile)

	for _, path := range typesDB {
		db, err := LoadTypesDB(path)
		if err != nil {
			return nil, err
		}

		if popts.TypesDB != nil {
			popts.TypesDB.Merge(db)
		} else {
			popts.TypesDB = db
		}
	}

	parser := CollectdParser{popts: popts,
		ParseMultiValue: split}
	return &parser, nil
}

func (p *CollectdParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	valueLists, err := network.Parse(buf, p.popts)
	if err != nil {
		return nil, fmt.Errorf("Collectd parser error: %s", err)
	}

	metrics := []telegraf.Metric{}
	for _, valueList := range valueLists {
		metrics = append(metrics, UnmarshalValueList(valueList, p.ParseMultiValue)...)
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

func (p *CollectdParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := p.Parse([]byte(line))
	if err != nil {
		return nil, err
	}

	if len(metrics) != 1 {
		return nil, errors.New("Line contains multiple metrics")
	}

	return metrics[0], nil
}

func (p *CollectdParser) SetDefaultTags(tags map[string]string) {
	p.DefaultTags = tags
}

// UnmarshalValueList translates a ValueList into a Telegraf metric.
func UnmarshalValueList(vl *api.ValueList, multiValue string) []telegraf.Metric {
	timestamp := vl.Time.UTC()

	var metrics []telegraf.Metric

	//set multiValue to default "split" if nothing is specified
	if multiValue == "" {
		multiValue = "split"
	}
	switch multiValue {
	case "split":
		for i := range vl.Values {
			var name string
			name = fmt.Sprintf("%s_%s", vl.Identifier.Plugin, vl.DSName(i))
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
			m, err := metric.New(name, tags, fields, timestamp)
			if err != nil {
				log.Printf("E! Dropping metric %v: %v", name, err)
				continue
			}

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

		m, err := metric.New(name, tags, fields, timestamp)
		if err != nil {
			log.Printf("E! Dropping metric %v: %v", name, err)
		}

		metrics = append(metrics, m)
	default:
		log.Printf("parse-multi-value config can only be 'split' or 'join'")
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
