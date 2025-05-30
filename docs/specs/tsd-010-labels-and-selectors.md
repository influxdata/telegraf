# Labels and Selectors for Plugin Enablement

## Objective

Introduce a label and selector system to enable or disable plugins dynamically in Telegraf.

## Keywords

configuration Management, dynamic plugin selection

## Overview

Currently, managing plugin configurations across multiple Telegraf instances is cumbersome. Methods like commenting out plugins or renaming configuration files are not scalable. A label and selector system provides a more flexible and elegant solution.

This feature aims to simplify plugin management in Telegraf by introducing a label and selector system inspired by [Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/). Labels are key-value pairs provided at runtime, and selectors are defined in plugin configurations to match against these labels. This approach allows to enable plugin dynamically at startup-time, making it easier to manage configurations in large-scale deployments where a single configuration is fetched from a centralized configuration-source.

## Command line flags

The Telegraf executable must accept one or more optional `--label` command-line flags. The passed value must be of the form:

```text
<key>=<value>
```

The `key` and `value` parts must not contain any wildcard characters but only alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `key` and `value` parts of a label are separated by an equal sign (`=`).

For example, users can start the telegraf instance with the following command line:

```console
telegraf --config config.conf --config-directory directory/ --label="app=payments" --label="region=us-east" --watch-config --print-plugin-config-source=true
```

## Plugin Selectors

Telegraf must implement a new optional `selectors` configuration setting. This setting must be available in all input, output, aggregator and processor plugins. The `selectors` configuration setting must accept a list of strings containing one or more comma-separated key-value pairs in `"<key>=<value>[,<key>=<value>]"` strings.

where the `key` part must not contain any wildcard characters but only alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `value` part might contain the wildcard characters asterix (`*`) for matching any number of characters or question mark (`?`)  for matching a single character. Furthermore the value may contain alpha-numerical values (`[A-Za-z0-9]`), dots (`.`), dashes (`-`) or underscores (`_`).

The `key` and `value` parts of a plugin selector are separated by an equal sign (`=`).

Multiple selectors might be specified by providing a string list like-

```toml
[[inputs.cpu]]
  selectors = ["app=payments,region=us-*", "env=prod"]
```

Telegraf must provide the setting without changes to existing or new plugins.

## Selection matching

Telegraf must match plugin selectors against the provided command-line labels to determine if a plugin should be enabled. The matching behavior is as follows:

Multiple selectors (strings in the array) are treated as a logical **OR** condition. If any selector string matches, the plugin will be enabled. Within each selector string, multiple key-value pairs separated by commas are treated as a logical **AND** condition - all conditions within that selector must match for it to be considered successful.

Selectors support exact matching as well as wildcard matching using `*` (multiple characters) and `?` (single character) in the label values. The key part does not support wildcards.

### Behavior Matrix

| **Telegraf Run State** | **Selector Present** | **Behavior**                                                                         |
| ---------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| With `--label`         | Yes                  | Plugin is selected if the selector matches the label. Otherwise, it is not selected. |
| With `--label`         | No                   | Plugin is selected (backward compatibility).                                         |
| Without `--label`      | Yes                  | Plugin is selected (no label to compare against).                                    |
| Without `--label`      | No                   | Plugin is selected (current behavior).                                               |

### Matching Examples

| Labels Provided                         | Plugin Selectors                      | Matching Behavior                                     | Result   |
| :-------------------------------------- | :------------------------------------ | :---------------------------------------------------- | :------- |
| `app=web`, `region=us-east`             | `["app=web"]`                         | `app=web` matches exactly                             | Selected |
| `app=web`, `region=us-east`             | `["app=api"]`                         | `app=api` does not match                              | Skipped  |
| `app=web`, `region=us-east`             | `["app=web,region=us-east"]`          | Both `app=web` **AND** `region=us-east` match         | Selected |
| `app=web`, `region=us-east`             | `["app=web,region=us-west"]`          | `region=us-west` does **not** match                   | Skipped  |
| `app=web`, `region=us-east`, `env=prod` | `["app=web,env=prod", "region=eu-*"]` | First selector matches (`app=web` **AND** `env=prod`) | Selected |
| `app=worker`, `region=us-central`       | `["region=us-*"]`                     | Wildcard match on `region=us-central`                 | Selected |
| `app=worker`, `region=us-central`       | `["region=eu-*"]`                     | Wildcard mismatch                                     | Skipped  |
| `app=api`                               | `["app=web", "app=api"]`              | Second selector matches (`app=api`)                   | Selected |

## Previous Issues

[Issue #9304](https://github.com/influxdata/telegraf/issues/9304)

[Issue #1317](https://github.com/influxdata/telegraf/issues/1317)

[Issue #10543](https://github.com/influxdata/telegraf/issues/10543)
