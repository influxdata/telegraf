package nats

import (
	"text/template/parse"

	"github.com/influxdata/telegraf"
)

type subMsgPair struct {
	subject string
	metric  telegraf.Metric
}
type metricSubjectTmplCtx struct {
	Name     string
	getTag   func(string) string
	getField func() string
}

func (m metricSubjectTmplCtx) GetTag(key string) string {
	return m.getTag(key)
}

func (m metricSubjectTmplCtx) Field() string {
	return m.getField()
}

func createmetricSubjectTmplCtx(metric telegraf.Metric) metricSubjectTmplCtx {
	return metricSubjectTmplCtx{
		Name: metric.Name(),
		getTag: func(key string) string {
			tagList := metric.TagList()
			for _, tag := range tagList {
				if tag.Key == key {
					return tag.Value
				}
			}
			return ""
		},
		getField: func() string {
			fields := metric.FieldList()
			if len(fields) == 0 {
				return "emptyFields"
			}
			if len(fields) > 1 {
				return "tooManyFields"
			}

			return fields[0].Key
		},
	}
}

// Check the template for any references to `.Field`.
// If the template includes a `.Field` reference, we will need to split the metric
// into separate messages based on the field.
func usesFieldField(node parse.Node) bool {
	switch n := node.(type) {
	case *parse.ListNode:
		for _, sub := range n.Nodes {
			if usesFieldField(sub) {
				return true
			}
		}
	case *parse.ActionNode:
		return usesFieldField(n.Pipe)
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			if usesFieldField(cmd) {
				return true
			}
		}
	case *parse.CommandNode:
		for _, arg := range n.Args {
			if usesFieldField(arg) {
				return true
			}
		}
	case *parse.FieldNode:
		// .Field will be represented as []string{"Field"}
		return len(n.Ident) == 1 && n.Ident[0] == "Field"
	}
	return false
}

// splitMetricByField will create a new metric that only contains the specified field.
// This is used when the user wants to include the field name in the subject.
func splitMetricByField(metric telegraf.Metric, field string) telegraf.Metric {
	metricCopy := metric.Copy()

	for _, f := range metric.FieldList() {
		if f.Key != field {
			// Remove all fields that are not the specified field
			metricCopy.RemoveField(f.Key)
		}
	}

	return metricCopy
}
