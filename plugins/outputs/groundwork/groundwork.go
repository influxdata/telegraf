package groundwork

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gwos/tcg/clients"
	"github.com/gwos/tcg/milliseconds"
	"github.com/gwos/tcg/transit"
	"github.com/hashicorp/go-uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultMonitoringRoute = "/api/monitoring?dynamic=true"
)

// Login and logout routes from Groundwork API
const (
	loginURL  = "/api/auth/login"
	logoutURL = "/api/auth/logout"
)

var (
	sampleConfig = `
  ## HTTP endpoint for your groundwork instance.
  endpoint = ""

  ## Agent uuid for Groundwork API Server
  agent_id = ""

  ## Groundwork application type
  app_type = ""

  ## Username to access Groundwork API
  username = ""

  ## Password to use in pair with username
  password = ""

  ## Default display name for the host with services(metrics)
  default_host = "default_telegraf"

  ## Default service state [default - "SERVICE_OK"]
  default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname [default - "host"]
  resource_tag = "host"
`
)

type Groundwork struct {
	Server              string `toml:"groundwork_endpoint"`
	AgentID             string `toml:"agent_id"`
	Username            string `toml:"username"`
	Password            string `toml:"password"`
	DefaultHost         string `toml:"default_host"`
	DefaultServiceState string `toml:"default_service_state"`
	ResourceTag         string `toml:"resource_tag"`
	authToken           string
}

func (g *Groundwork) SampleConfig() string {
	return sampleConfig
}

func (g *Groundwork) Init() error {
	if g.Server == "" {
		return errors.New("no 'groundwork_endpoint' provided")
	}
	if g.Username == "" {
		return errors.New("no 'username' provided")
	}
	if g.Password == "" {
		return errors.New("no 'password' provided")
	}
	if g.DefaultHost == "" {
		g.DefaultHost = "telegraf"
	}
	if g.ResourceTag == "" {
		g.ResourceTag = "host"
	}
	if g.DefaultServiceState == "" || !validStatus(g.DefaultServiceState) {
		g.DefaultServiceState = string(transit.ServiceOk)
	}

	return nil
}

func (g *Groundwork) Connect() error {
	byteToken, err := login(g.Server+loginURL, g.Username, g.Password)
	if err == nil {
		g.authToken = string(byteToken)
	}

	return err
}

func (g *Groundwork) Close() error {
	formValues := map[string]string{
		"gwos-app-name":  "telegraf",
		"gwos-api-token": g.authToken,
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	_, _, err := clients.SendRequest(http.MethodPost, g.Server+logoutURL, headers, formValues, nil)

	return err
}

func (g *Groundwork) Write(metrics []telegraf.Metric) error {
	resourceToServicesMap := make(map[string][]transit.DynamicMonitoredService)
	for _, metric := range metrics {
		resource, service := parseMetric(g.DefaultHost, g.DefaultServiceState, g.ResourceTag, metric)
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
			LastCheckTime: milliseconds.MillisecondTimestamp{Time: time.Now()},
			NextCheckTime: milliseconds.MillisecondTimestamp{Time: time.Now()},
			Services:      services,
		})
	}

	traceToken, _ := uuid.GenerateUUID()
	requestJSON, err := json.Marshal(transit.DynamicResourcesWithServicesRequest{
		Context: &transit.TracerContext{
			AppType:    "TELEGRAF",
			AgentID:    g.AgentID,
			TraceToken: traceToken,
			TimeStamp:  milliseconds.MillisecondTimestamp{Time: time.Now()},
			Version:    transit.ModelVersion,
		},
		Resources: resources,
		Groups:    nil,
	})

	if err != nil {
		return err
	}

	headers := map[string]string{
		"GWOS-APP-NAME":  "groundwork",
		"GWOS-API-TOKEN": g.authToken,
		"Content-Type":   "application/json",
		"Accept":         "application/json",
	}

	statusCode, _, err := clients.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJSON)
	if err != nil {
		return err
	}

	/* Re-login mechanism */
	if statusCode == 401 {
		if err = g.Connect(); err != nil {
			return err
		}
		headers["GWOS-API-TOKEN"] = g.authToken
		statusCode, body, httpErr := clients.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJSON)
		if httpErr != nil {
			return httpErr
		}
		if statusCode != 200 {
			return fmt.Errorf("something went wrong during processing an http request[http_status = %d, body = %s]", statusCode, string(body))
		}
	}

	return nil
}

func (g *Groundwork) Description() string {
	return "Send telegraf metrics to GroundWork Monitor"
}

func init() {
	outputs.Add("groundwork", func() telegraf.Output {
		return &Groundwork{}
	})
}

func parseMetric(defaultHostname, defaultServiceState, resourceTag string, metric telegraf.Metric) (string, transit.DynamicMonitoredService) {
	resource := defaultHostname

	if value, present := metric.GetTag(resourceTag); present {
		resource = value
	}

	service := metric.Name()
	if value, present := metric.GetTag("service"); present {
		service = value
	}

	status := defaultServiceState
	if value, present := metric.GetTag("status"); present {
		if validStatus(value) {
			status = value
		}
	}

	message := ""
	if value, present := metric.GetTag("message"); present {
		message = value
	}

	unitType := string(transit.UnitCounter)
	if value, present := metric.GetTag("unitType"); present {
		unitType = value
	}

	critical := -1.0
	if value, present := metric.GetTag("critical"); present {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			critical = s
		}
	}

	warning := -1.0
	if value, present := metric.GetTag("warning"); present {
		if s, err := strconv.ParseFloat(value, 64); err == nil {
			warning = s
		}
	}

	serviceObject := transit.DynamicMonitoredService{
		BaseTransitData: transit.BaseTransitData{
			Name:  service,
			Type:  transit.Service,
			Owner: resource,
		},
		Status:           transit.MonitorStatus(status),
		LastCheckTime:    milliseconds.MillisecondTimestamp{Time: time.Now()},
		NextCheckTime:    milliseconds.MillisecondTimestamp{},
		LastPlugInOutput: message,
		Metrics:          nil,
	}

	for _, value := range metric.FieldList() {
		var thresholds []transit.ThresholdValue
		thresholds = append(thresholds, transit.ThresholdValue{
			SampleType: transit.Warning,
			Label:      value.Key + "_wn",
			Value: &transit.TypedValue{
				ValueType:   transit.DoubleType,
				DoubleValue: warning,
			},
		})
		thresholds = append(thresholds, transit.ThresholdValue{
			SampleType: transit.Critical,
			Label:      value.Key + "_cr",
			Value: &transit.TypedValue{
				ValueType:   transit.DoubleType,
				DoubleValue: critical,
			},
		})

		valueType := transit.DoubleType
		var floatVal float64
		var stringVal string

		switch value.Value.(type) {
		case string:
			valueType = transit.StringType
			tmpStr := value.Value.(string)
			stringVal = tmpStr
		default:
			floatVal, _ = internal.ToFloat64(value.Value)
		}
		serviceObject.Metrics = append(serviceObject.Metrics, transit.TimeSeries{
			MetricName: value.Key,
			SampleType: transit.Value,
			Interval: &transit.TimeInterval{
				EndTime: milliseconds.MillisecondTimestamp{Time: metric.Time()},
			},
			Value: &transit.TypedValue{
				ValueType:   valueType,
				DoubleValue: floatVal,
				StringValue: stringVal,
			},
			Unit:       transit.UnitType(unitType),
			Thresholds: &thresholds,
		})
	}

	serviceObject.Status, _ = transit.CalculateServiceStatus(&serviceObject.Metrics)

	return resource, serviceObject
}

func login(url, username, password string) ([]byte, error) {
	formValues := map[string]string{
		"user":          username,
		"password":      password,
		"gwos-app-name": "groundwork",
	}
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "text/plain",
	}

	statusCode, body, err := clients.SendRequest(http.MethodPost, url, headers, formValues, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("[ERROR]: Http request failed. [Status code]: %d, [Response]: %s", statusCode, string(body))
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
