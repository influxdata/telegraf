package math

import (
	"fmt"
	"log"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// Math function congfig
type Math struct {
	Func   string   `toml:"function"`
	Metric string   `toml:"measurement_name"`
	Fields []string `toml:"fields"`
}

var mathFunctions = map[string]func(float64) float64{
	"abs":   math.Abs,
	"acos":  math.Acos,
	"acosh": math.Acosh,
	"asin":  math.Asin,
	"asinh": math.Asinh,
	"atan":  math.Atan,
	"atanh": math.Atanh,
	"cbrt":  math.Cbrt,
	"ceil":  math.Ceil,
	"cos":   math.Cos,
	"cosh":  math.Cosh,
	"erf":   math.Erf,
	"erfc":  math.Erfc,
	"exp":   math.Exp,
	"exp2":  math.Exp2,
	"expm1": math.Expm1,
	"floor": math.Floor,
	"gamma": math.Gamma,
	"j0":    math.J0,
	"j1":    math.J1,
	"log":   math.Log,
	"log10": math.Log10,
	"log1p": math.Log1p,
	"log2":  math.Log2,
	"logb":  math.Logb,
	"sin":   math.Sin,
	"sinh":  math.Sinh,
	"sqrt":  math.Sqrt,
	"tan":   math.Tan,
	"tanh":  math.Tanh,
	"trunc": math.Trunc,
	"y0":    math.Y0,
	"y1":    math.Y1,
}

func (p *Math) setFunction(function string) {
	p.Func = function
}

func (p *Math) setMetric(metric string) {
	p.Metric = metric
}

func (p *Math) setFields(fields []string) {
	p.Fields = fields
}

func newMath() telegraf.Processor {
	newmath := &Math{}
	return newmath
}

var sampleConfig = `
## Example config that processe all fields of the metric.
# [[processor.math]]
#   ## Math function
#   function = "abs"
#   ## The name of metric.
#   measurement_name = "cpu"

## Example config that processe only specific fields of the metric.
# [[processor.math]]
#   ## Math function
#   function = "abs"
#   ## The name of metric.
#   measurement_name = "diskio"
#   ## The concrete fields of metric
#   fields = ["io_time", "read_time", "write_time"]
`

func (p *Math) SampleConfig() string {
	return sampleConfig
}

func (p *Math) Description() string {
	return "Math metrics that pass through this filter."
}

func (p *Math) Apply(in ...telegraf.Metric) []telegraf.Metric {

	if _, ok := mathFunctions[p.Func]; ok {
		for _, metric := range in {
			if metric.Name() == p.Metric {
				mathFunc := mathFunctions[p.Func]
				if len(p.Fields) > 0 {
					for _, name := range p.Fields {
						if srcField, ok := metric.Fields()[name]; ok {
							field, ok := srcField.(float64)
							if ok {
								metric.AddField(
									fmt.Sprintf("%s_%s", name, p.Func),
									mathFunc(field))
							}
						} else {
							log.Printf(
								"E! Metric %s doesn't have field %s",
								p.Func,
								metric.Name())
						}
					}
				} else {
					for name, srcField := range metric.Fields() {
						field, ok := srcField.(float64)
						if ok {
							metric.AddField(
								fmt.Sprintf("%s_%s", name, p.Func),
								mathFunc(field))
						}
					}
				}
			}
		}
	} else {
		log.Printf(
			"E! Math function doesn't support %s",
			p.Func)
	}

	return in
}

func init() {
	processors.Add("math", func() telegraf.Processor {
		return newMath()
	})
}
