package statsd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	priorityNormal = "normal"
	priorityLow    = "low"
)

var uncommenter = strings.NewReplacer("\\n", "\n")

// this is adapted from datadog's apache licensed version at
// https://github.com/DataDog/datadog-agent/blob/fcfc74f106ab1bd6991dfc6a7061c558d934158a/pkg/dogstatsd/parser.go#L173
func (s *Statsd) parseEventMessage(now time.Time, message string, defaultHostname string) error {
	// _e{title.length,text.length}:title|text
	//  [
	//   |d:date_happened
	//   |p:priority
	//   |h:hostname
	//   |t:alert_type
	//   |s:source_type_nam
	//   |#tag1,tag2
	//  ]
	//
	//
	// tag is key:value
	messageRaw := strings.SplitN(message, ":", 2)
	if len(messageRaw) < 2 || len(messageRaw[0]) < 7 || len(messageRaw[1]) < 3 {
		return fmt.Errorf("Invalid message format")
	}
	header := messageRaw[0]
	message = messageRaw[1]

	rawLen := strings.SplitN(header[3:], ",", 2)
	if len(rawLen) != 2 {
		return fmt.Errorf("Invalid message format")
	}

	titleLen, err := strconv.ParseInt(rawLen[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid message format, could not parse title.length: '%s'", rawLen[0])
	}

	textLen, err := strconv.ParseInt(rawLen[1][:len(rawLen[1])-1], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid message format, could not parse text.length: '%s'", rawLen[0])
	}
	if titleLen+textLen+1 > int64(len(message)) {
		return fmt.Errorf("Invalid message format, title.length and text.length exceed total message length")
	}

	rawTitle := message[:titleLen]
	rawText := message[titleLen+1 : titleLen+1+textLen]
	message = message[titleLen+1+textLen:]

	if len(rawTitle) == 0 || len(rawText) == 0 {
		return fmt.Errorf("Invalid event message format: empty 'title' or 'text' field")
	}

	// Handle hostname, with a priority to the h: field, then the host:
	// tag and finally the defaultHostname value
	// Metadata
	m := cachedEvent{
		name: rawTitle,
	}
	m.tags = make(map[string]string, strings.Count(message, ",")+2) // allocate for the approximate number of tags
	m.fields = make(map[string]interface{}, 9)
	m.fields["alert-type"] = "info" // default event type
	m.fields["text"] = uncommenter.Replace(string(rawText))
	m.tags["source"] = defaultHostname
	m.fields["priority"] = priorityNormal
	m.ts = now
	if len(message) == 0 {
		s.events = append(s.events, m)
		return nil
	}

	if len(message) > 1 {
		rawMetadataFields := strings.Split(message[1:], "|")
		for i := range rawMetadataFields {
			if len(rawMetadataFields[i]) < 2 {
				log.Printf("W! [inputs.statsd] too short metadata field")
			}
			switch rawMetadataFields[i] {
			case "d:":
				ts, err := strconv.ParseInt(rawMetadataFields[i][2:], 10, 64)
				if err != nil {
					log.Printf("W! [inputs.statsd] skipping timestamp: %s", err)
					continue
				}
				m.fields["ts"] = ts
			case "p:":
				switch rawMetadataFields[i][2:] {
				case priorityLow:
					m.fields["priority"] = priorityLow
				case priorityNormal: // we already used this as a default
				default:
					log.Printf("W! [inputs.statsd] skipping priority")
					continue
				}
			case "h:":
				m.tags["source"] = rawMetadataFields[i][2:]
			case "t:":
				switch rawMetadataFields[i][2:] {
				case "error":
					m.fields["alert-type"] = "error"
				case "warning":
					m.fields["alert-type"] = "warning"
				case "success":
					m.fields["alert-type"] = "success"
				case "info": // already set for info
				default:
					log.Printf("W! [inputs.statsd] skipping alert type")
					continue
				}
			case "k:":
				// TODO(docmerlin): does this make sense?
				m.tags["aggregation-key"] = rawMetadataFields[i][2:]
			case "s:":
				m.fields["source-type-name"] = rawMetadataFields[i][2:]
			case "#":
				parseDataDogTags(m.tags, rawMetadataFields[i][2:])
			default:
				log.Printf("W! [inputs.statsd] unknown metadata type: '%s'", rawMetadataFields[i])
			}
		}
	}
	s.Lock()
	s.events = append(s.events, m)
	s.Unlock()
	return nil
}

func parseDataDogTags(tags map[string]string, message string) {
	start, i := 0, 0
	var k, v string
	var inVal bool // check if we are parsing the value part of the tag
	for i = range message {
		if message[i] == ',' {
			v = message[start:i]
			start = i + 1
			if k == "" {
				continue
			}
			tags[k] = v
			k, v, inVal = "", "", false // reset state vars
		} else if message[i] == ':' && !inVal {
			k = message[start:i]
			start = i + 1
			inVal = true
		}
	}
	// grab the last value
	if k != "" {
		tags[k] = message[start : i+1]
	}
}
