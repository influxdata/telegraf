package json_template

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
)

const definitions = `
{{$metrics := .}}

{{define "fields"}}
	{{- $comma := "" -}}
	{{- range $key, $value := .Fields -}}
		{{- $comma -}}{{- $comma = "," -}}
		"{{$key}}": {{$value | printf "%#v"}}
	{{- end -}}
{{end}}

{{define "tags"}}
	{{- $comma := "" -}}
	{{- range $key, $value := .Tags -}}
		{{- $comma -}}{{- $comma = "," -}}
		"{{$key}}": "{{$value}}"
	{{- end -}}
{{end}}
`

type tmplMetric struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}

type Serializer struct {
	Style string
	tmpl  *template.Template
}

func NewSerializer(templateCfg string, style string) (*Serializer, error) {
	if !choice.Contains(style, []string{"", "raw", "compact", "pretty"}) {
		return nil, fmt.Errorf("unknown style %q", style)
	}
	s := &Serializer{Style: style}

	funcMap := template.FuncMap{
		"last": func(n int, arr interface{}) bool {
			v := reflect.ValueOf(arr)
			switch v.Kind() {
			case reflect.Array, reflect.Slice:
				return n == v.Len()-1
			}
			panic(errors.New("'last' can only be used with arrays or slices"))
		},
		"uppercase": strings.ToUpper,
		"lowercase": strings.ToLower,
	}
	tmpl, err := template.New("test").Funcs(funcMap).Parse(definitions + templateCfg)
	if err != nil {
		return nil, err
	}
	s.tmpl = tmpl

	return s, nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	object := s.createObject(metric)

	var buf bytes.Buffer
	if err := s.tmpl.Execute(&buf, object); err != nil {
		return nil, err
	}
	if !json.Valid(buf.Bytes()) {
		// We got some invalid JSON, try to find out what's wrong
		var dummy map[string]interface{}
		return buf.Bytes(), json.Unmarshal(buf.Bytes(), &dummy)
	}

	return s.format(buf.Bytes())
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	objects := make([]tmplMetric, 0, len(metrics))
	for _, metric := range metrics {
		objects = append(objects, s.createObject(metric))
	}

	var buf bytes.Buffer
	if err := s.tmpl.Execute(&buf, objects); err != nil {
		return nil, err
	}
	if !json.Valid(buf.Bytes()) {
		// We got some invalid JSON, try to find out what's wrong
		var dummy map[string]interface{}
		return buf.Bytes(), json.Unmarshal(buf.Bytes(), &dummy)
	}

	return s.format(buf.Bytes())
}

func (s *Serializer) createObject(metric telegraf.Metric) tmplMetric {
	return tmplMetric{
		Name:   metric.Name(),
		Tags:   metric.Tags(),
		Fields: metric.Fields(),
		Time:   metric.Time(),
	}
}

func (s *Serializer) format(data []byte) ([]byte, error) {
	switch s.Style {
	case "", "raw":
		return data, nil
	case "compact":
		var buf bytes.Buffer
		if err := json.Compact(&buf, data); err != nil {
			return nil, fmt.Errorf("compacting json failed: %v", err)
		}
		return buf.Bytes(), nil
	case "pretty":
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return nil, fmt.Errorf("prettifying json failed: %v", err)
		}
		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("unknown style %q", s.Style)
}
