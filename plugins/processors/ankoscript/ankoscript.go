package ankoscript

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/mattn/anko/core"
	"github.com/mattn/anko/vm"
)

var sampleConfig = `

`

type AnkoScript struct {
	Script string

	env *vm.Env
}

func (this *AnkoScript) SampleConfig() string {
	return sampleConfig
}

func (this *AnkoScript) Description() string {
	return "Run anko code against metrics"
}

func (this *AnkoScript) Apply(in ...telegraf.Metric) []telegraf.Metric {
	var err error
	for _, metric := range in {
		err = this.env.Define("metric", metric)
		if err != nil {
			log.Printf("E! [Define]: %s", err)
		}

		_, err = this.env.Execute(this.Script)
		if err != nil {
			log.Printf("E! [Execute error]: %v", err)
		}
		this.env.Delete("metric")
	}
	return in
}

func String(i interface{}) string {
	return i.(string)
}

func init() {
	processors.Add("ankoscript", func() telegraf.Processor {
		anko := &AnkoScript{
			env: vm.NewEnv(),
		}

		err := anko.env.Define("log", log.Println)
		if err != nil {
			log.Printf("E! [Define]: %s", err)
		}

		err = anko.env.Define("println", fmt.Println)
		if err != nil {
			log.Printf("E! [Define]: %s", err)
		}

		err = anko.env.Define("sprintf", fmt.Sprintf)
		if err != nil {
			log.Printf("E! [Define]: %s", err)
		}

		err = anko.env.Define("String", String)
		if err != nil {
			log.Printf("E! [Define]: %s", err)
		}
		core.Import(anko.env)
		//packages.DefineImport(env)
		return anko
	})
}
