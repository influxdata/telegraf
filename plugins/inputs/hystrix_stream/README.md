# Hystrix Stream Servlet Input Plugin

The example plugin gathers metrics from the hystrix-stream-servlet.  
This description explains at a high level what the plugin does and 
provides links to where additional information can be found.

### Configuration:

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage hystrix-stream`.

```toml
[[inputs.HystrixStream]]
## Hystrix stream servlet to connect to (with port and full path)
   hystrix_servlet_url = "http://localhost:8090/hystrix"
```

### Metrics:

All counters from the hystrix-stream-servlet are collected. 
The rolling-values are not collected since these are aggregates and can be computed at a later point.

See the [hystrix-documentation](https://github.com/Netflix/Hystrix/wiki/Metrics-and-Monitoring) for details of the individual metrics.

The measurements are named after the CommandKey and the GroupKey and tagged with hostname and threadpoolname.

So if your command is called "MyCommand" in the group "MyGroup" in the threadpool "MyThreadPool" on the host "Myhost", 
an entry could look like this:

- MyGroupMyCommand
  - tags:
    - host (MyHost)
    - group (MyGroup)
    - type (HystrixCommand or HystrixThreadPool)
    - name
  - fields:
    - RequestCount
    - ErrorCount
    - ReportingHosts
    - ErrorPercentage
    - IsCircuitBreakerOpen
    - CurrentConcurrentExecutionCount
    - LatencyExecute(0,25,50,75,90,95,99,100)
    - LatencyTotal(0,25,50,75,90,95,99,100)
    

### Sample Queries:

This section should contain some useful InfluxDB queries that can be used to
get started with the plugin or to generate dashboards.  For each query listed,
describe at a high level what data is returned.

Get the max, mean, and min RequestCount for the command in the last hour:
```
SELECT max(RequestCount), mean(RequestCount), min(RequestCount) FROM MyGroupMyCommand WHERE time > now() - 1h 
```

### Example Output:


```
 MyGroupMyCommand,name=MyCommand,type=HystrixCommand,group=MyGroup,threadpool=MyThreadPool,host=yoga900 LatencyTotal99=507i,LatencyExecute90=442i,LatencyExecute95=504i,LatencyExecute99=507i,LatencyTotal0=1i,LatencyTotal50=270i,LatencyTotal95=504i,ReportingHosts=1i,CurrentConcurrentExecutionCount=1i,LatencyTotal25=144i,LatencyExecute0=0i,LatencyExecute100=507i,LatencyExecute25=144i,LatencyExecute50=270i,ErrorPercentage=20i,ErrorCount=2i,LatencyTotal90=442i,LatencyTotal100=508i,IsCircuitBreakerOpen=false,RequestCount=10i,LatencyTotal75=349i,LatencyExecute75=349i 1507189604000000000
 MyGroupMyCommand,name=MyCommand,type=HystrixCommand,group=MyGroup,threadpool=MyThreadPool,host=yoga900 LatencyTotal90=442i,LatencyExecute0=0i,ReportingHosts=1i,CurrentConcurrentExecutionCount=0i,ErrorCount=2i,LatencyTotal50=270i,LatencyTotal95=504i,LatencyTotal99=507i,LatencyExecute25=144i,LatencyExecute95=504i,LatencyExecute100=507i,ErrorPercentage=20i,LatencyTotal75=349i,LatencyTotal100=508i,LatencyExecute50=270i,LatencyExecute75=349i,IsCircuitBreakerOpen=false,RequestCount=10i,LatencyTotal0=1i,LatencyTotal25=144i,LatencyExecute90=442i,LatencyExecute99=507i 1507189605000000000
```