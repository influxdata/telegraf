# Cisco model-driven telemetry (MDT)

Cisco model-driven telemetry (MDT) is an input plugin that consumes
telemetry data from Cisco IOS XR, IOS XE and NX-OS platforms. It supports TCP & GRPC dialout (server) and GRPC dialin (client) transports.
GRPC-based transport can utilize TLS for authentication and encryption.
Telemetry data is expected to be GPB-KV (self-describing-gpb) encoded.

The GRPC dialout transport is supported on various IOS XR (64-bit) 6.1.x and later, IOS XE 16.10 and later, as well as NX-OS 7.x and later platforms.

The GRPC dialin transport is supported on IOS XR (64-bit) 6.1.x and later.

The TCP dialout transport is supported on IOS XR (32-bit and 64-bit) 6.1.x and later.


### Configuration:

This is a sample configuration for the plugin.

```toml
[[inputs.cisco_telemetry_mdt]]
  ## Telemetry transport (one of: tcp-dialout, grpc-dialout, grpc-dialin)
  transport = "grpc-dialout"

  ## Address and port to host telemetry listener on (dialout) or to connect to (dialin)
  service_address = ":57000"

  ## Enable TLS for transport
  # tls = true

  ## grpc-dialin: define credentials and subscription
  # username = "cisco"
  # password = "cisco"
  # subscription = "subscription"
  # redial = "10s"

  ## grpc-dialin: define TLS CA to authenticate the device
  # tls_ca = "/etc/telegraf/ca.pem"
  # insecure_skip_verify = true

  ## grpc-dialin: define client-side TLS certificate & key to authenticate to the device
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"


  ## grpc-dialout: define TLS certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## grpc-dialout: enable TLS client authentication and define allowed CA certificates
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
```
