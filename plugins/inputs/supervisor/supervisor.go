package supervisor

import (
	_ "embed"
	"fmt"
	"net/url"

	"github.com/kolo/xmlrpc"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Supervisor struct {
	Server     string          `toml:"url"`
	ServerTag  string          `toml:"server_tag"`
	MetricsInc []string        `toml:"metrics_include"`
	MetricsExc []string        `toml:"metrics_exclude"`
	Log        telegraf.Logger `toml:"-"`

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

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

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
	switch s.ServerTag {
	case "instance":
		tags["server"] = status.Ident
	case "host":
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
	var server string
	var err error
	switch s.ServerTag {
	case "instance":
		server = status.Ident
	case "host":
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
	// Checking validity of server_tag setting
	if !(s.ServerTag == "host" || s.ServerTag == "instance" || s.ServerTag == "none") {
		return fmt.Errorf("unknown value of server_tag in plugin configuration (%s)", s.ServerTag)
	}
	return nil
}

func init() {
	inputs.Add("supervisor", func() telegraf.Input {
		return &Supervisor{
			MetricsExc: []string{"pid", "rc"},
			ServerTag:  "none",
		}
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
