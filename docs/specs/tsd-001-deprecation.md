# Plugin and Plugin Option Deprecation

## Objective

Specifies the process of deprecating and removing plugins, plugin settings
including values of those settings or features.

## Keywords

procedure, framework, all plugins

## Overview

Over time the number of plugins, plugin options and plugin features grow and
some of those plugins or options are either not relevant anymore, have been
superseded or subsumed by other plugins or options. To be able to remove those,
this specification defines a process to deprecate plugins, plugin options and
plugin features including a timeline and minimal time-frames. Additionally, the
specification defines a framework to annotate deprecations in the code and
inform users about such deprecations.

## User experience

In the deprecation phase a warning will be shown at Telegraf startup with the
following content

```text
Plugin "inputs.logparser" deprecated since version 1.15.0 and will be removed in 2.0.0: use 'inputs.tail' with 'grok' data format instead
```

Similar warnings will be shown when removing plugin options or option values.
This provides users with time to replace the deprecated plugin in their
configuration file.

After the shown release (`v2.0.0` in this case) the warning will be promoted
to an error preventing Telegraf from starting. The user now has to adapt the
configuration file to start Telegraf.

## Deprecation Process

After reaching an agreement in the issue, you can start the deprecation process
by following the steps below.

### File issue

In the filed issue you should outline which plugin, plugin option or feature
you want to deprecate and *why*! Determine in which version the plugin should
be removed.
Consider moving the plugin to the Influx community repository as an external
plugin if there is no straight-forward replacement.

Try to reach an agreement in the issue before continuing and get a sign off
from the maintainers!

### Submit deprecation pull-request

Send a pull request adding deprecation information to the code and update the
plugin's `README.md` file. Depending on what you want to deprecate this
comprises different locations and steps as detailed below.

Once the deprecation pull-request is merged and Telegraf is released, we have
to wait for the targeted Telegraf version for actually removing the code.

You should consider writing a blog-post to make users aware of the scheduled
removal of plugins, plugin options or option values.

#### Deprecating a plugin

When deprecating a plugin you need to add an entry to the `deprecation.go` file
in the respective plugin category with the following format

```golang
    "<plugin name>": {
        Since:     "<x.y.z format version of the next minor release>",
        RemovalIn: "<x.y.z format version of the plugin removal>",
        Notice:    "<user-facing hint e.g. on replacements>",
    },
```

If you for example want to remove the `inputs.logparser` plugin you should add

```golang
    "logparser": {
        Since:     "1.15.0",
        RemovalIn: "1.40.0"
        Notice:    "use 'inputs.tail' with 'grok' data format instead",
    },
```

to `plugins/inputs.deprecations.go`. By doing this, Telegraf will show a
deprecation warning to the user starting from version `1.15.0` including the
`Notice` you provided. The plugin can then be remove in version `1.40.0`.

Additionally, you should update the plugin's `README.md` stating the plugin is
deprecated at the top (e.g. ***DEPRECATED***).

#### Deprecating an option

To deprecate a plugin open, remove the option from the `sample.conf` file and
add the deprecation information to the structure field in the code. If you for
for example want to deprecate the `ssl_enabled` option in `inputs.example` you
should add

```golang
type Example struct {
    ...
    SSLEnabled bool `toml:"ssl_enabled" deprecated:"1.3.0;1.40.0;use 'tls_*' options instead"`
}
```

to schedule the setting for removal in version `1.40.0`. The last element of
the `deprecated` tag is a user-facing notice similar to plugin deprecation.

#### Deprecating an option-value

Sometimes, certain option values become deprecated or superseded by other
options or values. To deprecate those option values, remove them from
`sample.conf` and add the deprecation info in the code if the deprecated value
is *actually used* via

```golang
func (e *Example) Init() error {
    ...
    if e.Mode == "old" {
        models.PrintOptionDeprecationNotice(telegraf.Warn, "inputs.example", "mode", telegraf.DeprecationInfo{
            Since:     "1.23.1",
            RemovalIn: "1.40.0",
            Notice:    "use 'v1' instead",
        })
    }
    ...
    return nil
}
```

This will show a warning if the deprecated `v1` value is used for the `mode`
setting in `inputs.example` with a user-facing notice.

### Submit pull-request for removing code

Once the plugin, plugin option or option-value is deprecated, we have to wait
for the `RemovedIn` release to remove the code. In the examples above, this
would be version `1.40.0`. After all scheduled bugfix-releases are done, with
`1.40.0` being the next release, you can create a pull-request to actually
remove the deprecated code.

Please make sure, you remove the plugin, plugin option or option value and the
code referencing those. This might also comprise the `all` files of your plugin
category, test-cases including those of other plugins, README files or other
documentation. For removed plugins, please keep the deprecation info in
`deprecations.go` so users can find a reference when switching from a really
old version.

Make sure you add an `Important Changes` sections to the `CHANGELOG.md` file
describing the removal with a reference to your PR.

### Time frames and considerations

When deprecating parts of Telegraf, it is important to provide users with enough
time to migrate to alternative solutions before actually removing those parts.

In general, plugins, plugin options or option values should only be deprecated
if a suitable alternative exists! In those cases, the deprecations should
predate the removal by at least one and a half year. In current release terms
this corresponds to six minor-versions. However, there might be circumstances
requiring a pro-longed time between deprecation and removal to ensure a smooth
transition for users.

## Deprecation Framework

Telegraf should provide a framework to notify users about deprecated plugins,
plugin options and option values including the version introducing the
deprecation (i.e. since when), the planned version for removal and a hint for
alternative approaches or  replacements.

This notification should be a log-message with *warning_*severity starting form
the deprecation version (since) up to the removal version. In this time,
Telegraf should operate normally even with deprecated plugins, plugin options
or option values being set in the configuration files. Starting from the removal
version, Telegraf will show an *error* message for the deprecated parts in the
configuration and stops running until those parts are removed.
