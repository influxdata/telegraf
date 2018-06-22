package lookup

import (
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
[processors.lookup]
lookups = [ '/path/to/model1', '/path/to/model2' ]    
`

type Lookup struct {
	Match        []string
	Trim         string
	Tagmap       map[string]string
	FieldAdd     map[string]string
	FieldMatch   string
	FieldReplace map[string]string
	TagAdd       map[string]string
	Message      string

	re           []*regexp.Regexp
	fields       *regexp.Regexp
	lookups      []map[string]string
	configParsed bool
}

func (p *Lookup) SampleConfig() string {
	return sampleConfig
}

func (p *Lookup) Description() string {
	return "Apply metric modifications using override semantics."
}

func (p *Lookup) ParseConfig() bool {
	for _, m := range p.Match {
		p.re = append(p.re, regexp.MustCompile(m))
	}

	p.fields = regexp.MustCompile(`\${([A-Za-z0-9_]+?)}`)

	return true
}

func (p *Lookup) Fields(in string) []string {
	var out []string

	for _, group := range p.fields.FindAllStringSubmatch(in, -1) {
		if len(group) == 2 {
			out = append(out, group[1])
		}
	}

	return out
}

func (p *Lookup) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if !p.configParsed {
		p.configParsed = p.ParseConfig()
	}

	for _, metric := range in {
		for fieldkey, fieldval := range metric.Fields() {
			m, err := regexp.MatchString(p.FieldMatch, fieldkey)
			if m && err == nil {
				v, ok := fieldval.(string)
				if ok {
					for _, re := range p.re {
						// Match the expression and create a map according to the match names
						matches := re.FindStringSubmatch(v)
						if matches == nil {
							continue
						}
						results := make(map[string]string)
						for i, name := range re.SubexpNames() {
							if i < len(matches) {
								results[name] = matches[i]
								keyname := name + "_" + matches[i]
								if v, ok := p.Tagmap[keyname]; ok {
									results[name] = v
								}
							}
						}

						// Substitute in all of the fields
						for key, value := range p.FieldAdd {
							for _, match := range p.Fields(value) {
								value = strings.Replace(value, "${"+match+"}", results[match], 1)
							}

							if len(value) > 0 {
								metric.AddField(key, value)
							}
						}

						// Substitute in all the tags
						for key, value := range p.TagAdd {
							for _, match := range p.Fields(value) {
								value = strings.Replace(value, "${"+match+"}", results[match], 1)
							}

							if len(value) > 0 {
								metric.AddTag(key, value)
							}
						}

						// Replace fields
						for key, value := range p.FieldReplace {
							for _, match := range p.Fields(value) {
								value = strings.Replace(value, "${"+match+"}", results[match], 1)
							}

							if len(value) > 0 {
								//metric.RemoveField(fieldkey)
								metric.AddField(key, value)
							}
						}
					}
				}
			}
		}
	}

	return in
}

func init() {
	processors.Add("lookup", func() telegraf.Processor {
		return &Lookup{}
	})
}
