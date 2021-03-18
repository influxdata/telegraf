package supervisor

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/kolo/xmlrpc"
	"net/url"
)

type Supervisor struct {
	Log telegraf.Logger `toml:"-"`

	Server         string `toml:"url"`
	PidGather      bool   `toml:"gather_pid"`
	ExitCodeGather bool   `toml:"gather_exit_code"`
	UseIdentTag    bool   `toml:"use_identification_tag"`

	Status SupervisorInfo
}

type ProcessInfo struct {
	Name          string `xmlrpc:"name"`
	Group         string `xmlrpc:"group"`
	Description   string `xmlrpc:"description"`
	Start         int32  `xmlrpc:"start"`
	Stop          int32  `xmlrpc:"stop"`
	Now           int32  `xmlrpc:"now"`
	State         int16  `xmlrpc:"state"`
	Statename     string `xmlrpc:"statename"`
	StdoutLogfile string `xmlrpc:"stdout_logfile"`
	StderrLogfile string `xmlrpc:"stderr_logfile"`
	SpawnErr      string `xmlrpc:"spawnerr"`
	ExitStatus    int8   `xmlrpc:"exitstatus"`
	Pid           int32  `xmlrpc:"pid"`
}

type SupervisorInfo struct {
	StateCode int8   `xmlrpc:"statecode"`
	StateName string `xmlrpc:"statename"`
	Ident     string
}

const sampleConfig = `
  ## Url of supervisor's XML-RPC endpoint
  # url="http://localhost:9001/RPC2"
  ## Use supervisor identification string as server tag
  use_identification_tag = false
  ## Gather PID of running processes
  gather_pid = false
  ## Gather exit codes of processes
  gather_exit_code = false
`

func (s *Supervisor) Description() string {
	return "Gather info about processes state, that running under supervisor using its XML-RPC API"
}

func (s *Supervisor) SampleConfig() string {
	return sampleConfig
}

func (s *Supervisor) Gather(acc telegraf.Accumulator) error {
	//Initializing XML-RPC client
	client, err := xmlrpc.NewClient(s.Server, nil)
	if err != nil {
		return err
	}
	defer client.Close()

	//API call to get information about all running processes
	var rawProcessData []ProcessInfo
	err = client.Call("supervisor.getAllProcessInfo", nil, &rawProcessData)
	if err != nil {
		return err
	}

	//API call to get information about instance status
	err = client.Call("supervisor.getState", nil, &s.Status)
	if err != nil {
		return err
	}

	//API call to get identification string
	err = client.Call("supervisor.getIdentification", nil, &s.Status.Ident)
	if err != nil {
		return err
	}

	//Iterating through array of structs with processes info and adding fields to accumulator
	for _, process := range rawProcessData {
		processTags, processFields, err := s.parseProcessData(process)
		if err != nil {
			return err
		}
		acc.AddFields("supervisor_processes", processFields, processTags)
	}
	// Adding instance info fields to accumulator
	instanceTags, instanceFields, err := s.parseInstanceData()
	if err != nil {
		return err
	}
	acc.AddFields("supervisor_instance", instanceFields, instanceTags)
	return nil
}

func (s *Supervisor) parseProcessData(pInfo ProcessInfo) (map[string]string, map[string]interface{}, error) {
	var err error
	tags := map[string]string{
		"process": pInfo.Name,
		"group":   pInfo.Group,
	}
	fields := map[string]interface{}{
		"uptime": pInfo.Now - pInfo.Start,
		"state":  pInfo.State,
	}
	if s.PidGather {
		fields["pid"] = pInfo.Pid
	}
	if s.ExitCodeGather {
		fields["exitCode"] = pInfo.ExitStatus
	}
	if s.UseIdentTag {
		tags["server"] = s.Status.Ident
	} else {
		tags["server"], err = beautifyServerString(s.Server)
		if err != nil {
			return map[string]string{}, map[string]interface{}{}, err
		}
	}
	return tags, fields, nil
}

func (s *Supervisor) parseInstanceData() (map[string]string, map[string]interface{}, error) {
	var err error
	tags := make(map[string]string, 1)
	fields := make(map[string]interface{}, 1)
	if s.UseIdentTag {
		tags["server"] = s.Status.Ident
	} else {
		tags["server"], err = beautifyServerString(s.Server)
		if err != nil {
			return map[string]string{}, map[string]interface{}{}, err
		}
	}
	fields["state"] = s.Status.StateCode
	return tags, fields, nil
}

func (s *Supervisor) Init() error {
	if s.Server == "" {
		s.Server = "http://localhost:9001/RPC2"
	}
	return nil
}

func init() {
	inputs.Add("supervisor", func() telegraf.Input {
		return &Supervisor{}
	})
}

//Function to get only address and port from URL
func beautifyServerString(rawurl string) (string, error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return parsedUrl.Host, nil
}
