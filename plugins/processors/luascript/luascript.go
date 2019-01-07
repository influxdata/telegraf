package luascript

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	lua "github.com/yuin/gopher-lua"
)

var sampleConfig = `

`

const luaMetric = "metric"

type LuaScript struct {
	Script string

	luaVM *lua.LState
}

func (p *LuaScript) SampleConfig() string {
	return sampleConfig
}

func (p *LuaScript) Description() string {
	return "Run LUA code against metrics"
}

func (p *LuaScript) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		lud := &lua.LUserData{
			Value: metric,
		}
		p.luaVM.SetGlobal("metric", lud)
		if err := p.luaVM.DoString(p.Script); err != nil {
			log.Println("metric", metric, "err", err)
		}
	}
	return in
}

func init() {
	processors.Add("luascript", func() telegraf.Processor {
		return &LuaScript{
			luaVM: lua.NewState(),
		}
	})
}
