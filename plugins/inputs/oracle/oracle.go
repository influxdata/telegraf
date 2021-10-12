package oracle

import (
	"fmt"
	"os"
	osExec "os/exec"
	"time"

	_ "embed" //nolint // golangci-lint@1.38.0 incorrect false positive

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/exec"
	"github.com/influxdata/telegraf/plugins/parsers"
)

//go:embed oracle_metrics.py
var pythonScript []byte

const sampleConfig = `
  ## Database user with SELECT_CATALOG_ROLE role granted, required.
  username = system
  password = oracle
  ## Data source name, required. 
	## See https://cx-oracle.readthedocs.io/en/latest/user_guide/connection_handling.html#connection-strings
  dsn = XE

  ## python executable, python3 by default 
  python=python3

  ## Timeout for metrics collector to complete.
  timeout = "5s"
`

type Oracle struct {
	Python  string          `toml:"python"`
	Env     []string        `toml:"env"`
	Timeout config.Duration `toml:"timeout"`

	Username string `toml:"username"`
	Password string `toml:"password"`
	DSN      string `toml:"dsn"`

	parser    parsers.Parser
	args      []string
	scriptEnv []string

	runner exec.Runner
}

func NewOracle() *Oracle {
	parser, _ := parsers.NewInfluxParser()
	return &Oracle{
		Python:  "python3",
		Timeout: config.Duration(time.Second * 5),
		runner:  exec.NewRunner(),
		parser:  parser,
	}
}

func (*Oracle) SampleConfig() string {
	return sampleConfig
}

func (*Oracle) Description() string {
	return "Read metrics from Oracle RDBMS"
}

func (o *Oracle) Gather(acc telegraf.Accumulator) error {
	out, errbuf, runErr := o.runner.Run(o.Python, o.args, o.scriptEnv, pythonScript, time.Duration(o.Timeout))
	if runErr != nil {
		err := fmt.Errorf("oracle: %s : %s", runErr, string(errbuf))
		acc.AddError(err)
		return nil
	}

	metrics, err := o.parser.Parse(out)
	if err != nil {
		acc.AddError(err)
		return nil
	}

	for _, m := range metrics {
		acc.AddMetric(m)
	}

	return nil
}

func (o *Oracle) Init() error {
	// validate username, password and dsn
	if o.Username == "" {
		return fmt.Errorf(`oracle: "username" is required`)
	}
	if o.Password == "" {
		return fmt.Errorf(`oracle: "password" is required`)
	}
	if o.DSN == "" {
		return fmt.Errorf(`oracle: "dsn" is required`)
	}
	if o.Env != nil {
		o.scriptEnv = append(os.Environ(), o.Env...)
	}
	o.args = []string{
		"-",
		"-u", o.Username,
		"-p", o.Password,
		"-d", o.DSN,
	}

	// validate that python executable exists
	if _, err := osExec.LookPath(o.Python); err != nil {
		return fmt.Errorf("oracle: python executable not found: %v", err)
	}

	return nil
}

func init() {
	inputs.Add("oracle", func() telegraf.Input {
		return NewOracle()
	})
}
