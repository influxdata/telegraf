package tag_limit

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type TagLimit struct {
	Limit    int             `toml:"limit"`
	Keep     []string        `toml:"keep"`
	Log      telegraf.Logger `toml:"-"`
	init     bool
	keepTags map[string]string
}

func (d *TagLimit) initOnce() error {
	if d.init {
		return nil
	}
	if len(d.Keep) > d.Limit {
		return fmt.Errorf("%d keep tags is greater than %d total tag limit", len(d.Keep), d.Limit)
	}
	d.keepTags = make(map[string]string)
	// convert list of tags-to-keep to a map so we can do constant-time lookups
	for _, tagKey := range d.Keep {
		d.keepTags[tagKey] = ""
	}
	d.init = true
	return nil
}

func (d *TagLimit) Apply(in ...telegraf.Metric) []telegraf.Metric {
	err := d.initOnce()
	if err != nil {
		d.Log.Errorf("Could not create tag_limit processor: %v", err)
		return in
	}
	for _, point := range in {
		pointOriginalTags := point.TagList()
		lenPointTags := len(pointOriginalTags)
		if lenPointTags <= d.Limit {
			continue
		}
		tagsToRemove := make([]string, lenPointTags-d.Limit)
		removeIdx := 0
		// remove extraneous tags, stop once we're at the limit
		for _, t := range pointOriginalTags {
			if _, ok := d.keepTags[t.Key]; !ok {
				tagsToRemove[removeIdx] = t.Key
				removeIdx++
				lenPointTags--
			}
			if lenPointTags <= d.Limit {
				break
			}
		}
		for _, t := range tagsToRemove {
			point.RemoveTag(t)
		}
	}

	return in
}

func init() {
	processors.Add("tag_limit", func() telegraf.Processor {
		return &TagLimit{}
	})
}
