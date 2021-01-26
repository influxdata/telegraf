package loki

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
)

type (
	Log []string

	Streams map[string]*Stream

	Stream struct {
		Labels map[string]string `json:"stream"`
		Logs   []Log             `json:"values"`
	}

	Request struct {
		Streams []Stream `json:"streams"`
	}
)

func (s Streams) insertLog(ts []*telegraf.Tag, l Log) {
	key := uniqKeyFromTagList(ts)

	if _, ok := s[key]; !ok {
		s[key] = newStream(ts)
	}

	s[key].Logs = append(s[key].Logs, l)
}

func (s Streams) MarshalJSON() ([]byte, error) {
	r := Request{
		Streams: make([]Stream, 0, len(s)),
	}

	for _, stream := range s {
		r.Streams = append(r.Streams, *stream)
	}

	return json.Marshal(r)
}

func uniqKeyFromTagList(ts []*telegraf.Tag) (k string) {
	for _, t := range ts {
		k += fmt.Sprintf("%s-%s-",
			strings.ReplaceAll(t.Key, "-", "--"),
			strings.ReplaceAll(t.Value, "-", "--"),
		)
	}

	return k
}

func newStream(ts []*telegraf.Tag) *Stream {
	s := &Stream{
		Logs:   make([]Log, 0),
		Labels: map[string]string{},
	}

	for _, t := range ts {
		s.Labels[t.Key] = t.Value
	}

	return s
}
