## Cloud Insight

The Naver Cloud Platform cloud environment's performance/operation indicators
are integrated and managed, and a monitoring service is provided that can be 
quickly delivered to the person in charge in the event of a failure.

This plugin sends custom metrics to Cloud Insight. Cloud Insight's metric resolution 
can be specified as an interval value. Cloud Insight aggregates and processes, 
so the output plugin sends metrics to Cloud Insight and automatically 
processes them in the cloud.

The metrics for each plugin are recorded in a separate Cloud Insight production 
name prefixed with `Custom/`. Each metric's field name is stored as the metric name 
in Cloud Insight.

[introduce] https://www.fin-ncloud.com/product/management/cloudInsight

### Configuration

```toml
[[outputs.cloudinsight]]
  ## Financial Cloud Region (fin-ncloud.com)
  # region = "FKR"

  ## This option specifies the custom metric type recognized by Cloud Insight.
  ## The prefix must be typed as `/Custom`.
  # product_name = "Custom/"

  ## Key assigned when registering a custom schema in Cloud Insight is complete
  # cw_key = ""

  ## These are the basic authentication keys to access the cloud, and basically, 
  ## an access and secret key are required.
  # access_key = ""
  # secret_key = ""

  ## Key issued by API Gateway
  # api_gateway_key = ""

  ## Instance ID is a unique ID of each VM, and you need to enter the information 
  ## of the VM on which telegraf will run
  # instance_id
```

### Setup

1. [Register the `nbp.cloudinsight` resource provider in you Naver Cloud Platform subscription][resource provider].
2. Use a region that supports Cloud Insight Custom Metrics.
    For regions with Custom Metrics support, an endpoint will be available with
    the format ``. The following regions are currently known to be supported.
     - FKR

### Authentication

Naver Cloud Platform integration API is provided as an open API that can be called 
and used immediately by sending only the Client ID and Client Secret values in the HTTP header.

The plugin will authenticate using the first available of the
following configurations:

** Client Credentials with API Gateway**: NCP Application, Secret and API Gateway.
   
   Set the following environment variables:
   
   - `NCP_ACCESS_KEY`: Specifies the application client ID to use.
   - `NCP_SECRET_KEY`: Specifies the app secret ID to use.
   - `NCP_API_GATEWAY_KEY`: Specifies the certificate API Gateway to use.
   
[credential]: https://apidocs.fin-ncloud.com/ko/common/naver_api/naverapi/

### Dimensions

The Dimension and Metric fields, the default value is false and if it is false it can
be overridden in the request. For the Metric field, an aggregation is specified for
each interval. If there is no aggregation in the field element, the aggregation method
is automatically specified for each interval.