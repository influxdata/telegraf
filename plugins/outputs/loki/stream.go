package loki

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type (
	Streams struct {
		Streams []Stream `json:"streams"`
	}

	Stream struct {
		key string `json:"-"`

		Labels map[string]string `json:"stream"`
		Logs   []Log             `json:"values"`
	}

	Log struct {
		Timestamp int64
		Line      string
	}
)

func (l Log) MarshalJSON() ([]byte, error) {
	v := fmt.Sprintf("[\"%d\",\"%s\"]", l.Timestamp, l.Line)

	return []byte(v), nil
}

func (s *Streams) insertLog(ts []*telegraf.Tag, l Log) {
	var (
		key   = uniqKeyFromTagList(ts)
		index int
		found bool
	)

	for i, s := range s.Streams {
		if s.key == key {
			index, found = i, true
			break
		}
	}

	if !found {
		s.Streams = append(s.Streams, newStream(key, ts))
		index = len(s.Streams) - 1
	}

	s.Streams[index].Logs = append(s.Streams[index].Logs, l)
}

func uniqKeyFromTagList(ts []*telegraf.Tag) (k string) {
	for _, t := range ts {
		k += fmt.Sprintf("%s%s", t.Key, t.Value)
	}

	return
}

func newStream(uniqKey string, ts []*telegraf.Tag) Stream {
	s := Stream{
		key:    uniqKey,
		Logs:   make([]Log, 0),
		Labels: map[string]string{},
	}

	for _, t := range ts {
		s.Labels[t.Key] = t.Value
	}

	return s
}
