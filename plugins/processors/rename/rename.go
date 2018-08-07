package rename

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"sync"
)

const sampleConfig = `
  ## Measurement, tag, and field renamings are stored in separate sub-tables.
  ## Specify one sub-table per rename operation.
  # [[processors.rename.measurement]]
  #   ## measurement to change
  #   from = "kilobytes_per_second"
  #   to = "kbps"

  # [[processors.rename.tag]]
  #   ## tag to change
  #   from = "host"
  #   to = "hostname"

  # [[processors.rename.field]]
  #   ## field to change
  #   from = "lower"
  #   to = "min"

  # [[processors.rename.field]]
  #   ## field to change
  #   from = "upper"
  #   to = "max"
`

type renamer struct {
	From string
	To   string
}

type Rename struct {
	Measurement  []renamer
	Tag          []renamer
	Field        []renamer
	measurements map[string]string
	tags         map[string]string
	fields       map[string]string
	once         sync.Once
}

func (r *Rename) SampleConfig() string {
	return sampleConfig
}

func (r *Rename) Description() string {
	return "Rename measurements, tags, and fields that pass through this filter."
}

func (r *Rename) Apply(in ...telegraf.Metric) []telegraf.Metric {
	r.once.Do(r.init)

	for _, point := range in {
		if newMeasurementName, ok := r.measurements[point.Name()]; ok {
			point.SetName(newMeasurementName)
		}
		for oldTagName, tagValue := range point.Tags() {
			if newTagName, ok := r.tags[oldTagName]; ok {
				point.RemoveTag(oldTagName)
				point.AddTag(newTagName, tagValue)
			}
		}
		for oldFieldName, fieldValue := range point.Fields() {
			if newFieldName, ok := r.fields[oldFieldName]; ok {
				point.RemoveField(oldFieldName)
				point.AddField(newFieldName, fieldValue)
			}
		}
	}

	return in
}

func (r *Rename) init() {
	if r.measurements == nil || r.tags == nil || r.fields == nil {
		r.measurements = make(map[string]string, len(r.Measurement))
		for _, o := range r.Measurement {
			r.measurements[o.From] = o.To
		}
		r.tags = make(map[string]string, len(r.Tag))
		for _, o := range r.Tag {
			r.tags[o.From] = o.To
		}
		r.fields = make(map[string]string, len(r.Field))
		for _, o := range r.Field {
			r.fields[o.From] = o.To
		}
	}
}

func init() {
	processors.Add("rename", func() telegraf.Processor {
		return &Rename{}
	})
}
