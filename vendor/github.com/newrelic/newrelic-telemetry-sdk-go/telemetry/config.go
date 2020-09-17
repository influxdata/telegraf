// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Config customizes the behavior of a Harvester.
type Config struct {
	// APIKey is required.
	APIKey string
	// Client is the http.Client used for making requests.
	Client *http.Client
	// HarvestTimeout is the total amount of time including retries that the
	// Harvester may use trying to harvest data.  By default, HarvestTimeout
	// is set to 15 seconds.
	HarvestTimeout time.Duration
	// CommonAttributes are the attributes to be applied to all metrics that
	// use this Config. They are not applied to spans.
	CommonAttributes map[string]interface{}
	// HarvestPeriod controls how frequently data will be sent to New Relic.
	// If HarvestPeriod is zero then NewHarvester will not spawn a goroutine
	// to send data and it is incumbent on the consumer to call
	// Harvester.HarvestNow when data should be sent. By default, HarvestPeriod
	// is set to 5 seconds.
	HarvestPeriod time.Duration
	// ErrorLogger receives errors that occur in this sdk.
	ErrorLogger func(map[string]interface{})
	// DebugLogger receives structured debug log messages.
	DebugLogger func(map[string]interface{})
	// AuditLogger receives structured log messages that include the
	// uncompressed data sent to New Relic.  Use this to log all data sent.
	AuditLogger func(map[string]interface{})
	// MetricsURLOverride overrides the metrics endpoint if not not empty.
	MetricsURLOverride string
	// SpansURLOverride overrides the spans endpoint if not not empty.
	SpansURLOverride string
	// Product is added to the User-Agent header. eg. "NewRelic-Go-OpenCensus"
	Product string
	// ProductVersion is added to the User-Agent header. eg. "0.1.0".
	ProductVersion string
}

// ConfigAPIKey sets the Config's APIKey which is required.
func ConfigAPIKey(key string) func(*Config) {
	return func(cfg *Config) {
		cfg.APIKey = key
	}
}

// ConfigCommonAttributes adds the given attributes to the Config's
// CommonAttributes.
func ConfigCommonAttributes(attributes map[string]interface{}) func(*Config) {
	return func(cfg *Config) {
		cfg.CommonAttributes = attributes
	}
}

// ConfigHarvestPeriod sets the Config's HarvestPeriod field which controls the
// rate data is reported to New Relic.  If it is set to zero then the Harvester
// will never report data unless HarvestNow is called.
func ConfigHarvestPeriod(period time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.HarvestPeriod = period
	}
}

func newBasicLogger(w io.Writer) func(map[string]interface{}) {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	lg := log.New(w, "", flags)
	return func(fields map[string]interface{}) {
		if js, err := json.Marshal(fields); nil != err {
			lg.Println(err.Error())
		} else {
			lg.Println(string(js))
		}
	}
}

// ConfigBasicErrorLogger sets the error logger to a simple logger that logs
// to the writer provided.
func ConfigBasicErrorLogger(w io.Writer) func(*Config) {
	return func(cfg *Config) {
		cfg.ErrorLogger = newBasicLogger(w)
	}
}

// ConfigBasicDebugLogger sets the debug logger to a simple logger that logs
// to the writer provided.
func ConfigBasicDebugLogger(w io.Writer) func(*Config) {
	return func(cfg *Config) {
		cfg.DebugLogger = newBasicLogger(w)
	}
}

// ConfigBasicAuditLogger sets the audit logger to a simple logger that logs
// to the writer provided.
func ConfigBasicAuditLogger(w io.Writer) func(*Config) {
	return func(cfg *Config) {
		cfg.AuditLogger = newBasicLogger(w)
	}
}

// configTesting is the config function to be used when testing. It sets the
// APIKey but disables the harvest goroutine.
func configTesting(cfg *Config) {
	cfg.APIKey = "api-key"
	cfg.HarvestPeriod = 0
}

func (cfg *Config) logError(fields map[string]interface{}) {
	if nil == cfg.ErrorLogger {
		return
	}
	cfg.ErrorLogger(fields)
}

func (cfg *Config) logDebug(fields map[string]interface{}) {
	if nil == cfg.DebugLogger {
		return
	}
	cfg.DebugLogger(fields)
}

func (cfg *Config) auditLogEnabled() bool {
	return cfg.AuditLogger != nil
}

func (cfg *Config) logAudit(fields map[string]interface{}) {
	if nil == cfg.AuditLogger {
		return
	}
	cfg.AuditLogger(fields)
}

const (
	defaultSpanURL   = "https://trace-api.newrelic.com/trace/v1"
	defaultMetricURL = "https://metric-api.newrelic.com/metric/v1"
)

func (cfg *Config) spanURL() string {
	if cfg.SpansURLOverride != "" {
		return cfg.SpansURLOverride
	}
	return defaultSpanURL
}

func (cfg *Config) metricURL() string {
	if cfg.MetricsURLOverride != "" {
		return cfg.MetricsURLOverride
	}
	return defaultMetricURL
}

// userAgent creates the User-Agent header version according to the spec here:
// https://github.com/newrelic/newrelic-telemetry-sdk-specs/blob/master/communication.md#user-agent
func (cfg *Config) userAgent() string {
	agent := "NewRelic-Go-TelemetrySDK/" + version
	if "" != cfg.Product {
		agent += " " + cfg.Product
		if "" != cfg.ProductVersion {
			agent += "/" + cfg.ProductVersion
		}
	}
	return agent
}
