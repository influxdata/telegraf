package gw8

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/gw8/addons"
	"github.com/patrickmn/go-cache"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultMonitoringRoute = "/api/monitoring?dynamic=true"
)

// Login and logout routes from Groundwork API
const (
	loginUrl  = "/api/auth/login"
	logoutUrl = "/api/auth/logout"
)

var (
	tracerOnce sync.Once
)

// Variables for building and updating tracer token
var (
	tracerToken         []byte
	cacheKeyTracerToken = "cacheKeyTraceToken"
	tracerCache         = cache.New(-1, -1)
)

var (
	sampleConfig = `
  ## HTTP endpoint for your groundwork instance.
  groundwork_endpoint = ""

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

  ## Default service state [default - "host"]
  default_service_state = "SERVICE_OK"

  ## The name of the tag that contains the hostname
  resource_tag = "host"
`
)

type GW8 struct {
	Server              string `toml:"groundwork_endpoint"`
	AgentId             string `toml:"agent_id"`
	AppType             string `toml:"app_type"`
	Username            string `toml:"username"`
	Password            string `toml:"password"`
	DefaultHost         string `toml:"default_host"`
	DefaultServiceState string `toml:"default_service_state"`
	ResourceTag         string `toml:"resource_tag"`
	authToken           string
}

func (g *GW8) SampleConfig() string {
	return sampleConfig
}

func (g *GW8) Connect() error {
	if g.Server == "" {
		return errors.New("Groundwork endpoint\\username\\password are not provided ")
	}

	if byteToken, err := login(g.Server+loginUrl, g.Username, g.Password); err == nil {
		g.authToken = string(byteToken)
	} else {
		return err
	}

	return nil
}

func (g *GW8) Close() error {
	formValues := map[string]string{
		"gwos-app-name":  "gw8",
		"gwos-api-token": g.authToken,
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	_, _, err := addons.SendRequest(http.MethodPost, g.Server+logoutUrl, headers, formValues, nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *GW8) Write(metrics []telegraf.Metric) error {
	resourceToServicesMap := make(map[string][]addons.DynamicMonitoredService)
	for _, metric := range metrics {
		resource, service := parseMetric(g.DefaultHost, g.DefaultServiceState, g.ResourceTag, metric)
		resourceToServicesMap[resource] = append(resourceToServicesMap[resource], service)
	}

	var resources []addons.DynamicMonitoredResource
	for resourceName, services := range resourceToServicesMap {
		resources = append(resources, addons.DynamicMonitoredResource{
			Name:          resourceName,
			Type:          addons.Host,
			Status:        addons.HostUp,
			LastCheckTime: &addons.Timestamp{Time: time.Now()},
			Services:      services,
		})
	}

	requestJson, err := json.Marshal(addons.DynamicResourcesWithServicesRequest{
		Context: &addons.TracerContext{
			AppType:    g.AppType,
			AgentID:    g.AgentId,
			TraceToken: makeTracerToken(),
			TimeStamp:  &addons.Timestamp{Time: time.Now()},
			Version:    addons.ModelVersion,
		},
		Resources: resources,
		Groups:    nil,
	})

	if err != nil {
		return err
	}

	headers := map[string]string{
		"GWOS-APP-NAME":  "gw8",
		"GWOS-API-TOKEN": g.authToken,
		"Content-Type":   "application/json",
		"Accept":         "application/json",
	}

	statusCode, _, httpErr := addons.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJson)
	if err != nil {
		return httpErr
	}

	/* Re-login mechanism */
	if statusCode == 401 {
		if err = g.Connect(); err != nil {
			return err
		}
		headers["GWOS-API-TOKEN"] = g.authToken
		statusCode, body, httpErr := addons.SendRequest(http.MethodPost, g.Server+defaultMonitoringRoute, headers, nil, requestJson)
		if httpErr != nil {
			return httpErr
		}
		if statusCode != 200 {
			return errors.New(fmt.Sprintf("something went wrong during processing an http request[http_status = %d, body = %s]", statusCode, string(body)))
		}
	}

	return nil
}

func parseMetric(defaultHostname, defaultServiceState, resourceTag string, metric telegraf.Metric) (string, addons.DynamicMonitoredService) {
	resource := "default_telegraf"
	if defaultHostname != "" {
		resource = defaultHostname
	}
	if resourceTag == "" {
		resourceTag = "host"
	}
	if value, present := metric.GetTag(resourceTag); present {
		resource = value
	}

	service := metric.Name()
	if value, present := metric.GetTag("service"); present {
		service = value
	}

	status := string(addons.ServiceOk)
	if defaultServiceState != "" && validStatus(defaultServiceState) {
		status = defaultServiceState
	}
	if value, present := metric.GetTag("status"); present {
		if validStatus(value) {
			status = value
		}
	}

	message := ""
	if value, present := metric.GetTag("message"); present {
		message = value
	}

	unitType := string(addons.UnitCounter)
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

	serviceObject := addons.DynamicMonitoredService{
		Name:             service,
		Type:             addons.Service,
		Owner:            resource,
		Status:           addons.MonitorStatus(status),
		LastCheckTime:    &addons.Timestamp{Time: metric.Time()},
		LastPlugInOutput: message,
		Metrics:          nil,
	}

	for _, value := range metric.FieldList() {
		var thresholds []addons.ThresholdValue
		thresholds = append(thresholds, addons.ThresholdValue{
			SampleType: addons.Warning,
			Label:      value.Key + "_wn",
			Value: &addons.TypedValue{
				ValueType:   addons.DoubleType,
				DoubleValue: warning,
			},
		})
		thresholds = append(thresholds, addons.ThresholdValue{
			SampleType: addons.Critical,
			Label:      value.Key + "_cr",
			Value: &addons.TypedValue{
				ValueType:   addons.DoubleType,
				DoubleValue: critical,
			},
		})

		val, _ := internal.ToFloat64(value.Value)
		serviceObject.Metrics = append(serviceObject.Metrics, addons.TimeSeries{
			MetricName: value.Key,
			SampleType: addons.Value,
			Interval: &addons.TimeInterval{
				EndTime:   &addons.Timestamp{Time: time.Now()},
				StartTime: &addons.Timestamp{Time: time.Now()},
			},
			Value: &addons.TypedValue{
				ValueType:   addons.DoubleType,
				DoubleValue: val,
			},
			Unit:       addons.UnitType(unitType),
			Thresholds: &thresholds,
		})
	}

	serviceObject.Status, _ = calculateServiceStatus(&serviceObject.Metrics)

	return resource, serviceObject
}

func (g *GW8) Description() string {
	return "Send telegraf metrics to groundwork"
}

func login(url, username, password string) ([]byte, error) {
	formValues := map[string]string{
		"user":          username,
		"password":      password,
		"gwos-app-name": "gw8",
	}
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "text/plain",
	}

	statusCode, body, err := addons.SendRequest(http.MethodPost, url, headers, formValues, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, errors.New(fmt.Sprintf("[ERROR]: Http request failed. [Status code]: %d, [Response]: %s",
			statusCode, string(body)))
	}

	return body, nil
}

func calculateServiceStatus(metrics *[]addons.TimeSeries) (addons.MonitorStatus, error) {
	if metrics == nil || len(*metrics) == 0 {
		return addons.ServiceUnknown, nil
	}
	previousStatus := addons.ServiceOk
	for _, metric := range *metrics {
		if metric.Thresholds != nil {
			var warning, critical addons.ThresholdValue
			for _, threshold := range *metric.Thresholds {
				switch threshold.SampleType {
				case addons.Warning:
					warning = threshold
				case addons.Critical:
					critical = threshold
				default:
					return addons.ServiceOk, fmt.Errorf("unsupported threshold Sample type")
				}
			}

			status := calculateStatus(metric.Value, warning.Value, critical.Value)
			if addons.MonitorStatusWeightService[status] > addons.MonitorStatusWeightService[previousStatus] {
				previousStatus = status
			}
		}
	}
	return previousStatus, nil
}

func calculateStatus(value *addons.TypedValue, warning *addons.TypedValue, critical *addons.TypedValue) addons.MonitorStatus {
	if warning == nil && critical == nil {
		return addons.ServiceOk
	}

	var warningValue float64
	var criticalValue float64

	if warning != nil {
		switch warning.ValueType {
		case addons.DoubleType:
			warningValue = warning.DoubleValue
		}
	}

	if critical != nil {
		switch critical.ValueType {
		case addons.DoubleType:
			criticalValue = critical.DoubleValue
		}
	}

	switch value.ValueType {
	case addons.DoubleType:
		if warning == nil && criticalValue == -1 {
			if value.DoubleValue >= criticalValue {
				return addons.ServiceUnscheduledCritical
			}
			return addons.ServiceOk
		}
		if critical == nil && (warning != nil && warningValue == -1) {
			if value.DoubleValue >= warningValue {
				return addons.ServiceWarning
			}
			return addons.ServiceOk
		}
		if (warning != nil && critical != nil) && (warningValue == -1 || criticalValue == -1) {
			return addons.ServiceOk
		}
		// is it a reverse comparison (low to high)
		if warningValue > criticalValue {
			if value.DoubleValue <= criticalValue {
				return addons.ServiceUnscheduledCritical
			}
			if value.DoubleValue <= warningValue {
				return addons.ServiceWarning
			}
			return addons.ServiceOk
		} else {
			if value.DoubleValue >= criticalValue {
				return addons.ServiceUnscheduledCritical
			}
			if value.DoubleValue >= warningValue {
				return addons.ServiceWarning
			}
			return addons.ServiceOk
		}
	}
	return addons.ServiceOk
}

func validStatus(status string) bool {
	return status == string(addons.ServiceOk) ||
		status == string(addons.ServiceWarning) ||
		status == string(addons.ServicePending) ||
		status == string(addons.ServiceScheduledCritical) ||
		status == string(addons.ServiceUnscheduledCritical) ||
		status == string(addons.ServiceUnknown)
}

// makeTracerContext
func makeTracerToken() string {
	tracerOnce.Do(initTracerToken)

	/* combine TraceToken from fixed and incremental parts */
	tokenBuf := make([]byte, 16)
	copy(tokenBuf, tracerToken)
	if tokenInc, err := tracerCache.IncrementUint64(cacheKeyTracerToken, 1); err == nil {
		binary.PutUvarint(tokenBuf, tokenInc)
	} else {
		/* fallback with timestamp */
		binary.PutVarint(tokenBuf, time.Now().UnixNano())
	}
	traceToken, _ := uuid.FormatUUID(tokenBuf)

	return traceToken
}

func initTracerToken() {
	/* prepare random tracerToken */
	token := []byte("aaaabbbbccccdddd")
	if randBuf, err := uuid.GenerateRandomBytes(16); err == nil {
		copy(tracerToken, randBuf)
	} else {
		/* fallback with multiplied timestamp */
		binary.PutVarint(tracerToken, time.Now().UnixNano())
		binary.PutVarint(tracerToken[6:], time.Now().UnixNano())
	}
	tracerCache.Set(cacheKeyTracerToken, uint64(1), -1)
	tracerToken = token
}

func init() {
	outputs.Add("gw8", func() telegraf.Output {
		return &GW8{}
	})
}
