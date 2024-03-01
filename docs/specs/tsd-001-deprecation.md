# Plugin and Plugin Option Deprecation

## Objective

Specifies the process of deprecating and removing plugins, plugin settings
including values of those settings or features.

## Keywords

procedure, removal, all plugins

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
Plugin "inputs.logparser" deprecated since version 1.15.0 and will be removed in 1.40.0: use 'inputs.tail' with 'grok' data format instead
```

Similar warnings will be shown when removing plugin options or option values.
This provides users with time to replace the deprecated plugin in their
configuration file.

After the shown release (`v1.40.0` in this case) the warning will be promoted
to an error preventing Telegraf from starting. The user now has to adapt the
configuration file to start Telegraf.

## Time frames and considerations

When deprecating parts of Telegraf, it is important to provide users with enough
time to migrate to alternative solutions before actually removing those parts.

In general, plugins, plugin options or option values should only be deprecated
if a suitable alternative exists! In those cases, the deprecations should
predate the removal by at least one and a half years. In current release terms
this corresponds to six minor-versions. However, there might be circumstances
requiring a prolonged time between deprecation and removal to ensure a smooth
transition for users.

Versions between deprecation and removal of plugins, plugin options or option
values, Telegraf must log a *warning* on startup including information about
the version introducing the deprecation, the version of removal and an
user-facing hint on suitable replacements. In this phase Telegraf should
operate normally even with deprecated plugins, plugin options or option values
being set in the configuration files.

Starting from the removal version, Telegraf must show an *error* message for
deprecated plugins present in the configuration including all information listed
above. Removed plugin options and option values should be handled as invalid
settings in the configuration files and must lead to an error. In this phase,
Telegraf should *stop running* until all deprecated plugins, plugin options and
option values are removed from the configuration files.

## Deprecation Process

The deprecation process comprises the following the steps below.

### File issue

In the filed issue you should outline which plugin, plugin option or feature
you want to deprecate and *why*! Determine in which version the plugin should
be removed.

Try to reach an agreement in the issue before continuing and get a sign off
from the maintainers!

### Submit deprecation pull-request

Send a pull request adding deprecation information to the code and update the
plugin's `README.md` file. Depending on what you want to deprecate this
comprises different locations and steps as detailed below.

Once the deprecation pull-request is merged and Telegraf is released, we have
to wait for the targeted Telegraf version for actually removing the code.

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

to `plugins/inputs/deprecations.go`. By doing this, Telegraf will show a
deprecation warning to the user starting from version `1.15.0` including the
`Notice` you provided. The plugin can then be remove in version `1.40.0`.

Additionally, you should update the plugin's `README.md` adding a paragraph
mentioning since when the plugin is deprecated, when it will be removed and a
hint to alternatives or replacements. The paragraph should look like this

```text
**Deprecated in version v1.15.0 and scheduled for removal in v1.40.0**:
Please use the [tail][] plugin with the [`grok` data format][grok parser]
instead!
```

#### Deprecating an option

To deprecate a plugin option, remove the option from the `sample.conf` file and
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
