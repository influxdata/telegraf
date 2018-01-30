package inputs

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/rancherio/go-rancher/v3"
	"os"
	"regexp"
)

var err error

const (
	sampleConfig = `
  ## You can skip the client setup portion of this config if one of two conditions are met:
  ## One, you set the following environemnt variables manually: CATTLE_URL, CATTLE_ACCESS_KEY, CATTLE_SECRET_KEY on your host or in a container
  ## Two, you set 'io.rancher.container.create_agent: true' and 'io.rancher.container.agent.role: environment' labels and run the container
  ## in a rancher environment. This will create a service account for the container and eliminate the need for managing the API keys.
  ## Very important note is that using these labels and not passing an account API creds will only gather information for the environment
  ## this container is deployed in.

  ## Specify the rancher Api Url. This can also be auto detected from  the env variable CATTLE_URL.
  api_url = "http://rancher-host:8080/v3"

  ## The api access key for the rancher API. This can also be extracted from CATTLE_ACCESS_KEY
  api_access_key = ""

  ## The api secret key for the rancher API. This can also be extracted from CATTLE_SECRET_KEY
  api_secret_key = ""

  ## Set host to true when you want to also obtain host state stats
  host_data = true

  ## Set stack_data to true when you want to also obtain stack state stats
  stack_data = true

  ## Set service_data to true when you want to also obtain service state stats
  service_data = true
`
)

// Rancher plugin that pulls Host/Stack/Service data
type Rancher struct {
	client      *client.RancherClient
	ApiUrl      string `toml:"api_url"`
	AccessKey   string `toml:"api_access_key"`
	SecretKey   string `toml:"api_secret_key"`
	HostData    bool
	StackData   bool
	ServiceData bool
	clusterName string
}

func init() {
	inputs.Add("rancher", func() telegraf.Input { return &Rancher{} })
}

// Description will appear directly above the plugin definition in the config file
func (r *Rancher) Description() string {
	return `This plugin gets Host, Stack, and Service data from the rancher api`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (r *Rancher) SampleConfig() string {
	return sampleConfig
}

// This checks if the rancher API env variables are set
// An example of when this would be set is when telegraf is ran on a rancher env in a container with the
// labels 'io.rancher.container.create_agent: true' and 'io.rancher.container.agent.role: environment'
// You can also just set these in the container ENV variables when you deploy the container
func (r *Rancher) checkEnvVariables() {
	if os.Getenv("CATTLE_URL") != "" {
		r.ApiUrl = os.Getenv("CATTLE_URL")
	}

	if os.Getenv("CATTLE_ACCESS_KEY") != "" {
		r.AccessKey = os.Getenv("CATTLE_ACCESS_KEY")
	}

	if os.Getenv("CATTLE_SECRET_KEY") != "" {
		r.SecretKey = os.Getenv("CATTLE_SECRET_KEY")
	}
}

// Create the api client connection to rancher
func (r *Rancher) createAPIClient() (*client.RancherClient, error) {
	// Check to se if env variables are set and set r.ApiUrl, r.AccessKey, r.SecretKey
	r.checkEnvVariables()

	// Build the client conf
	conf := &client.ClientOpts{Url: r.ApiUrl,
		AccessKey: r.AccessKey,
		SecretKey: r.SecretKey}

	c, err := client.NewRancherClient(conf)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Determine the int value of the state fields
// Making it easier to graph state
func setStateValue(stateField string) int {

	// Using regex so it is easy to add additional keywords for each
	ok, _ := regexp.Compile("(active|healthy|running)")
	warn, _ := regexp.Compile("(reconnecting|unhealthy|paused)")
	errState, _ := regexp.Compile("(deactivated|inactive|stopped)")

	var state int

	if ok.MatchString(stateField) {
		state = 0
	}

	if warn.MatchString(stateField) {
		state = 1
	}

	if errState.MatchString(stateField) {
		state = 2
	}

	return state
}

// Gather all the host data from the API
func (r *Rancher) gatherHostData(acc telegraf.Accumulator) error {
	hosts, err := r.client.Host.List(nil)

	if err != nil {
		return err
	}

	for _, host := range hosts.Data {

		if err := r.getClusterName(host.ClusterId); err != nil {
			return err
		}

		hostFields := map[string]interface{}{
			"state":      setStateValue(host.State),
			"containers": len(host.InstanceIds),
			"agentState": setStateValue(host.AgentState),
		}

		hostTags := map[string]string{
			"id":          host.Id,
			"hostname":    host.Hostname,
			"agentState":  host.AgentState,
			"agentId":     host.AgentId,
			"agentIp":     host.AgentIpAddress,
			"clusterId":   host.ClusterId,
			"clusterName": r.clusterName,
			"state":       host.State,
			"name":        host.Name,
		}

		acc.AddFields("rancher_host", hostFields, hostTags)
	}

	return nil
}

// gather the Stack information from the API
func (r *Rancher) gatherStackData(acc telegraf.Accumulator) error {
	stacks, err := r.client.Stack.List(nil)

	if err != nil {
		return err
	}

	for _, stack := range stacks.Data {

		if err := r.getClusterName(stack.ClusterId); err != nil {
			return err
		}

		stackFields := map[string]interface{}{
			"state":       setStateValue(stack.State),
			"healthState": setStateValue(stack.HealthState),
			"containers":  len(stack.ServiceIds),
		}

		stackTags := map[string]string{
			"id":          stack.Id,
			"state":       stack.State,
			"healthState": stack.HealthState,
			"name":        stack.Name,
			"clusterId":   stack.ClusterId,
		}

		acc.AddFields("rancher_stack", stackFields, stackTags)
	}
	return nil
}

// Gather service data information
func (r *Rancher) gatherServiceData(acc telegraf.Accumulator) error {
	services, err := r.client.Service.List(nil)

	if err != nil {
		return err
	}

	for _, service := range services.Data {

		if err := r.getClusterName(service.ClusterId); err != nil {
			return err
		}

		serviceFileds := map[string]interface{}{
			"state":        setStateValue(service.State),
			"containers":   len(service.InstanceIds),
			"scale":        service.Scale,
			"currentScale": service.CurrentScale,
		}

		serviceTags := map[string]string{
			"id":          service.Id,
			"state":       service.State,
			"name":        service.Name,
			"clusterId":   service.ClusterId,
			"stackId":     service.StackId,
			"clusterName": r.clusterName,
		}

		acc.AddFields("rancher_service", serviceFileds, serviceTags)

	}

	return nil
}

func (r *Rancher) getClusterName(clusterId string) error {
	name, err := r.client.Cluster.ById(clusterId)

	if err != nil {
		return err
	}

	r.clusterName = name.Name

	return nil
}

// Establish the connection to the rancher client
func (r *Rancher) GetApiConn() error {
	conf := &client.ClientOpts{Url: r.ApiUrl,
		AccessKey: r.AccessKey,
		SecretKey: r.SecretKey}

	if r.client, err = client.NewRancherClient(conf); err != nil {
		return err
	}

	return nil
}

// Gather defines what data the plugin will gather.
func (r *Rancher) Gather(acc telegraf.Accumulator) error {

	// Set rancher config if env variables set
	r.checkEnvVariables()

	// Establish the connection
	err = r.GetApiConn()

	if err != nil {
		acc.AddError(fmt.Errorf("error getting connection to rancher API: %s", err))
		return err
	}

	if r.HostData {
		if err = r.gatherHostData(acc); err != nil {
			acc.AddError(fmt.Errorf("error getting host data from rancher: %s", err))
			return err
		}
	}

	if r.StackData {
		if err := r.gatherStackData(acc); err != nil {
			acc.AddError(fmt.Errorf("error getting stack data from rancher: %s", err))
			return err
		}
	}

	if r.ServiceData {
		if err = r.gatherServiceData(acc); err != nil {
			acc.AddError(fmt.Errorf("error getting service data from rancher: %s", err))
			return err
		}
	}

	return nil
}
