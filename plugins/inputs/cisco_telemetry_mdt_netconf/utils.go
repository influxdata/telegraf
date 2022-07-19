package cisco_telemetry_mdt_netconf

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// State is a flag that tracks the state of an object
type state uint

// Types of state
const (
	initialized = iota + 1
	updated
)

// Numerical base
const base = 10

type metricTracker struct {
	testutil.Metric
	measurementFlag  uint
	measurementLevel uint
}

// lookupMeasurement composes the measurement and then searches for it
// in a predefined list of xpaths
func (m *metricTracker) lookupMeasurement(
	xpaths map[string]interface{}, e stackElem) {
	if m.measurementFlag == initialized {
		if len(e.Tree.Children) > 0 {
			if len(m.Measurement) != 0 {
				m.Measurement = m.Measurement + "/"
			}
			m.Measurement = m.Measurement + e.Tree.XMLName.Local
		}
		if _, measurementFound := xpaths[m.Measurement]; measurementFound {
			m.measurementFlag = updated
		}
		m.measurementLevel++
	}
}

// addValue adds the value of the stack element e to the metric.
// The value is stored either as a tag (if its name is in a predefined
// list of tags), or as a field.
func (m *metricTracker) addValue(rn string, path []string,
	tags map[string]interface{}, e stackElem) error {
	tag := strings.Join(path, "/")
	if _, tagFound := tags[tag]; !tagFound {
		// Set the entire traversal path as tag only for
		// the fields that are deeper than the measurement level
		if uint(len(path))-1 <= m.measurementLevel {
			m.Fields[e.Tree.XMLName.Local] = e.Tree.Value
		} else {
			m.Fields[strings.Join(path[m.measurementLevel:],
				"/")] = e.Tree.Value
		}
	} else {
		// Append namespace sent with XML reply
		tag = rn + tag
		switch value := e.Tree.Value.(type) {
		case uint64:
			m.Tags[tag] = strconv.FormatUint(value, base)
		case int64:
			m.Tags[tag] = strconv.FormatInt(value, base)
		case float64:
			m.Tags[tag] = strconv.FormatFloat(value, 'f', -1, 64)
		case bool:
			m.Tags[tag] = strconv.FormatBool(value)
		case string:
			m.Tags[tag] = value
		default:
			// Assume type is unknown
			// Return unknown data type error
			return fmt.Errorf(
				"failed to traverse telemetry tree: field %v has value %v with unsupported tag data type %T",
				e.Tree.XMLName.Local,
				value,
				value,
			)
		}
	}
	return nil
}

// removeValues removes fields and tags from a metric
func (m *metricTracker) removeValues(path []string, e stackElem) {
	// Remove fields
	for f := range m.Fields {
		if uint(len(path)) <= m.measurementLevel {
			if strings.Contains(f, e.Tree.XMLName.Local) {
				delete(m.Fields, f)
			}
		} else {
			if strings.Contains(f, path[len(path)-1]+"__") {
				delete(m.Fields, f)
			}
		}
	}

	// Remove tags
	for t := range m.Tags {
		if strings.Contains(t, path[len(path)-1]+"__") {
			delete(m.Tags, t)
		}
	}
}

// Representation of a tree as an element of a stack data structure
// Has Tree as TelemetryTree
// Has Child as counter of traversed children
// Has Tags as counter of traversed tags
type stackElem struct {
	Tree  TelemetryTree
	Child int
}

// getCoreName trims the prefix and suffix in a namespace and returns
// the core namespace
func getCoreName(s string) string {
	// Trim namespaces of format:
	// http://cisco.com/ns/yang/Cisco-IOS-XE-memory-oper
	splits := strings.Split(strings.TrimSuffix(s, "/"), "/")
	if len(splits) == 1 {
		// Trim namespaces of format:
		// urn:ietf:params:xml:ns:yang:ietf-interfaces
		splits = strings.Split(splits[0], ":")
	}
	// Get the root of the encoding path
	return splits[len(splits)-1]
}

// TelemetryTree defines a structure used for XML unmarshaling
type TelemetryTree struct {
	XMLName  xml.Name        `xml:""`
	Children []TelemetryTree `xml:",any"`
	Value    interface{}     `xml:",chardata"`
}

// TraverseTree traverses an XML tree and builds entries in Influx LINE format
func (t *TelemetryTree) TraverseTree(
	userTags map[string]interface{},
	userXpaths map[string]interface{},
	source string, timestamp time.Time) (*metric.SeriesGrouper, error) {
	m := &metricTracker{
		Metric: testutil.Metric{
			Tags: map[string]string{
				"source": source,
			},
			Fields: make(map[string]interface{}),
			Time:   timestamp,
		},
		measurementFlag:  initialized,
		measurementLevel: uint(0),
	}

	grouper := metric.NewSeriesGrouper()

	// Stack used for tree traversal
	stack := make([]stackElem, 0)
	// Current node in the tree, initally set to root
	current := stackElem{Tree: *t}
	// Path as tree is traversed
	traversalPath := make([]string, 0)

	rootNamespace := getCoreName(current.Tree.XMLName.Space)
	if rootNamespace != "" {
		rootNamespace = rootNamespace + ":"
	} else {
		rootNamespace = "/"
	}

	// Loop as long as there is something in the stack
	for {
		// Node to be processed exists
		if current.Tree.XMLName != (xml.Name{}) {
			m.lookupMeasurement(userXpaths, current)

			// If no children of this node have yet been traversed
			if current.Child == 0 {
				// Push node to the stack
				stack = append(stack, current)

				// Mark the traversal path
				traversalPath = append(traversalPath, current.Tree.XMLName.Local)
			}

			// If the node is a leaf node, then store its value and pop the node
			if len(current.Tree.Children) == 0 {
				if e := m.addValue(rootNamespace,
					traversalPath, userTags, current); e != nil {
					return nil, e
				}

				// Pop node name from the traversal path
				if len(traversalPath) > 0 {
					traversalPath = traversalPath[:len(traversalPath)-1]
				}

				// Pop node from the stack
				stack = stack[:len(stack)-1]
				// Reset node to nil to enforce going to the next node
				current = stackElem{}
			} else {
				// Reset node to its next child
				current = stackElem{Tree: current.Tree.Children[current.Child]}
			}
		} else {
			// If the stack is empty, then tree traversal is complete
			if len(stack) == 0 {
				break
			} else {
				// Pop the next node from the stack
				current = stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				// If this node still has children to traverse, then
				// reinitialize the node with its next child
				if current.Child < (len(current.Tree.Children) - 1) {
					// Push node to the stack
					stack = append(stack, (stackElem{
						Tree:  current.Tree,
						Child: current.Child + 1,
					}))
					// Reset node to next child
					current = stackElem{Tree: current.Tree.Children[current.Child+1]}
				} else {
					// Assume we traversed all children of the node
					for k, v := range m.Fields {
						// Update the measurement with namespace and add the
						//fields and tags to the metric.SeriesGrouper
						grouper.Add(rootNamespace+m.Measurement, m.Tags, m.Time, k, v)
					}

					m.removeValues(traversalPath, current)

					// Pop element from traversal path
					if len(traversalPath) > 0 {
						traversalPath = traversalPath[:len(traversalPath)-1]
					}

					// Reset node to nil to enforce going to the next node
					current = stackElem{}
				}
			}
		}
	}
	return grouper, nil
}

// UnmarshalXML implements the UnmarshalXML method of xml.Unmarshaler interface
// for a TelemetryTree
func (t *TelemetryTree) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	t.XMLName = start.Name
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch tokenInstance := token.(type) {

		// This token marks the beginning of a subtree
		case xml.StartElement:
			var child TelemetryTree
			if err := child.UnmarshalXML(d, tokenInstance); err != nil {
				// Return UnmarshalXML error
				return err
			}

			// Add child tree to list of children for current node
			t.Children = append(t.Children, child)

		// This token marks the end of a subtree
		case xml.EndElement:
			// Check if end tag matches start tag
			if tokenInstance.Name.Space == start.Name.Space &&
				tokenInstance.Name.Local == start.Name.Local {
				return nil
			} else {
				// Return wrong end element error
				return fmt.Errorf(
					"failed to unmarshal XML: wrong XML end element - expected %q, %q instead of %q, %q",
					start.Name.Space,
					start.Name.Local,
					tokenInstance.Name.Space,
					tokenInstance.Name.Local,
				)
			}

		// This token marks XML character data
		case xml.CharData:
			data := string(tokenInstance)

			// Infer data type
			if data != "" {
				if value, err := strconv.ParseUint(data, base, 64); err == nil {
					// This is an uint. Represent it on 64 bits. Infer the base.
					t.Value = uint64(value)
				} else if value, err := strconv.ParseInt(data, base, 64); err == nil {
					// This is an int. Represent it on 64 bits. Infer the base.
					t.Value = int64(value)
				} else if value, err := strconv.ParseFloat(data, 64); err == nil {
					// This is a float. Represent it on 64 bits.
					t.Value = float64(value)
				} else if value, err := strconv.ParseBool(data); err == nil {
					// This is a bool
					t.Value = bool(value)
				} else {
					// Assume this is a string
					t.Value = string(data)
				}
			}

		// Otherwise
		default:
			// Return unknown token type error
			return fmt.Errorf(
				"failed to unmarshal XML: unknown XML token type %T",
				token,
			)
		}
	}
}
