# Labels and Selectors for Plugin Enablement

## Objective

Introduce a label and selector system to enable or disable plugins dynamically in Telegraf.

## Overview

This feature aims to simplify plugin management in Telegraf by introducing a label and selector system inspired by Kubernetes. Labels are key-value pairs provided at runtime, and selectors are defined in plugin configurations to match against these labels. This approach allows for dynamic plugin enablement, making it easier to manage configurations in large-scale deployments.

### Motivation

Currently, managing plugin configurations across multiple Telegraf instances is cumbersome. Methods like commenting out plugins or renaming configuration files are not scalable. A label and selector system provides a more flexible and elegant solution.

### Use Case

- Deploy multiple Telegraf instances fetching configurations from a centralized source.
- Activate only the plugins relevant to each instance based on labels and selectors.

## Keywords

Configuration Management, Enable and disable plugin, Labels and Selectors.

## Configuration Options

### Labels

Labels are key-value pairs provided at runtime and attached to the Telegraf instance.

At Telegraf startup, users can provide labels via a CLI flag:

```console
telegraf --config config.conf --config-directory directory/ --label="app=payments" --label="region=us-east" --watch-config --print-plugin-config-source=true
```

### Selectors

A selector is a string describing one or more required key-value matches.

Each plugin definition in the configuration file can optionally include a selectors field.

Example:

```toml
[[inputs.cpu]]
  selectors = ["app=payments,region=us-east", "env=prod"]
```

### Boolean behaviour

- Multiple selectors are treated as **OR**

- Within a single selector string, multiple conditions separated by commas are treated as **AND**

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

### Behavior Matrix

| **Telegraf Run State** | **Selector Present** | **Behavior**                                                                         |
| ---------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| With `--label`         | Yes                  | Plugin is selected if the selector matches the label. Otherwise, it is not selected. |
| With `--label`         | No                   | Plugin is selected (backward compatibility).                                         |
| Without `--label`      | Yes                  | Plugin is selected (no label to compare against).                                    |
| Without `--label`      | No                   | Plugin is selected (current behavior).                                               |

## Previous Issues

[Issue #9304](https://github.com/influxdata/telegraf/issues/9304)

[Issue #1317](https://github.com/influxdata/telegraf/issues/1317)

[Issue #10543](https://github.com/influxdata/telegraf/issues/10543)
