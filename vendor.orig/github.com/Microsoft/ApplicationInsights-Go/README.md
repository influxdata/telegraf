# Microsoft Application Insights SDK for Go

[![Build Status](https://travis-ci.org/Microsoft/ApplicationInsights-Go.svg?branch=master)](https://travis-ci.org/Microsoft/ApplicationInsights-Go) [![Documentation](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go?status.svg)](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights) [![Release](https://img.shields.io/github/release/Microsoft/ApplicationInsights-Go/all.svg)](https://github.com/Microsoft/ApplicationInsights-Go/releases)

This project provides a Go SDK for Application Insights.
[Application Insights](http://azure.microsoft.com/en-us/services/application-insights/)
is a service that allows developers to keep their applications available,
performant, and successful.  This go package will allow you to send
telemetry of various kinds (event, metric, trace) to the Application
Insights service where they can be visualized in the Azure Portal.

## Status
This SDK is considered to be pre-production.  It has not reached parity with
some of the more mature SDK's.  In particular, the gaps are:

* Operation correlation is not supported, but this can be managed by the
  caller through the interfaces that exist today.
* Sampling is not supported.  The more mature SDKs support dynamic sampling,
  but at present this does not even support manual sampling.
* Automatic collection of events is not supported.  All telemetry must be
  explicitly collected and sent by the user.
* Offline storage of telemetry is not supported.  The .Net SDK is capable of
  spilling events to disk in case of network interruption.  This SDK has no
  such feature.

Additionally, this is considered a community-supported SDK.  Read more about
the status of this and other SDK's in the
[ApplicationInsights-Home](https://github.com/Microsoft/ApplicationInsights-Home)
repository.

## Requirements
**Install**
```
go get github.com/Microsoft/ApplicationInsights-Go/appinsights
```
**Get an instrumentation key**
>**Note**: an instrumentation key is required before any data can be sent. Please see the "[Getting an Application Insights Instrumentation Key](https://github.com/Microsoft/AppInsights-Home/wiki#getting-an-application-insights-instrumentation-key)" section of the wiki for more information. To try the SDK without an instrumentation key, set the instrumentationKey config value to a non-empty string.

# Usage

## Setup

To start tracking telemetry, you'll want to first initialize a
[telemetry client](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#TelemetryClient).

```go
import "github.com/Microsoft/ApplicationInsights-Go/appinsights"

func main() {
	client := appinsights.NewTelemetryClient("<instrumentation key>")
}
```

If you want more control over the client's behavior, you should initialize a
new [TelemetryConfiguration](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#TelemetryConfiguration)
object and use it to create a client:

```go
import "time"
import "github.com/Microsoft/ApplicationInsights-Go/appinsights"

func main() {
	telemetryConfig := appinsights.NewTelemetryConfiguration("<instrumentation key>")
	
	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192
	
	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second
	
	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
}
```

This client will be used to submit all of your telemetry to Application
Insights.  This SDK does not presently collect any telemetry automatically,
so you will use this client extensively to report application health and
status.  You may want to store it in a global variable or otherwise include
it in your data model.

## Telemetry submission

The [TelemetryClient](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#TelemetryClient)
itself has several methods for submitting telemetry:

```go
type TelemetryClient interface {
	// (much omitted)

	// Log a user action with the specified name
	TrackEvent(name string)

	// Log a numeric value that is not specified with a specific event.
	// Typically used to send regular reports of performance indicators.
	TrackMetric(name string, value float64)

	// Log a trace message with the specified severity level.
	TrackTrace(name string, severity contracts.SeverityLevel)

	// Log an HTTP request with the specified method, URL, duration and
	// response code.
	TrackRequest(method, url string, duration time.Duration, responseCode string)

	// Log a dependency with the specified name, type, target, and
	// success status.
	TrackRemoteDependency(name, dependencyType, target string, success bool)

	// Log an availability test result with the specified test name,
	// duration, and success status.
	TrackAvailability(name string, duration time.Duration, success bool)

	// Log an exception with the specified error, which may be a string,
	// error or Stringer. The current callstack is collected
	// automatically.
	TrackException(err interface{})
}
```

These may be used directly to log basic telemetry a manner you might expect:

```go
client.TrackMetric("Queue Length", len(queue))

client.TrackEvent("Client connected")
```

But the inputs to these methods only capture the very basics of what these
telemetry types can represent.  For example, all telemetry supports custom
properties, which are inaccessible through the above methods.  More complete
versions are available through use of *telemetry item* classes, which can
then be submitted through the `TelemetryClient.Track` method, as illustrated
in the below sections:

### Trace
[Trace telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#TraceTelemetry)
represent printf-like trace statements that can be text searched.  They have
an associated severity level, values for which are found in the package's
constants:

```go
const (
	Verbose     contracts.SeverityLevel = contracts.Verbose
	Information contracts.SeverityLevel = contracts.Information
	Warning     contracts.SeverityLevel = contracts.Warning
	Error       contracts.SeverityLevel = contracts.Error
	Critical    contracts.SeverityLevel = contracts.Critical
)
```

Trace telemetry is fairly simple, but common telemetry properties are also
available:

```go
trace := appinsights.NewTraceTelemetry("message", appinsights.Warning)

// You can set custom properties on traces
trace.Properties["module"] = "server"

// You can also fudge the timestamp:
trace.Timestamp = time.Now().Sub(time.Minute)

// Finally, track it
client.Track(trace)
```

### Events
[Event telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#EventTelemetry)
represent structured event records.

```go
event := appinsights.NewEventTelemetry("button clicked")
event.Properties["property"] = "value"
client.Track(event)
```

### Single-value metrics
[Metric telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#MetricTelemetry)
each represent a single data point.

```go
metric := appinsights.NewMetricTelemetry("Queue length", len(q.items))
metric.Properties["Queue name"] = q.name
client.Track(metric)
```

### Pre-aggregated metrics
To reduce the number of metric values that may be sent through telemetry,
when using a particularly high volume of measurements, metric data can be
[pre-aggregated by the client](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#AggregateMetricTelemetry)
and submitted all at once.

```go
aggregate := appinsights.NewAggregateMetricTelemetry("metric name")

var dataPoints []float64
// ...collect data points...

// If the data is sampled, then one should use the AddSampledData method to
// feed data to this telemetry type.
aggregate.AddSampledData(dataPoints)

// If the entire population of data points is known, then one should instead
// use the AddData method.  The difference between the two is the manner in
// which the standard deviation is calculated.
aggregate.AddData(dataPoints)

// Alternatively, you can aggregate the data yourself and supply it to this
// telemetry item:
aggregate.Value = sum(dataPoints)
aggregate.Min = min(dataPoints)
aggregate.Max = max(dataPoints)
aggregate.Count = len(dataPoints)
aggregate.StdDev = stdDev(dataPoints)

// Custom properties could be further added here...

// Finally, track it:
client.Track(aggregate)
```

### Requests
[Request telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#RequestTelemetry)
represent completion of an external request to the application and contains
a summary of that request execution and results.  This SDK's request
telemetry is focused on HTTP requests.

```go
request := appinsights.NewRequestTelemetry("GET", "https://microsoft.com/", duration, "<response code>")

// Note that the timestamp will be set to time.Now() minus the
// specified duration.  This can be overridden by either manually
// setting the Timestamp and Duration fields, or with MarkTime:
request.MarkTime(requestStartTime, requestEndTime)

// Source of request
request.Source = clientAddress

// Success is normally inferred from the responseCode, but can be overridden:
request.Success = false

// Request ID's are randomly generated GUIDs, but this can also be overridden:
request.Id = "<id>"

// Custom properties and measurements can be set here
request.Properties["user-agent"] = request.headers["User-agent"]
request.Measurements["POST size"] = float64(len(data))

// Context tags become more useful here as well
request.Tags.Session().SetId("<session id>")
request.Tags.User().SetAccountId("<user id>")

// Finally track it
client.Track(request)
```

### Dependencies
[Remote dependency telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#RemoteDependencyTelemetry)
represent interactions of the monitored component with a remote
component/service like SQL or an HTTP endpoint.

```go
dependency := appinsights.NewRemoteDependencyTelemetry("Redis cache", "Redis", "<target>", true /* success */)

// The result code is typically an error code or response status code
dependency.ResultCode = "OK"

// Id's can be used for correlation if the remote end is also logging
// telemetry through application insights.
dependency.Id = "<request id>"

// Data may contain the exact URL hit or SQL statements
dependency.Data = "MGET <args>"

// The duration can be set directly:
dependency.Duration = time.Minute
// or via MarkTime:
dependency.MarkTime(startTime, endTime)

// Properties and measurements may be set.
dependency.Properties["shard-instance"] = "<name>"
dependency.Measurements["data received"] = float64(len(response.data))

// Submit the telemetry
client.Track(dependency)
```

### Exceptions
[Exception telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#ExceptionTelemetry)
represent handled or unhandled exceptions that occurred during the execution
of the monitored application.  This SDK is geared towards handling panics or
unexpected results from important functions:

To handle a panic:

```go
func method(client appinsights.TelemetryClient) {
	defer func() {
		if r := recover(); r != nil {
			// Track the panic
			client.TrackException(r)

			// Optionally, you may want to re-throw the panic:
			panic(r)
		}
	}()
	
	// Panics in any code below will be handled by the above.
	panic("AHHHH!!")
}
```

This can be condensed with a helper function:

```go
func method(client appinsights.TelemetryClient) {
	// false indicates that we should have this handle the panic, and
	// not re-throw it.
	defer appinsights.TrackPanic(client, false)
	
	// Panics in any code below will be handled by the above.
	panic("AHHHH!!")
}
```

This will capture and report the call stack of the panic, including the site
of the function that handled the panic.  Do note that Go does not unwind the
callstack while processing panics, so the trace will include any functions
that may be called by `method` in the example above leading up to the panic.

This SDK will handle panic messages that are any of the types: `string`,
`error`, or anything that implements [fmt.Stringer](https://golang.org/pkg/fmt/#Stringer)
or [fmt.GoStringer](https://golang.org/pkg/fmt/#GoStringer).

While the above example uses `client.TrackException`, you can also use the
longer form as in earlier examples -- and not only for panics:

```go
value, err := someMethod(argument)
if err != nil {
	exception := appinsights.NewExceptionTelemetry(err)
	
	// Set the severity level -- perhaps this isn't a critical
	// issue, but we'd *really rather* it didn't fail:
	exception.SeverityLevel = appinsights.Warning
	
	// One could tweak the number of stack frames to skip by
	// reassigning the callstack -- for instance, if you were to
	// log this exception in a helper method.
	exception.Frames = appinsights.GetCallstack(3 /* frames to skip */)
	
	// Properties are available as usual
	exception.Properties["input"] = argument
	
	// Track the exception
	client.Track(exception)
}
```

### Availability
[Availability telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights/#AvailabilityTelemetry)
represent the result of executing an availability test.  This is useful if
you are writing availability monitors in Go.

```go
availability := appinsights.NewAvailabilityTelemetry("test name", callDuration, true /* success */)

// The run location indicates where the test was run from
availability.RunLocation = "Phoenix"

// Diagnostics message
availability.Message = diagnostics

// Id is used for correlation with the target service
availability.Id = requestId

// Timestamp and duration can be changed through MarkTime, similar
// to other telemetry types with Duration's
availability.MarkTime(testStartTime, testEndTime)

// Submit the telemetry
client.Track(availability)
```

### Page Views
[Page view telemetry items](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights/#PageViewTelemetry)
represent generic actions on a page like a button click.  These are typically
generated by the client side rather than the server side, but is available
here nonetheless.

```go
pageview := appinsights.NewPageViewTelemetry("Event name", "http://testuri.org/page")

// A duration is available here.
pageview.Duration = time.Minute

// As are the usual Properties and Measurements...

// Track
client.Track(pageview)
```

### Context tags
Telemetry items all have a `Tags` property that contains information *about*
the submitted telemetry, such as user, session, and device information.  The
`Tags` property is an instance of the
[contracts.ContextTags](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts/#ContextTags)
type, which is a `map[string]string` under the hood, but has helper methods
to access the most commonly used data.  An instance of
[TelemetryContext](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights/#TelemetryContext)
exists on the `TelemetryClient`, and also contains a `Tags` property.  These
tags are applied to all telemetry sent through the client.  If a context tag
is found on both the client's `TelemetryContext` and in the telemetry item's
`Tags`, the value associated with the telemtry takes precedence.

A few examples for illustration:

```go
import (
	"os"
	
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
)

func main() {
	client := appinsights.NewTelemetryClient("<ikey>")
	
	// Set role instance name globally -- this is usually the
	// name of the service submitting the telemetry
	client.Context().Tags.Cloud().SetRole("my_go_server")
	
	// Set the role instance to the host name.  Note that this is
	// done automatically by the SDK.
	client.Context().Tags.Cloud().SetRoleInstance(os.Hostname())
	
	// Make a request to fiddle with the telemetry's context
	req := appinsights.NewRequestTelemetry("GET", "http://server/path", time.Millisecond, "200")
	
	// Set the account ID context tag, for this telemetry item
	// only.  The following are equivalent:
	req.Tags.User().SetAccountId("<user account retrieved from request>")
	req.Tags[contracts.UserAccountId] = "<user account retrieved from request>"
	
	// This request will have all context tags above.
	client.Track(req)
}
```

### Common properties

In the same way that context tags can be written to all telemetry items, the
`TelemetryContext` has a `CommonProperties` map.  Entries in this map will
be added to all telemetry items' custom properties (unless a telemetry item
already has that property set -- the telemetry item always has precedence). 
This is useful for contextual data that may not be captured in the context
tags, for instance cluster identifiers or resource groups.

```go
func main() {
	client := appinsights.NewTelemetryClient("<ikey>")
	
	client.Context().CommonProperties["Resource group"] = "My resource group"
	// ...
}
```

### Shutdown
The Go SDK submits data asynchronously.  The [InMemoryChannel](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights/#InMemoryChannel)
launches its own goroutine used to accept and send telemetry.  If you're not
careful, this may result in lost telemetry when the service needs to shut
down.  The channel has a few methods to deal with this case:

* `Flush` will trigger telemetry submission for buffered items.  It returns
  immediately and telemetry is not guaranteed to have been sent.
* `Stop` will immediately shut down the channel and discard any unsubmitted
  telemetry.  Useful if you need to exit NOW.
* `Close` will cause the channel to stop accepting new telemetry, submit any
  pending telemetry, and returns a channel that closes when the telemetry
  buffer is fully empty.  If telemetry submission fails, then `Close` will
  retry until the specified duration elapses.  If no duration is specified,
  then it will give up if any telemetry submission fails.

If at all possible, you should use `Close`:

```go
func main() {
	client := appinsights.NewTelemetryClient("<ikey>")
	
	// ... run the service ...
	
	// on shutdown:
	
	select {
	case <-client.Channel().Close(10 * time.Second):
		// Ten second timeout for retries.
		
		// If we got here, then all telemetry was submitted
		// successfully, and we can proceed to exiting.
	case <-time.After(30 * time.Second):
		// Thirty second absolute timeout.  This covers any
		// previous telemetry submission that may not have
		// completed before Close was called.
		
		// There are a number of reasons we could have
		// reached here.  We gave it a go, but telemetry
		// submission failed somewhere.  Perhaps old events
		// were still retrying, or perhaps we're throttled.
		// Either way, we don't want to wait around for it
		// to complete, so let's just exit.
	}
}
```

We recommend something similar to the above to minimize lost telemetry
through shutdown.
[The documentation](https://godoc.org/github.com/Microsoft/ApplicationInsights-Go/appinsights#TelemetryChannel)
explains in more detail what can lead to the cases above.

### Diagnostics
If you find yourself missing some of the telemetry that you thought was
submitted, diagnostics can be turned on to help troubleshoot problems with
telemetry submission.

```go
appinsights.NewDiagnosticsMessageListener(func(msg string) error {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
	return nil
})

// go about your business...
```

The SDK will emit messages during every telemetry submission.  Successful
submissions will look something like this:

```
[Tue Nov 21 18:59:41 PST 2017] --------- Transmitting 16 items ---------
[Tue Nov 21 18:59:41 PST 2017] Telemetry transmitted in 708.382896ms
[Tue Nov 21 18:59:41 PST 2017] Response: 200
```

If telemetry is rejected, the errors from the data collector endpoint will
be displayed:

```
[Tue Nov 21 18:58:39 PST 2017] --------- Transmitting 16 items ---------
[Tue Nov 21 18:58:40 PST 2017] Telemetry transmitted in 1.034608896s
[Tue Nov 21 18:58:40 PST 2017] Response: 206
[Tue Nov 21 18:58:40 PST 2017] Items accepted/received: 15/16
[Tue Nov 21 18:58:40 PST 2017] Errors:
[Tue Nov 21 18:58:40 PST 2017] #9 - 400 109: Field 'name' on type 'RemoteDependencyData' is required but missing or empty. Expected: string, Actual:
[Tue Nov 21 18:58:40 PST 2017] Telemetry item:
        {"ver":1,"name":"Microsoft.ApplicationInsights.RemoteDependency","time":"2017-11-22T02:58:39Z","sampleRate":100,"seq":"","iKey":"<ikey>","tags":{"ai.cloud.roleInstance":"<hostname>","ai.device.id":"<hostname>","ai.device.osVersion":"linux","ai.internal.sdkVersion":"go:0.4.0-pre","ai.operation.id":"bf755161-7725-490c-872e-69815826a94c"},"data":{"baseType":"RemoteDependencyData","baseData":{"ver":2,"name":"","id":"","resultCode":"","duration":"0.00:00:00.0000000","success":true,"data":"","target":"http://bing.com","type":"HTTP"}}}

[Tue Nov 21 18:58:40 PST 2017] Refusing to retry telemetry submission (retry==false)
```

Information about retries, server throttling, and more from the SDK's
perspective will also be available.

Please include this diagnostic information (with ikey's blocked out) when
submitting bug reports to this project.
