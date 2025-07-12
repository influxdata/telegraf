# Labels and Selectors for Plugin Enablement

## Objective

Introduce a label and selector system to enable or disable plugins dynamically
in Telegraf.

## Keywords

configuration, dynamic plugin selection

## Overview

Currently, managing plugin configurations across multiple Telegraf instances is
cumbersome. Methods like commenting out plugins or renaming configuration files
are not scalable. A label and selector system provides a more flexible and elegant solution.

This feature aims to simplify plugin management in Telegraf by introducing a
label and selector system inspired by [Kubernetes][k8s_labels]. Selectors are key-value
pairs provided at runtime, and labels are defined in plugin configurations to
match against these selectors. This approach allows to enable a plugin dynamically at
startup-time, making it easier to manage configurations in large-scale deployments
where a single configuration is fetched from a centralized configuration-source.

[k8s_labels]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

## Command line flags

The Telegraf executable must accept one or more optional `--selector` command-line
flags. The passed value must be of the form:

```text
<key>=<value>[,<key>=<value>]
```

The `key` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `value` part might contain the wildcard characters asterix (`*`) for matching
any number of characters or question mark (`?`) for matching a single character.
Furthermore the value may contain alpha-numerical values (`[A-Za-z0-9]`),
dots (`.`), dashes (`-`) or underscores (`_`).

The `key` and `value` parts of a selector are separated by an equal sign (`=`).
Multiple key-value pairs in a single selector are separated by commas.

For example, users can start the telegraf instance with the following command line:

```console
telegraf --config config.conf --config-directory directory/ --selector="app=payments,region=us-*" --selector="env=prod" --watch-config --print-plugin-config-source=true
```

If there are same keys within a selector, the last occurrence of that key
will be used for matching.

For example, if a selector is defined as `--selector="app=web,region=us-east,app=web*"`,
the last occurrence of `app` i.e. `web*` will be considered.

## Plugin Labels

Telegraf must implement a new optional `labels` configuration setting. This
setting must be available in all input, output, aggregator and processor plugins.
The `labels` configuration setting must accept a list of strings with each string
containing a single key-value pair in the form `"<key>=<value>"`.

where the `key` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `value` part must not contain any wildcard characters but only
alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `key` and `value` parts of a plugin label are separated by an equal sign (`=`).

Multiple labels might be specified by providing a string list like-

```toml
[[inputs.cpu]]
  labels = ["app=payments", "region=us-east", "env=prod"]
```

Telegraf must provide the setting without changes to existing or new plugins.

If there are multiple labels with the same key, the last occurrence of that key
will be used for matching against the selectors.

For example, if a plugin has the following labels:

```toml
[[inputs.cpu]]
  labels = ["app=payments", "region=us-east", "app=payments-prod"]
```

the last occurrence of `app` i.e. `payments-prod` will be used for matching against
the selectors.

## Selection matching

Telegraf must match command-line selectors against the plugin labels to
determine if a plugin should be enabled. The matching behavior is as follows:

Multiple selectors (strings provided via command line) are treated as a logical **OR** condition.
If any selector string matches, the plugin will be enabled.
Within each selector string, multiple key-value pairs separated by commas are
treated as a logical **AND** condition - all conditions within that selector
must match for it to be considered successful.

Selectors support exact matching as well as wildcard matching
using `*` (multiple characters) and `?` (single character) in the selector values.
The key part does not support wildcards.

### Behavior Matrix

| **Telegraf Run State** | **Label Present** | **Behavior**                                                                         |
| ---------------------- | ----------------- | ------------------------------------------------------------------------------------ |
| With `--selector`      | Yes               | Plugin is selected if the selector matches the label. Otherwise, it is not selected. |
| With `--selector`      | No                | Plugin is selected (backward compatibility).                                         |
| Without `--selector`   | Yes               | Plugin is selected (no selector to compare against).                                 |
| Without `--selector`   | No                | Plugin is selected (current behavior).                                               |

### Matching Examples

| CLI Selectors (`--selector`)                         | Plugin Labels                             | Matching Behavior                                                                                                        | Result   |
| ---------------------------------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | -------- |
| `app=web`                                            | `["app=web"]`                             | Selector requires `app=web`; plugin label has `app=web`, so it matches                                                   | Selected |
| `app=web`                                            | `["app=api"]`                             | Selector requires `app=web`, but plugin has `app=api`; no match                                                          | Skipped  |
| `app=web`                                            | `["app=web", "region=us-east"]`           | Selector only cares about `app=web`, which is present; extra labels are ignored                                          | Selected |
| `app=web,region=us-east`                             | `["app=web", "region=us-east"]`           | Selector requires both `app=web` and `region=us-east`; both are present                                                  | Selected |
| `app=web,region=us-west`                             | `["app=web", "region=us-east"]`           | Selector requires `region=us-west`, but plugin has `region=us-east`; mismatch                                            | Skipped  |
| `env=prod*`                                          | `["env=production"]`                      | Selector wants `env` starting with `prod`; `production` matches the wildcard                                             | Selected |
| `env=prod*`                                          | `["env=staging"]`                         | Selector wants `env` starting with `prod`, but `staging` doesn't match                                                   | Skipped  |
| `env=*`                                              | `["env=qa"]`                              | Wildcard `*` matches any value of `env`; `qa` satisfies it                                                               | Selected |
| `app=web,env=prod`, `region=eu-*`                    | `["app=web", "env=prod"]`                 | First selector requires `app=web` AND `env=prod`; both are present, second selector is ignored                           | Selected |
| `app=web,env=prod`, `region=eu-*`                    | `["app=web", "env=staging", "region=us"]` | First selector fails due to `env=staging`; second selector fails due to `region=us`; no match                            | Skipped  |
| `env=prod`, `env=staging`                            | `["env=prod"]`                            | At least one selector (`env=prod`) matches label exactly                                                                 | Selected |
| `env=prod`, `env=staging`                            | `["env=qa"]`                              | Neither `env=prod` nor `env=staging` match `env=qa`                                                                      | Skipped  |
| `app=web,env=prod`, `app=api,env=prod`               | `["app=api", "env=prod"]`                 | Second selector requires `app=api` AND `env=prod`; both are present in plugin labels                                     | Selected |
| `app=web,env=test*,region=eu-west`, `app=*,env=test` | `["app=web", "env=test"]`                 | First selector requires `app=web` AND `env=test*` AND `region=eu-west`, but `region` is missing; second selector matches | Selected |

## Previous Issues

- [issue #1317](https://github.com/influxdata/telegraf/issues/1317) for allowing to enable/disable plugin instances
- [issue #9304](https://github.com/influxdata/telegraf/issues/9304) for partially enabling a config file
- [issue #10543](https://github.com/influxdata/telegraf/issues/10543) for allowing to enable/disable plugin instances
