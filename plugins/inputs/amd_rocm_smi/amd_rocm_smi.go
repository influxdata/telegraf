package amd_rocm_smi

import (
	_ "embed" // Required for embedding the parser config file
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/toml"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/transport"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	generic "github.com/influxdata/telegraf/plugins/common/receive_parse"
)

//go:embed amd_rocm_smi_parser.conf
var cfgfile []byte

func NewAMDSMI() *generic.ReceiveAndParse {
	var cfg parsers.Config
	if err := toml.Unmarshal(cfgfile, &cfg); err != nil {
		panic(fmt.Errorf("cannot unmarshal 'amd_rocm_smi_parser.conf': %v", err))
	}
	parser, err := parsers.NewParser(&cfg)
	if err != nil {
		panic(fmt.Errorf("cannot instantiate parser for 'amd_rocm_smi': %v", err))
	}

	return &generic.ReceiveAndParse{
		DescriptionText: "Query statistics from AMD Graphics cards using rocm-smi binary",
		Receiver: &transport.Exec{
			BinPath: "/opt/rocm/bin/rocm-smi",
			Timeout: config.Duration(5 * time.Second),
			BinArgs: []string{
				"-o",
				"-l",
				"-m",
				"-M",
				"-g",
				"-c",
				"-t",
				"-u",
				"-i",
				"-f",
				"-p",
				"-P",
				"-s",
				"-S",
				"-v",
				"--showreplaycount",
				"--showpids",
				"--showdriverversion",
				"--showmemvendor",
				"--showfwinfo",
				"--showproductname",
				"--showserial",
				"--showuniqueid",
				"--showbus",
				"--showpendingpages",
				"--showpagesinfo",
				"--showmeminfo",
				"all",
				"--showretiredpages",
				"--showunreservablepages",
				"--showmemuse",
				"--showvoltage",
				"--showtopo",
				"--showtopoweight",
				"--showtopohops",
				"--showtopotype",
				"--showtoponuma",
				"--json"},
		},
		Parser: parser,
		PostProcessors: []generic.PostProcessor{
			{
				Name:    "memory_free computation",
				Process: postProcessMemoryFree,
			},
			{
				Name:    "driver_version conversion",
				Process: postProcessDriverVersion,
			},
		},
	}
}

func postProcessMemoryFree(m telegraf.Metric) error {
	fields := m.Fields()

	iTotal, found := fields["memory_total"]
	if !found {
		return fmt.Errorf("memory_total missing")
	}
	total, ok := iTotal.(int64)
	if !ok {
		return fmt.Errorf("memory_total is not int64 but %T", iTotal)
	}

	iUsed, found := fields["memory_used"]
	if !found {
		return fmt.Errorf("memory_used missing")
	}
	used, ok := iUsed.(int64)
	if !ok {
		return fmt.Errorf("memory_used is not int64 but %T", iUsed)
	}

	m.AddField("memory_free", total-used)
	return nil
}

func postProcessDriverVersion(m telegraf.Metric) error {
	iVersion, found := m.GetField("driver_version")
	if !found {
		return nil
	}

	sVersion, ok := iVersion.(string)
	if !ok {
		return fmt.Errorf("driver_version is not string but %T", iVersion)
	}

	sVersion = strings.Replace(sVersion, ".", "", -1)
	version, err := strconv.ParseInt(sVersion, 10, 64)
	if err != nil {
		return err
	}

	m.AddField("driver_version", version)
	return nil
}

func init() {
	inputs.Add("amd_rocm_smi", func() telegraf.Input { return NewAMDSMI() })
}
