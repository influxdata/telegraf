package taglimit

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"log"
)

const sampleConfig = `
  ## Maximum number of tags to preserve
  limit = 10

  ## List of tags to preferentially preserve
  keep = ["foo", "bar", "baz"]
`

type TagLimit struct {
	Limit int      `toml:"limit"`
	Keep  []string `toml:"keep"`
}

func (d *TagLimit) SampleConfig() string {
	return sampleConfig
}

func (d *TagLimit) Description() string {
	return "Restricts the number of tags that can pass through this filter and chooses which tags to preserve when over the limit."
}

func (d *TagLimit) Apply(in ...telegraf.Metric) []telegraf.Metric {
	// convert list of tags to a map so we can do constant-time lookups
	keepTags := make(map[string]string)
	for _, tag := range d.Keep {
		keepTags[tag] = ""
	}
	for _, point := range in {
		pointTags := point.Tags()
		if len(pointTags) <= d.Limit {
			continue
		}
		// remove extraneous tags, stop once we're at the limit
		for k := range pointTags {
			if _, ok := keepTags[k]; !ok {
				delete(pointTags, k)
			}
			if len(pointTags) <= d.Limit {
				break
			}
		}
		// we've trimmed off all the non-keep tags but we're still
		// over the limit, so start trimming
		for k := range pointTags {
			if len(pointTags) > d.Limit {
				log.Printf("Warning: point has %d tags even after trimming non-preferred tags; removing tag %s", len(pointTags), pointTags[k])
				delete(pointTags, k)
			} else {
				break
			}
		}
		// update point tags with trimmed map
		point.SetTags(pointTags)
	}

	return in
}

func init() {
	processors.Add("tag_limit", func() telegraf.Processor {
		return &TagLimit{}
	})
}
