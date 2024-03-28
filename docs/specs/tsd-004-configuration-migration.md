# Configuration Migration

## Objective

Provides a subcommand and framework to migrate configurations containing
deprecated settings to a corresponding recent configuration.

## Keywords

configuration, deprecation, telegraf command

## Overview

With the deprecation framework of [TSD-001](tsd-001-deprecation.md) implemented
we see more and more plugins and options being scheduled for removal in the
future. Furthermore, deprecations become visible to the user due to the warnings
issued for removed plugins, plugin options and plugin option values.

To aid the user in mitigating deprecated configuration settings this
specifications proposes the implementation of a `migrate` sub-command to the
Telegraf `config` command for automatically migrate the user's existing
configuration files away from the deprecated settings to an equivalent, recent
configuration. Furthermore, the specification describes the layout and
functionality of a plugin-based migration framework to implement migrations.

### `migrate` sub-command

The `migrate` sub-command of the `config` command should take a set of
configuration files and configuration directories and apply available migrations
to deprecated plugins, plugin options or plugin option-values in order to
generate new configuration files that do not make use of deprecated options.

In the process, the migration procedure must ensure that only plugins with
applicable migrations are modified. Existing configuration must be kept and not
be overwritten without manual confirmation of the user. This should be
accomplished by storing modified configuration files with a `.migrated` suffix
and leaving it to the user to overwrite the existing configuration with the
generated counterparts. If no migration is applied in a configuration file, the
command might not generate a new file and leave the original file untouched.

During migration, the configuration, plugin behavior, resulting metrics and
comments should be kept on a best-effort basis. Telegraf must inform the user
about applied migrations and potential changes in the plugin behavior or
resulting metrics. If a plugin cannot be automatically migrated but requires
manual intervention, Telegraf should inform the user.

### Migration implementations

To implement migrations for deprecated plugins, plugin option or plugin option
values, Telegraf must provide a plugin-based infrastructure to register and
apply implemented migrations based on the plugin-type. Only one migration per
plugin-type must be registered.

Developers must implement the required interfaces and register the migration
to the mentioned framework. The developer must provide the possibility to
exclude the migration at build-time according to
[TSD-002](tsd-002-custom-builder.md). Existing migrations can be extended but
must be cumulative such that any previous configuration migration functionality
is kept.

Resulting configurations should generate metrics equivalent to the previous
setup also making use of metric selection, renaming and filtering mechanisms.
In cases this is not possible, there must be a clear information to the user
what to expect and which differences might occur.
A migration can only be informative, i.e. notify the user that a plugin has to
manually be migrated and should point users to additional information.

Deprecated plugins and plugin options must be removed from the migrated
configuration.
