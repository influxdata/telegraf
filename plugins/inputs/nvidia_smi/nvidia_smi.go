package nvidia_smi

import (
	_ "embed" // Required for embedding the parser config file
	"fmt"
	"time"

	"github.com/influxdata/toml"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/transport"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	generic "github.com/influxdata/telegraf/plugins/common/receive_parse"
)

//go:embed nvidia_smi_parser.conf
var cfgfile []byte

func NewNvidiaSMI() *generic.ReceiveAndParse {
	var cfg parsers.Config
	if err := toml.Unmarshal(cfgfile, &cfg); err != nil {
		panic(fmt.Errorf("cannot unmarshal 'nvidia_smi_parser.conf': %v", err))
	}
	parser, err := parsers.NewParser(&cfg)
	if err != nil {
		panic(fmt.Errorf("cannot instantiate parser for 'nvidia_smi': %v", err))
	}

	return &generic.ReceiveAndParse{
		DescriptionText: "Pulls statistics from nvidia GPUs attached to the host",
		Receiver: &transport.Exec{
			BinPath: "/usr/bin/nvidia-smi",
			Timeout: config.Duration(5 * time.Second),
			BinArgs: []string{"-q", "-x"},
		},
		Parser: parser,
	}
}

func init() {
	inputs.Add("nvidia_smi", func() telegraf.Input { return NewNvidiaSMI() })
}
