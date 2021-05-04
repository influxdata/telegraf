
# New Config API issues
x test that metrics flow through after startup and add/remove plugins
x review chan and ctx closure on shutdown vs delete plugin
x starting plugins in any order?
x change agent to use config api style exclusively.
  - change test/once to use config api style
v ability to modify existing plugin without re-creating it
x move configapi to a plugin
x make sure config and storage plugins are initialized from toml
x start config api in config plugin init() if configured
x loop processors when last input is removed and ctx not cancelled
x started plugins shouldn't run until the agent tells them to, or at least the metrics shouldn't flow until the agent says "go"

- close config/storage plugins on shutdown
- wait for all plugins to stop
- pull config out to its own package
- check that all inputs write to the input dst channel,
- add support for --test and --once

[done]  design api spec
[done]  generate JSON schema from structs
[done]  supporting setting config from JSON
[done]  build api package with plugin-control functionality
[done]  make aggregtors into processors, so that aggregators can now be ordered explicitly
[done]  default order from config
[done]  add config plugin support, connect endpoints to this?

[todo]  add REST interface to api package
[maybe] add gRPC interface to api package?
[todo]  support persistenting api configuration
[todo] migration tool to migrate to config api plugins


## out of scope

[future] separate toml config parsing from Telegraf config
[future] support config versioning to revert to older config versions and record changes
[future] support outputting toml configuration from the current config?

Config UI for Cloud support (out of current scope)
[future] config validation rules described in JSON
[future] web uses config schema to build config-editing forms dynamically
[future] web uses config validation rules to validate forms before submitting, displaying validation warnings

Service Discovery (out of current scope)
[future] service discovery (configuration plugin or stand-alone): monitors a service and issues config commands based on configuration templates
[future] error management?

Path to 2.0?
// Migration path from telegraf.toml to api_service.toml
// - use TOML config for defining agent/config settings, including [config.api]
//
// option 2: config can not define plugins, the api has full control over plugins.
//
// addendum a: Telegraf can always start and load plugins from a file. always good!!
// addendum b: can export running plugins as TOML
//
// telegraf --config=old.toml --migrate-to-api // or add [config.api] to toml:
// output: I've copied your old.toml into two new files:
//         api_service.toml, old_plugins.toml.
// 		you won't need plugins.toml to run the API as all plugins are loaded dynamically
// 		do you want to load your previous plugins into telegraf? [y/n]
// telegraf --config=api_service.toml --load-once-from=old_plugins.toml
//
//    To run telegraf normally, run telegraf --config=api_service.toml
//
// contents of api_service.toml are :
// [global_tags] ...
// [agent] ...
// [config.api] ...
