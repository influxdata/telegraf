package expr

import (
	"log"

	"github.com/antonmedv/expr"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `

`

type ExprMetric struct {
	telegraf.Metric
}

type Expr struct {
	Expressions []string
}

func (p *Expr) SampleConfig() string {
	return sampleConfig
}

func (p *Expr) Description() string {
	return "Run LUA code against metrics"
}

func (em *ExprMetric) SetField(field string, value interface{}) {
	em.RemoveField(field)
	em.AddField(field, value)
}

func (em *ExprMetric) SetTag(tag string, value string) {
	em.RemoveTag(tag)
	em.AddTag(tag, value)
}

func (em *ExprMetric) SeparateFieldToTag(field string, newfield string, newtag string) {
	if em.HasField(field) {
		val, _ := em.GetField(field)
		em.RemoveField(field)
		em.AddField(newfield, val)
		em.AddTag(newtag, field)
	}
}

func (p *Expr) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		emetric := ExprMetric{metric}
		env := map[string]interface{}{
			"Print":              log.Println,
			"SetField":           emetric.SetField,
			"SetTag":             emetric.SetTag,
			"SeparateFieldToTag": emetric.SeparateFieldToTag,
		}

		for _, expression := range p.Expressions {
			parsed, err := expr.Parse(expression)
			if err != nil {
				log.Println("ERROR [expr.Parse]:", err)
				continue
			}

			_, err = expr.Run(parsed, env)
			if err != nil {
				log.Println("ERROR [expr.Run]:", err)
				continue
			}
		}
	}
	return in
}

func init() {
	processors.Add("expr", func() telegraf.Processor {
		return &Expr{}
	})
}
