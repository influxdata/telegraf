package inputs

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Input
type CreatorCtx func() telegraf.InputCtx

var Inputs = map[string]Creator{}
var InputsCtx = map[string]CreatorCtx{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}

func AddCtx(name string, creatorCtx CreatorCtx) {
	InputsCtx[name] = creatorCtx
}
