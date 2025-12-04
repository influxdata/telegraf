# OPC UA Client Reader Input Plugin

This plugin gathers data from an [OPC UA][opcua] server by subscribing to the
configured nodes.

⭐ Telegraf v1.16.0
🏷️ iot
💻 all

[opcua]: https://opcfoundation.org/about/opc-technologies/opc-ua/

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `username` and
`password` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Retrieve data from OPCUA devices
[[inputs.opcua]]
  ## Metric name
  # name = "opcua"

  ## OPC UA Endpoint URL
  # endpoint = "opc.tcp://localhost:4840"

  ## Maximum time allowed to establish a connect to the endpoint.
  # connect_timeout = "10s"

  ## Maximum time allowed for a request over the established connection.
  # request_timeout = "5s"

  ## Maximum time that a session shall remain open without activity.
  # session_timeout = "20m"

  ## Retry options for failing reads e.g. due to invalid sessions
  ## If the retry count is zero, the read will fail after the initial attempt.
  # read_retry_timeout = "100ms"
  # read_retry_count = 0

  ## Number of consecutive errors before forcing a reconnection
  ## If set to 1 (default), the client will reconnect after a single failed read
  # reconnect_error_threshold = 1

  ## Security policy, one of "None", "Basic128Rsa15", "Basic256",
  ## "Basic256Sha256", or "auto"
  # security_policy = "auto"

  ## Security mode, one of "None", "Sign", "SignAndEncrypt", or "auto"
  # security_mode = "auto"

  ## Path to client certificate and private key files, must be specified together.
  ## If none of the options are specified, a temporary self-signed certificate
  ## will be created. If the options are specified but the files do not exist, a
  ## self-signed certificate will be created and stored permanently at the
  ## given locations.
  # certificate = "/etc/telegraf/cert.pem"
  # private_key = "/etc/telegraf/key.pem"

  ## Path to additional, explicitly trusted certificate for the remote endpoint
  # remote_certificate = "/etc/telegraf/opcua_server_cert.pem"

  ## Authentication Method, one of "Certificate", "UserName", or "Anonymous".  To
  ## authenticate using a specific ID, select 'Certificate' or 'UserName'
  # auth_method = "Anonymous"

  ## Username and password required for auth_method = "UserName"
  # username = ""
  # password = ""

  ## Option to select the metric timestamp to use. Valid options are:
  ##     "gather" -- uses the time of receiving the data in telegraf
  ##     "server" -- uses the timestamp provided by the server
  ##     "source" -- uses the timestamp provided by the source
  # timestamp = "gather"

  ## Client trace messages
  ## When set to true, and debug mode enabled in the agent settings, the OPCUA
  ## client's messages are included in telegraf logs. These messages are very
  ## noisey, but essential for debugging issues.
  # client_trace = false

  ## Include additional Fields in each metric
  ## Available options are:
  ##   DataType -- OPC-UA Data Type (string)
  # optional_fields = []

  ## Node ID configuration
  ## name              - field name to use in the output
  ## namespace         - OPC UA namespace of the node (integer value 0 thru 3)
  ## namespace_uri     - OPC UA namespace URI (alternative to namespace for stable references)
  ## identifier_type   - OPC UA ID type (s=string, i=numeric, g=guid, b=opaque)
  ## identifier        - OPC UA ID (tag as shown in opcua browser)
  ## default_tags      - extra tags to be added to the output metric (optional)
  ##
  ## Note: Specify either 'namespace' or 'namespace_uri', not both.
  ## Use either the inline notation or the bracketed notation, not both.

  ## Inline notation (default_tags not supported yet)
  # nodes = [
  #   {name="", namespace="", identifier_type="", identifier=""},
  # ]

  ## Bracketed notation
  # [[inputs.opcua.nodes]]
  #   name = "node1"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #   default_tags = { tag1 = "value1", tag2 = "value2" }
  #
  # [[inputs.opcua.nodes]]
  #   name = "node2"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #
  # [[inputs.opcua.nodes]]
  #   name = "node3"
  #   namespace_uri = "http://opcfoundation.org/UA/"
  #   identifier_type = ""
  #   identifier = ""

  ## Node Group
  ## Sets defaults so they aren't required in every node.
  ## Default values can be set for:
  ## * Metric name
  ## * OPC UA namespace
  ## * Identifier
  ## * Default tags
  ##
  ## Multiple node groups are allowed
  #[[inputs.opcua.group]]
  ## Group Metric name. Overrides the top level name.  If unset, the
  ## top level name is used.
  # name =

  ## Group default namespace. If a node in the group doesn't set its
  ## namespace, this is used.
  # namespace =

  ## Group default namespace URI. Alternative to namespace for stable references.
  ## If a node in the group doesn't set its namespace_uri, this is used.
  # namespace_uri =

  ## Group default identifier type. If a node in the group doesn't set its
  ## identifier_type, this is used.
  # identifier_type =

  ## Default tags that are applied to every node in this group. Can be
  ## overwritten in a node by setting a different value for the tag name.
  ##   example: default_tags = { tag1 = "value1" }
  # default_tags = {}

  ## Node ID Configuration. Array of nodes with the same settings as above.
  ## Use either the inline notation or the bracketed notation, not both.

  ## Inline notation (default_tags not supported yet)
  # nodes = [
  #  {name="node1", namespace="", identifier_type="", identifier=""},
  #  {name="node2", namespace="", identifier_type="", identifier=""},
  #]

  ## Bracketed notation
  # [[inputs.opcua.group.nodes]]
  #   name = "node1"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""
  #   default_tags = { tag1 = "override1", tag2 = "value2" }
  #
  # [[inputs.opcua.group.nodes]]
  #   name = "node2"
  #   namespace = ""
  #   identifier_type = ""
  #   identifier = ""

  ## Enable workarounds required by some devices to work correctly
  # [inputs.opcua.workarounds]
  #   ## Set additional valid status codes, StatusOK (0x0) is always considered valid
  #   # additional_valid_status_codes = ["0xC0"]

  # [inputs.opcua.request_workarounds]
  #   ## Use unregistered reads instead of registered reads
  #   # use_unregistered_reads = false
```

### Client Certificate Configuration

When using security modes other than "None", Telegraf acts as an OPC UA client
and requires a client certificate to authenticate itself to the server. The
plugin supports three certificate management approaches:

#### Temporary Self-Signed Certificates (Default)

If both `certificate` and `private_key` options are left empty or commented
out, Telegraf will automatically generate a self-signed certificate in a
temporary directory on each startup.

> [!NOTE]
> These certificates are recreated on every Telegraf restart, requiring
> re-authorization by the OPC UA server each time. This is suitable for testing
> but not recommended for production environments.

#### Persistent Self-Signed Certificates (Recommended for Testing)

To maintain the same client identity across restarts, specify paths for both
`certificate` and `private_key`. If the files don't exist, Telegraf will
generate them at the specified locations and reuse them on subsequent restarts.

> [!IMPORTANT]
> Ensure Telegraf has write permissions to the specified paths. On first run,
> Telegraf will generate the certificates and log their locations. On subsequent
> restarts, Telegraf will reuse the existing certificates, preventing the need
> to re-authorize the client in the server's trust store.

#### Manual Certificate Management (Production)

For production environments, manually generate and deploy certificates using
your organization's PKI infrastructure. Place the certificate and private key
files at the configured paths before starting Telegraf. If both files exist,
Telegraf will use them without modification.

#### Certificate Validation Rules

- Both `certificate` and `private_key` must be specified together, or both
  must be empty
- If one file exists but the other doesn't, Telegraf will return an error

## Node Configuration

An OPC UA node ID may resemble: "ns=3;s=Temperature". In this example:

- ns=3 is indicating the `namespace` is 3
- s=Temperature is indicting that the `identifier_type` is a string and
  `identifier` value is 'Temperature'
- This example temperature node has a value of 79.0

To gather data from this node enter the following line into the 'nodes'
property above:

```text
{field_name="temp", namespace="3", identifier_type="s", identifier="Temperature"},
```

This node configuration produces a metric like this:

```text
opcua,id=ns\=3;s\=Temperature temp=79.0,Quality="OK (0x0)" 1597820490000000000
```

With 'DataType' entered in Additional Metrics, this node configuration
produces a metric like this:

```text
opcua,id=ns\=3;s\=Temperature temp=79.0,Quality="OK (0x0)",DataType="Float" 1597820490000000000
```

If the value is an array, each element is unpacked into a field
using indexed keys. For example:

```text
opcua,id=ns\=3;s\=Temperature temp[0]=79.0,temp[1]=38.9,Quality="OK (0x0)",DataType="Float" 1597820490000000000
```

### Namespace Index vs Namespace URI

OPC UA supports two ways to specify namespaces:

1. **Namespace Index** (`namespace`): An integer (0-3 or higher) that references
   a position in the server's namespace array. This is simpler but can change if
   the server is restarted or reconfigured.

2. **Namespace URI** (`namespace_uri`): A string URI that uniquely identifies
   the namespace. This is more stable across server restarts but requires the
   plugin to fetch the namespace array from the server to resolve the URI to an index.

**When to use namespace index:**

- For standard OPC UA namespaces (0 = OPC UA, 1 = Local Server)
- When namespace stability is not a concern
- For simpler configuration

**When to use namespace URI:**

- When you need consistent node references across server restarts
- For production environments where namespace indices might change
- When working with vendor-specific namespaces

**Example using namespace URI:**

```toml
[[inputs.opcua.nodes]]
  name = "ServerStatus"
  namespace_uri = "http://opcfoundation.org/UA/"
  identifier_type = "i"
  identifier = "2256"
```

This produces the same node ID internally as:

```toml
[[inputs.opcua.nodes]]
  name = "ServerStatus"
  namespace = "0"
  identifier_type = "i"
  identifier = "2256"
```

Note: You must specify either `namespace` or `namespace_uri`, not both.

## Group Configuration

Groups can set default values for the namespace (index or URI), identifier type,
and tags settings. The default values apply to all the nodes in the group. If a
default is set, a node may omit the setting altogether. This simplifies node
configuration, especially when many nodes share the same namespace or identifier
type.

The output metric will include tags set in the group and the node.  If
a tag with the same name is set in both places, the tag value from the
node is used.

This example group configuration has three groups with two nodes each:

```toml
  # Group 1
  [[inputs.opcua.group]]
    name = "group1_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group1_tag = "val1" }
    [[inputs.opcua.group.nodes]]
      name = "name"
      identifier = "1001"
      default_tags = { node1_tag = "val2" }
    [[inputs.opcua.group.nodes]]
      name = "name"
      identifier = "1002"
      default_tags = {node1_tag = "val3"}

  # Group 2
  [[inputs.opcua.group]]
    name = "group2_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group2_tag = "val3" }
    [[inputs.opcua.group.nodes]]
      name = "saw"
      identifier = "1003"
      default_tags = { node2_tag = "val4" }
    [[inputs.opcua.group.nodes]]
      name = "sin"
      identifier = "1004"

  # Group 3
  [[inputs.opcua.group]]
    name = "group3_metric_name"
    namespace = "3"
    identifier_type = "i"
    default_tags = { group3_tag = "val5" }
    nodes = [
      {name="name", identifier="1001"},
      {name="name", identifier="1002"},
    ]
```

### Server Certificate Trust

When connecting to OPC UA servers with self-signed certificates using
secure modes (Sign or SignAndEncrypt), you need to explicitly trust the
server's certificate. Use the `remote_certificate` option to specify the
path to the server's certificate file.

Most OPC UA servers provide their certificate through their management interface
or configuration directory. Consult your OPC UA server's documentation to locate
the certificate, typically found in the server's PKI (Public Key Infrastructure)
directory. Alternatively, you can export the certificate using OPC UA client tools.

## Connection Service

This plugin actively reads to retrieve data from the OPC server.
This is done every `interval`.

## Metrics

The metrics collected by this input plugin will depend on the
configured `nodes` and `group`.

## Example Output

```text
group1_metric_name,group1_tag=val1,id=ns\=3;i\=1001,node1_tag=val2 name=0,Quality="OK (0x0)" 1606893246000000000
group1_metric_name,group1_tag=val1,id=ns\=3;i\=1002,node1_tag=val3 name=-1.389117,Quality="OK (0x0)" 1606893246000000000
group2_metric_name,group2_tag=val3,id=ns\=3;i\=1003,node2_tag=val4 Quality="OK (0x0)",saw=-1.6 1606893246000000000
group2_metric_name,group2_tag=val3,id=ns\=3;i\=1004 sin=1.902113,Quality="OK (0x0)" 1606893246000000000
```
