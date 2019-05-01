package statsd

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

const (
	priorityNormal = "normal"
	priorityLow    = "low"
)

// this is adapted from datadog's apache licensed version at
// https://github.com/DataDog/datadog-agent/blob/fcfc74f106ab1bd6991dfc6a7061c558d934158a/pkg/dogstatsd/parser.go#L173
func (s *Statsd) parseDataDogEventMessage(message []byte, defaultHostname string) error {
	// _e{title.length,text.length}:title|text
	//  [
	//   |d:date_happened
	//   |p:priority
	//   |h:hostname
	//   |t:alert_type
	//   |s:source_type_nam
	//   |#tag1,tag2
	//  ]

	messageRaw := bytes.SplitN(message, []byte(":"), 2)
	if len(messageRaw) < 2 || len(messageRaw[0]) < 7 || len(messageRaw[1]) < 3 {
		return fmt.Errorf("Invalid message format")
	}
	header := messageRaw[0]
	message = messageRaw[1]

	rawLen := bytes.SplitN(header[3:], []byte(","), 2)
	if len(rawLen) != 2 {
		return fmt.Errorf("Invalid message format")
	}

	titleLen, err := strconv.ParseInt(string(rawLen[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid message format, could not parse title.length: '%s'", rawLen[0])
	}

	textLen, err := strconv.ParseInt(string(rawLen[1][:len(rawLen[1])-1]), 10, 64)
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

	// event := metrics.Event{
	// 	Priority:  metrics.EventPriorityNormal,
	// 	AlertType: metrics.EventAlertTypeInfo,
	// 	Title:     string(rawTitle),
	// 	Text:      string(bytes.Replace(rawText, []byte("\\n"), []byte("\n"), -1)),
	// }

	// Handle hostname, with a priority to the h: field, then the host:
	// tag and finally the defaultHostname value
	// Metadata
	m := cachedEvent{
		name: string(rawTitle),
	}
	m.tags = make(map[string]string, bytes.Count(message[1:], []byte{','})+1) // allocate for the approximate number of tags
	m.fields = make(map[string]interface{}, 8)
	m.fields["alert_type"] = "info" // default event type
	m.fields["text"] = string(rawText)
	m.tags["hostname"] = defaultHostname
	if len(message) > 1 {
		rawMetadataFields := bytes.Split(message[1:], []byte{'|'})
		for i := range rawMetadataFields {
			if len(rawMetadataFields[i]) < 2 {
				log.Printf("W! [inputs.statsd] too short metadata field")
			}
			switch string(rawMetadataFields[i]) {
			case "d:":
				ts, err := strconv.ParseInt(string(rawMetadataFields[i][2:]), 10, 64)
				if err != nil {
					log.Printf("W! [inputs.statsd] skipping timestamp: %s", err)
					continue
				}
				m.ts = time.Unix(ts, 0)
			case "p:":
				switch string(rawMetadataFields[i][2:]) {
				case priorityLow:
					m.fields["priority"] = priorityLow
				case priorityNormal:
					m.fields["priority"] = priorityNormal
				default:
					log.Printf("W! [inputs.statsd] skipping priority: %s", err)
					continue
				}
			case "h:":
				m.tags["hostname"] = string(rawMetadataFields[i][2:])
			case "t:":
				switch string(rawMetadataFields[i][2:]) {
				case "error":
					m.fields["alert_type"] = "error"
				case "warning":
					m.fields["alert_type"] = "warning"
				case "success":
					m.fields["alert_type"] = "success"
				case "info":
					m.fields["alert_type"] = "info"
				default:
					log.Printf("W! [inputs.statsd] skipping priority: %s", err)
					continue
				}
			case "k:":
				// TODO(docmerlin): does this make sense?
				m.tags["aggregation_key"] = string(rawMetadataFields[i][2:])
			case "s:":
				m.fields["source_type_name"] = string(rawMetadataFields[i][2:])
			case "#":
				parseDataDogTags(m.tags, string(rawMetadataFields[i][2:]))
				//event.Tags, hostFromTags = parseTags(, defaultHostname)
			default:
				log.Printf("W! [inputs.statsd] unknown metadata type: '%s'", rawMetadataFields[i])
			}
		}
	}
	return nil
}

func parseDataDogTags(tags map[string]string, message string) {
	//tags := make(map[string]string, strings.Count(message, ","))
	start := 0
	var k, v string
	for i := range message {
		switch message[i] {
		case ',':
			v = message[start:i]
			start = i + 1
			if k == "" || v == "" {
				continue
			}
			tags[k] = v
		case ':':
			k = message[start:i]
			start = i + 1
		}
	}
}
