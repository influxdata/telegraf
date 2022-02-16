package supervisor

import (
	"fmt"
	"net/url"

	"github.com/kolo/xmlrpc"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Supervisor struct {
	Server      string          `toml:"url"`
	UseIdentTag bool            `toml:"use_identification_tag"`
	MetricsInc  []string        `toml:"metrics_include"`
	MetricsExc  []string        `toml:"metrics_exclude"`
	Log         telegraf.Logger `toml:"-"`

	rpcClient   *xmlrpc.Client
	fieldFilter filter.Filter
}

type processInfo struct {
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

type supervisorInfo struct {
	StateCode int8   `xmlrpc:"statecode"`
	StateName string `xmlrpc:"statename"`
	Ident     string
}

const sampleConfig = `
  ## Url of supervisor's XML-RPC endpoint if basic auth enabled in supervisor http server,
  ## than you have to add credentials to url (ex. http://login:pass@localhost:9001/RPC2)
  # url="http://localhost:9001/RPC2"
  ## Use supervisor identification string as server tag
  # use_identification_tag = false
  ## With settings below you can manage gathering additional information about processes
  ## If both of them empty, then all additional information will be collected.
  ## Currently supported supported additional metrics are: pid, rc
  # metrics_include = []
  # metrics_exclude = ["pid", "rc"]
`

func (s *Supervisor) Description() string {
	return "Gather info about processes state, that running under supervisor using its XML-RPC API"
}

func (s *Supervisor) SampleConfig() string {
	return sampleConfig
}

func (s *Supervisor) Gather(acc telegraf.Accumulator) error {
	// API call to get information about all running processes
	var rawProcessData []processInfo
	err := s.rpcClient.Call("supervisor.getAllProcessInfo", nil, &rawProcessData)
	if err != nil {
		return fmt.Errorf("failed to get processes info: %v", err)
	}

	// API call to get information about instance status
	var status supervisorInfo
	err = s.rpcClient.Call("supervisor.getState", nil, &status)
	if err != nil {
		return fmt.Errorf("failed to get processes info: %v", err)
	}

	// API call to get identification string
	err = s.rpcClient.Call("supervisor.getIdentification", nil, &status.Ident)
	if err != nil {
		return fmt.Errorf("failed to get instance identification: %v", err)
	}

	// Iterating through array of structs with processes info and adding fields to accumulator
	for _, process := range rawProcessData {
		processTags, processFields, err := s.parseProcessData(process, status)
		if err != nil {
			acc.AddError(err)
			continue
		}
		acc.AddFields("supervisor_processes", processFields, processTags)
	}
	// Adding instance info fields to accumulator
	instanceTags, instanceFields, err := s.parseInstanceData(status)
	if err != nil {
		return fmt.Errorf("failed to parse instance data: %v", err)
	}
	acc.AddFields("supervisor_instance", instanceFields, instanceTags)
	return nil
}

func (s *Supervisor) parseProcessData(pInfo processInfo, status supervisorInfo) (map[string]string, map[string]interface{}, error) {
	tags := map[string]string{
		"process": pInfo.Name,
		"group":   pInfo.Group,
	}
	fields := map[string]interface{}{
		"uptime": pInfo.Now - pInfo.Start,
		"state":  pInfo.State,
	}
	if s.fieldFilter.Match("pid") {
		fields["pid"] = pInfo.Pid
	}
	if s.fieldFilter.Match("rc") {
		fields["exitCode"] = pInfo.ExitStatus
	}
	if s.UseIdentTag {
		tags["server"] = status.Ident
	} else {
		var err error
		tags["server"], err = beautifyServerString(s.Server)
		if err != nil {
			return map[string]string{}, map[string]interface{}{}, err
		}
	}
	return tags, fields, nil
}

// Parsing of supervisor instance data
func (s *Supervisor) parseInstanceData(status supervisorInfo) (map[string]string, map[string]interface{}, error) {
	server := status.Ident
	// Using server URL for server tag instead of instance identification, if plugin configured accordingly
	if !s.UseIdentTag {
		var err error
		server, err = beautifyServerString(s.Server)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decorate server string: %v", err)
		}
	}
	tags := map[string]string{"server": server}
	fields := map[string]interface{}{"state": status.StateCode}
	return tags, fields, nil
}

func (s *Supervisor) Init() error {
	// Using default server URL if none was specified in config
	if s.Server == "" {
		s.Server = "http://localhost:9001/RPC2"
	}
	var err error
	// Initializing XML-RPC client
	s.rpcClient, err = xmlrpc.NewClient(s.Server, nil)
	if err != nil {
		return fmt.Errorf("XML-RPC client initialization failed: %v", err)
	}
	// Setting filter for additional metrics
	s.fieldFilter, err = filter.NewIncludeExcludeFilter(s.MetricsInc, s.MetricsExc)
	if err != nil {
		return fmt.Errorf("metrics filter setup failed: %v", err)
	}
	return nil
}

func init() {
	inputs.Add("supervisor", func() telegraf.Input {
		return &Supervisor{MetricsExc: []string{"pid", "rc"}}
	})
}

// Function to get only address and port from URL
func beautifyServerString(rawurl string) (string, error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}
