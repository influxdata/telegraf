# Labels and Selectors for Plugin Enablement

## Objective

Introduce a label and selector system to enable or disable plugins dynamically
in Telegraf.

## Keywords

configuration, dynamic plugin selection

## Overview

Currently, managing plugin configurations across multiple Telegraf instances is
cumbersome. Methods like commenting out plugins or renaming configuration files
are not scalable. A label and selector system provides a more flexible and
elegant solution.

This feature aims to simplify plugin management in Telegraf by introducing a
label and selector system inspired by [Kubernetes][k8s_labels]. Selectors are
key-value pairs provided at runtime, and labels are defined in plugin
configurations to match against these selectors. This approach allows to enable
a plugin dynamically at startup-time, making it easier to manage configurations
in large-scale deployments where a single configuration is fetched from a
centralized configuration-source.

[k8s_labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

## Command line flags

The Telegraf executable must accept one or more optional `--select` command-line
flags. The passed value must be of the form:

```text
<key>=<value>[,<key>=<value>]
```

The `key` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores
(`_`).

The `value` part might contain the wildcard characters asterix (`*`) for
matching any number of characters or question mark (`?`) for matching a single
character. Furthermore the value may contain alpha-numerical values
(`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `key` and `value` parts of a selector are separated by an equal sign (`=`).
Multiple key-value pairs in a single selector are separated by commas.

For example, users can start the telegraf instance with the following command
line:

```console
telegraf --config config.conf --config-directory directory/ --select="app=payments,region=us-*" --select="env=prod" --watch-config --print-plugin-config-source=true
```

Specifying the same `key` multiple times within a single `--select` statement
causes an error at Telegraf startup.
However, the same `key` can be used in _different_ `--select` statements.

## Plugin Labels

Telegraf must implement a new optional `labels` configuration setting. This
setting must be available in all input, output, aggregator and processor
plugins. The `labels` configuration setting must accept a map where each entry
is a single key-value pair.

The `key` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores
(`_`).

The `value` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores
(`_`).

```toml
[[inputs.cpu]]
  [inputs.cpu.labels]
    app = "payments"
    region = "us-east"
    env = "prod"
```

Telegraf must provide the setting without changes to existing or new plugins.

> [!NOTE]
> Due to limitations in the TOML format, maps must be defined _after_ top-level
> plugin-settings e.g. at the end of the plugin configuration!

## Selection matching

Telegraf must match command-line selectors against the plugin labels to
determine if a plugin should be enabled. The matching behavior is as follows:

Multiple `--select` command-line parameters are treated as a logical **OR**
condition. If any select statement matches, the plugin will be enabled.
Within each `--select` command-line parameter, multiple key-value pairs,
separated by comma, are treated as a logical **AND** condition. All conditions
within that select statement must match for a plugin to be selected.

Selectors support exact matching as well as wildcard matching
using `*` (multiple characters) and `?` (single character) in the selector
values. The key part does not support wildcards.

### Behavior Matrix

| **Telegraf Run State** | **Label Present** | **Behavior**                                                                         |
| ---------------------- | ----------------- | ------------------------------------------------------------------------------------ |
| With `--select`        | Yes               | Plugin is selected if the selector matches the label. Otherwise, it is not selected. |
| With `--select`        | No                | Plugin is selected (backward compatibility).                                         |
| Without `--select`     | Yes               | Plugin is selected (no selector to compare against).                                 |
| Without `--select`     | No                | Plugin is selected (current behavior).                                               |

### Matching Examples

| CLI Selectors (`--select`)                           | Plugin Labels                             | Matching Behavior                                                                                                        | Result   |
| ---------------------------------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | -------- |
| `app=web`                                            | `app="web"`                               | Selector requires `app=web`; plugin label has `app=web`, so it matches                                                   | Selected |
| `app=web`                                            | `app="api"`                               | Selector requires `app=web`, but plugin has `app=api`; no match                                                          | Skipped  |
| `app=web`                                            | `app="web", region="us-east"`             | Selector only cares about `app=web`, which is present; extra labels are ignored                                          | Selected |
| `app=web,region=us-east`                             | `app="web", region="us-east"`             | Selector requires both `app=web` and `region=us-east`; both are present                                                  | Selected |
| `app=web,region=us-west`                             | `app="web", region="us-east"`             | Selector requires `region=us-west`, but plugin has `region=us-east`; mismatch                                            | Skipped  |
| `env=prod*`                                          | `env="production"`                        | Selector wants `env` starting with `prod`; `production` matches the wildcard                                             | Selected |
| `env=prod*`                                          | `env="staging"`                           | Selector wants `env` starting with `prod`, but `staging` doesn't match                                                   | Skipped  |
| `env=*`                                              | `env="qa"`                                | Wildcard `*` matches any value of `env`; `qa` satisfies it                                                               | Selected |
| `app=web,env=prod`, `region=eu-*`                    | `app="web", env="prod"`                   | First selector requires `app=web` AND `env=prod`; both are present, second selector is ignored                           | Selected |
| `app=web,env=prod`, `region=eu-*`                    | `app="web", env="staging", region="us"`   | First selector fails due to `env=staging`; second selector fails due to `region=us`; no match                            | Skipped  |
| `env=prod`, `env=staging`                            | `env="prod"`                              | At least one selector (`env=prod`) matches label exactly                                                                 | Selected |
| `env=prod`, `env=staging`                            | `env="qa"`                                | Neither `env=prod` nor `env=staging` match `env=qa`                                                                      | Skipped  |
| `app=web,env=prod`, `app=api,env=prod`               | `app="api", env="prod"`                   | Second selector requires `app=api` AND `env=prod`; both are present in plugin labels                                     | Selected |
| `app=web,env=test*,region=eu-west`, `app=*,env=test` | `app="web", env="test"`                   | First selector requires `app=web` AND `env=test*` AND `region=eu-west`, but `region` is missing; second selector matches | Selected |

## Related Issues

- [issue #1317](https://github.com/influxdata/telegraf/issues/1317)
  for allowing to enable/disable plugin instances
- [issue #9304](https://github.com/influxdata/telegraf/issues/9304)
  for partially enabling a config file
- [issue #10543](https://github.com/influxdata/telegraf/issues/10543)
  for allowing to enable/disable plugin instances
