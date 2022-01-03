package filepath

import (
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Options struct {
	BaseName []BaseOpts `toml:"basename"`
	DirName  []BaseOpts `toml:"dirname"`
	Stem     []BaseOpts
	Clean    []BaseOpts
	Rel      []RelOpts
	ToSlash  []BaseOpts `toml:"toslash"`
}

type ProcessorFunc func(s string) string

// BaseOpts contains options applicable to every function
type BaseOpts struct {
	Field string
	Tag   string
	Dest  string
}

type RelOpts struct {
	BaseOpts
	BasePath string
}

const sampleConfig = `
  ## Treat the tag value as a path and convert it to its last element, storing the result in a new tag 
  # [[processors.filepath.basename]]
  #   tag = "path"
  #   dest = "basepath"

  ## Treat the field value as a path and keep all but the last element of path, typically the path's directory 
  # [[processors.filepath.dirname]]
  #   field = "path"

  ## Treat the tag value as a path, converting it to its the last element without its suffix
  # [[processors.filepath.stem]]
  #   tag = "path"

  ## Treat the tag value as a path, converting it to the shortest path name equivalent
  ## to path by purely lexical processing 
  # [[processors.filepath.clean]]
  #   tag = "path"

  ## Treat the tag value as a path, converting it to a relative path that is lexically
  ## equivalent to the source path when joined to 'base_path' 
  # [[processors.filepath.rel]]
  #   tag = "path"
  #   base_path = "/var/log"

  ## Treat the tag value as a path, replacing each separator character in path with a '/' character. Has only
  ## effect on Windows
  # [[processors.filepath.toslash]]
  #   tag = "path"
`

func (o *Options) SampleConfig() string {
	return sampleConfig
}

func (o *Options) Description() string {
	return "Performs file path manipulations on tags and fields"
}

// applyFunc applies the specified function to the metric
func (o *Options) applyFunc(bo BaseOpts, fn ProcessorFunc, metric telegraf.Metric) {
	if bo.Tag != "" {
		if v, ok := metric.GetTag(bo.Tag); ok {
			targetTag := bo.Tag

			if bo.Dest != "" {
				targetTag = bo.Dest
			}
			metric.AddTag(targetTag, fn(v))
		}
	}

	if bo.Field != "" {
		if v, ok := metric.GetField(bo.Field); ok {
			targetField := bo.Field

			if bo.Dest != "" {
				targetField = bo.Dest
			}

			// Only string fields are considered
			if v, ok := v.(string); ok {
				metric.AddField(targetField, fn(v))
			}
		}
	}
}

func stemFilePath(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// processMetric processes fields and tag values for a given metric applying the selected transformations
func (o *Options) processMetric(metric telegraf.Metric) {
	// Stem
	for _, v := range o.Stem {
		o.applyFunc(v, stemFilePath, metric)
	}
	// Basename
	for _, v := range o.BaseName {
		o.applyFunc(v, filepath.Base, metric)
	}
	// Rel
	for _, v := range o.Rel {
		o.applyFunc(v.BaseOpts, func(s string) string {
			relPath, _ := filepath.Rel(v.BasePath, s)
			return relPath
		}, metric)
	}
	// Dirname
	for _, v := range o.DirName {
		o.applyFunc(v, filepath.Dir, metric)
	}
	// Clean
	for _, v := range o.Clean {
		o.applyFunc(v, filepath.Clean, metric)
	}
	// ToSlash
	for _, v := range o.ToSlash {
		o.applyFunc(v, filepath.ToSlash, metric)
	}
}

func (o *Options) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, m := range in {
		o.processMetric(m)
	}

	return in
}

func init() {
	processors.Add("filepath", func() telegraf.Processor {
		return &Options{}
	})
}
