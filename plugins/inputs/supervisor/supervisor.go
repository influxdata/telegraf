package supervisor

import (
	"github.com/fatih/structs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/kolo/xmlrpc"
)

type Supervisor struct {
	Host string `toml:"host"`
}

type ProcessInfo struct {
	Name          string `xmlrpc:"name"`
	Group         string `xmlrpc:"group"`
	Description   string `xmlrpc:"description"`
	Start         int64  `xmlrpc:"start"`
	Stop          int64  `xmlrpc:"stop"`
	Now           int64  `xmlrpc:"now"`
	State         int64  `xmlrpc:"state"`
	Statename     string `xmlrpc:"statename"`
	StdoutLogfile string `xmlrpc:"stdout_logfile"`
	StderrLogfile string `xmlrpc:"stderr_logfile"`
	SpawnErr      string `xmlrpc:"spawnerr"`
	ExitStatus    int64  `xmlrpc:"exitstatus"`
	Pid           int64  `xmlrpc:"pid"`
}

func (s *Supervisor) Description() string {
	return "Read supervisor's stats from server"
}

func (s *Supervisor) SampleConfig() string {
	return `
  ## Works with supervisor's XML-RPC API
  ## API needs to be enabled in supervisor's config
  ## Host from which to read supervisor stats:
  host = "http://localhost:9001/RPC2"
`
}

func (s *Supervisor) Gather(acc telegraf.Accumulator) error {
	if s.Host == "" {
		s.Host = "http://localhost:9001/RPC2"
	}

	client, err := xmlrpc.NewClient(s.Host, nil)

	if err != nil {
		return err
	}

	defer client.Close()

	var processes []ProcessInfo
	if err = client.Call("supervisor.getAllProcessInfo", nil, &processes); err != nil {
		return err
	}

	for _, process := range processes {
		tags := map[string]string{"server": s.Host, "process": process.Name}
		fields := structs.Map(process)
		acc.AddFields("supervisor", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("supervisor", func() telegraf.Input {
		return &Supervisor{}
	})
}
