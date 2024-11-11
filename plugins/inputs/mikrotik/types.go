package mikrotik

type common []commonData

type commonData map[string]string

var tagFields = []string{
	".id",
	"action",
	"chain",
	"comment",
	"connection-state",
	"default-name",
	"disabled",
	"dst-address",
	"dst-port",
	"endpoint-address",
	"in-interface",
	"interface",
	"interface-type",
	"last-ip",
	"mac-address",
	"master-interface",
	"name",
	"out-interface",
	"owner",
	"protocol",
	"running",
	"slave",
	"src-address",
	"src-port",
	"ssid",
	"status",
	"type",
}

var valueFields = []string{
	"bytes",
	"cpu-frequency",
	"cpu-load",
	"distance",
	"fp-rx-byte",
	"fp-rx-packet",
	"fp-tx-byte",
	"fp-tx-packet",
	"frame-bytes",
	"frames",
	"free-hdd-space",
	"free-memory",
	"hw-frame-bytes",
	"hw-frames",
	"last-seen",
	"link-downs",
	"orig-bytes",
	"orig-fasttrack-bytes",
	"orig-fasttrack-packets",
	"orig-packets",
	"orig-rate",
	"packets",
	"repl-bytes",
	"repl-fasttrack-bytes",
	"repl-fasttrack-packets",
	"repl-packets",
	"repl-rate",
	"run-count",
	"rx",
	"rx-byte",
	"rx-drop",
	"rx-error",
	"rx-packet",
	"total-memory",
	"tx",
	"tx-byte",
	"tx-drop",
	"tx-error",
	"tx-frames-timed-out",
	"tx-packet",
	"tx-queue-drop",
	"uptime",
	"write-sect-since-reboot",
	"write-sect-total",
}

var durationParseFieldNames = []string{
	"last-seen",
	"uptime",
}

var systemResources = []string{
	"architecture-name",
	"board-name",
	"cpu",
	"platform",
	"version",
}

var systemRouterBoard = []string{
	"current-firmware",
	"firmware-type",
	"model",
	"serial-number",
}

type parsedPoint struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

type mikrotikEndpoint struct {
	name, url string
}
