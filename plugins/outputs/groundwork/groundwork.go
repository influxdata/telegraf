package groundwork

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gwos/tcg/sdk/clients"
	"github.com/gwos/tcg/sdk/transit"
	"github.com/hashicorp/go-uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultMonitoringRoute = "/api/monitoring?dynamic=true"
)

// Login and logout routes from Groundwork API
const (
	loginURL  = "/api/auth/login"
	logoutURL = "/api/auth/logout"
)

const (
	sampleConfig = `
  ## URL of your groundwork instance.
  url = "https://groundwork.example.com"

  ## Agent uuid for GroundWork API Server.
  agent_id = ""

  ## Username and password to access GroundWork API.
  username = ""
  password = ""

  ## Default display name for the host with services(metrics).
  # default_host = "telegraf"

  ## Default service state.
  # default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname.
  # resource_tag = "host"
`
)

type Groundwork struct {
	Server              string          `toml:"url"`
	AgentID             string          `toml:"agent_id"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	DefaultHost         string          `toml:"default_host"`
	DefaultServiceState string          `toml:"default_service_state"`
	ResourceTag         string          `toml:"resource_tag"`
	Log                 telegraf.Logger `toml:"-"`

	authToken string
}

func (g *Groundwork) SampleConfig() string {
	return sampleConfig
}

func (g *Groundwork) Init() error {
	if g.Server == "" {
		return errors.New("no 'url' provided")
	}
	if g.AgentID == "" {
		return errors.New("no 'agent_id' provided")
	}
	if g.Username == "" {
		return errors.New("no 'username' provided")
	}
	if g.Password == "" {
		return errors.New("no 'password' provided")
	}
	if g.DefaultHost == "" {
		return errors.New("no 'default_host' provided")
	}
	if g.ResourceTag == "" {
		return errors.New("no 'resource_tag' provided")
	}
	if !validStatus(g.DefaultServiceState) {
		return errors.New("invalid 'default_service_state' provided")
	}

	return nil
}

func (g *Groundwork) Connect() error {
	byteToken, err := login(g.Server+loginURL, g.Username, g.Password)
	if err != nil {
		return fmt.Errorf("could not log in at %s: %v", g.Server+loginURL, err)
	}

	g.authToken = string(byteToken)

	return nil
}

func (g *Groundwork) Close() error {
	formValues := map[string]string{
		"gwos-app-name":  "telegraf",
		"gwos-api-token": g.authToken,
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   internal.ProductToken(),
	}

	if _, _, err := clients.SendRequest(http.MethodPost, g.Server+logoutURL, headers, formValues, nil); err != nil {
		return fmt.Errorf("could not log out at %s: %v", g.Server+logoutURL, err)
	}

	return nil
}

func (g *Groundwork) Write(metrics []telegraf.Metric) error {
	resourceToServicesMap := make(map[string][]transit.DynamicMonitoredService)
	for _, metric := range metrics {
		resource, service := g.parseMetric(metric)
		resourceToServicesMap[resource] = append(resourceToServicesMap[resource], service)
	}

	var resources []transit.DynamicMonitoredResource
	for resourceName, services := range resourceToServicesMap {
		resources = append(resources, transit.DynamicMonitoredResource{
			BaseResource: transit.BaseResource{
				BaseTransitData: transit.BaseTransitData{
					Name: resourceName,
					Type: transit.Host,
				},
			},
			Status:        transit.HostUp,
			LastCheckTime: transit.NewTimestamp(),
			Services:      services,
		})
	}

	traceToken, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	requestJSON, err := json.Marshal(transit.DynamicResourcesWithServicesRequest{
		Context: &transit.TracerContext{
			AppType:    "TELEGRAF",
			AgentID:    g.AgentID,
			TraceToken: traceToken,
			TimeStamp:  transit.NewTimestamp(),
			Version:    transit.ModelVersion,
		},
		Resources: resources,
		Groups:    nil,
	})

	if err != nil {
		return err
	}

	headers := map[string]string{
		"GWOS-APP-NAME":  "telegraf",
		"GWOS-API-TOKEN": g.authToken,
		"Content-Type":   "application/json",
		"Accept":         "application/json",
		"User-Agent":     internal.ProductToken(),
	}

	statusCode, body, err := clients.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJSON)
	if err != nil {
		return fmt.Errorf("error while sending: %v", err)
	}

	/* Re-login mechanism */
	if statusCode == 401 {
		if err = g.Connect(); err != nil {
			return fmt.Errorf("re-login failed: %v", err)
		}
		headers["GWOS-API-TOKEN"] = g.authToken
		statusCode, body, err = clients.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJSON)
		if err != nil {
			return fmt.Errorf("error while sending: %v", err)
		}
	}

	if statusCode != 200 {
		return fmt.Errorf("HTTP request failed. [Status code]: %d, [Response]: %s", statusCode, string(body))
	}

	return nil
}

func (g *Groundwork) Description() string {
	return "Send telegraf metrics to GroundWork Monitor"
}

func init() {
	outputs.Add("groundwork", func() telegraf.Output {
		return &Groundwork{
			ResourceTag:         "host",
			DefaultHost:         "telegraf",
			DefaultServiceState: string(transit.ServiceOk),
		}
	})
}

func (g *Groundwork) parseMetric(metric telegraf.Metric) (string, transit.DynamicMonitoredService) {
	resource := g.DefaultHost
	if value, present := metric.GetTag(g.ResourceTag); present {
		resource = value
	}

	service := metric.Name()
	if value, present := metric.GetTag("service"); present {
		service = value
	}

	status := g.DefaultServiceState
	value, statusPresent := metric.GetTag("status")
	if validStatus(value) {
		status = value
	}

	message, _ := metric.GetTag("message")

	unitType := string(transit.UnitCounter)
	if value, present := metric.GetTag("unitType"); present {
		unitType = value
	}

	critical := -1.0
	value, criticalPresent := metric.GetTag("critical")
	if criticalPresent {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			critical = s
		}
	}

	warning := -1.0
	value, warningPresent := metric.GetTag("warning")
	if warningPresent {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			warning = s
		}
	}

	lastCheckTime := transit.NewTimestamp()
	lastCheckTime.Time = metric.Time()
	serviceObject := transit.DynamicMonitoredService{
		BaseTransitData: transit.BaseTransitData{
			Name:  service,
			Type:  transit.Service,
			Owner: resource,
		},
		Status:           transit.MonitorStatus(status),
		LastCheckTime:    lastCheckTime,
		LastPlugInOutput: message,
		Metrics:          nil,
	}

	for _, value := range metric.FieldList() {
		var thresholds []transit.ThresholdValue
		if warningPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Warning,
				Label:      value.Key + "_wn",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: warning,
				},
			})
		}
		if criticalPresent {
			thresholds = append(thresholds, transit.ThresholdValue{
				SampleType: transit.Critical,
				Label:      value.Key + "_cr",
				Value: &transit.TypedValue{
					ValueType:   transit.DoubleType,
					DoubleValue: critical,
				},
			})
		}

		typedValue := new(transit.TypedValue)
		err := typedValue.FromInterface(value.Value)
		if err != nil {
			typedValue = nil
			g.Log.Errorf("%v", err)
		}

		serviceObject.Metrics = append(serviceObject.Metrics, transit.TimeSeries{
			MetricName: value.Key,
			SampleType: transit.Value,
			Interval: &transit.TimeInterval{
				EndTime: lastCheckTime,
			},
			Value:      typedValue,
			Unit:       transit.UnitType(unitType),
			Thresholds: &thresholds,
		})
	}

	if !statusPresent {
		serviceStatus, err := transit.CalculateServiceStatus(&serviceObject.Metrics)
		if err != nil {
			g.Log.Infof("could not calculate service status, reverting to default_service_state: %v", err)
			serviceObject.Status = transit.MonitorStatus(g.DefaultServiceState)
		}
		serviceObject.Status = serviceStatus
	}

	return resource, serviceObject
}

func login(url, username, password string) ([]byte, error) {
	formValues := map[string]string{
		"user":          username,
		"password":      password,
		"gwos-app-name": "telegraf",
	}
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "text/plain",
		"User-Agent":   internal.ProductToken(),
	}

	statusCode, body, err := clients.SendRequest(http.MethodPost, url, headers, formValues, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("request failed with status-code %d: %v", statusCode, string(body))
	}

	return body, nil
}

func validStatus(status string) bool {
	switch transit.MonitorStatus(status) {
	case transit.ServiceOk, transit.ServiceWarning, transit.ServicePending, transit.ServiceScheduledCritical,
		transit.ServiceUnscheduledCritical, transit.ServiceUnknown:
		return true
	}
	return false
}
