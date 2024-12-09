# State-persistence for plugins

## Purpose

Plugin state-persistence allows a plugin to save its state across restarts of
Telegraf. This might be necessary if data-input (or output) is stateful and
depends on the result of a previous operation.

If you for example query data from a service providing a `next` token, your
plugin would need to know the last token received in order to make the next
query. However, this token is lost after a restart of Telegraf if not persisted
and thus your only chance is to restart the query chain potentially resulting
in handling redundant data producing unnecessary traffic.

This is where state-persistence comes into play. The state-persistence framework
allows your plugin to store a _state_ on shutdown and load that _state_ again
on startup of Telegraf.

## State format

The _state_ of a plugin can be any structure or datatype that is serializable
using Golang's JSON serializer. It can be a key-value map or a more complex
structure. E.g.

```go
type MyState struct {
    CurrentToken string
    LastToken    string
    NextToken    string
    FilterIDs    []int64
}
```

would represent a valid state.

## Implementation

To enable state-persistence in your plugin you need to implement the
`StatefulPlugin` interface defined in `plugin.go`. The interface looks as
follows:

```go
type StatefulPlugin interface {
    GetState() interface{}
    SetState(state interface{}) error
}
```

The `GetState()` function should return the current state of the plugin
(see [state format](#state-format)). Please note that this function should
_always_ succeed and should always be callable directly after `Init()`. So make
sure your relevant data-structures are initialized in `Init` to prevent panics.

Telegraf will call the `GetState()` function on shutdown and will then compile
an overall Telegraf state from the information of all stateful plugins. This
state is then persisted to disk if (and only if) the `statefile` option in the
`agent` section is set. You do _not_ need take care of any serialization or
writing, Telegraf will handle this for you.

When starting Telegraf, the overall persisted Telegraf state will be restored,
if `statefile` is set. To do so, the `SetState()` function is called with the
deserialized state of the plugin. Please note that this function is called
directly _after_ the `Init()` function of your plugin. You need to make sure
that the given state is what you expect using a type-assertion! Make sure this
won't panic but rather return a meaningful error.

To assign the state to the correct plugin, Telegraf relies on a plugin ID.
See the ["State assignment" section](#state-assignment) for more details on
the procedure and ["Plugin Identifier" section](#plugin-identifier) for more
details on ID generation.

## State assignment

When restoring the state on loading, Telegraf needs to ensure that each plugin
_instance_ gets the correct state. To do so, a plugin ID is used. By default
this ID is generated automatically for each plugin instance but can be
overwritten if necessary (see [Plugin Identifier](#plugin-identifier)).

State assignment needs to be able to handle multiple instances of the same
plugin type correctly, e.g. if the user has configured multiple instances of
your plugin with different `server` settings. Here, the state saved for
`foo.example.com` needs to be restored to the plugin instance handling
`foo.example.com` on next startup of Telegraf and should _not_ end up at server
`bar.example.com`. So the plugin identifier used for the assignment should be
consistent over restarts of Telegraf.

In case plugin instances are added to the configuration between restarts, no
state is restored _for those instances_. Furthermore, all states referencing
plugin identifier that are no-longer valid are dropped and will be ignored. This
can happen in case plugin instances are removed or changed in ID.

## Plugin Identifier

As outlined above, the plugin identifier (plugin ID) is crucial when assigning
states to plugin instances. By default, Telegraf will automatically generate an
identifier for each plugin configured when starting up. The ID is consistent
over restarts of Telegraf and is based on the _entire configuration_ of the
plugin. This means for each plugin instance, all settings in the configuration
will be concatenated and hashed to derive the ID. The resulting ID will then be
used in both save and restore operations making sure the state ends up in a
plugin with _exactly_ the same configuration that created the state.

However, this also means that the plugin identifier _is changing_ whenever _any_
of the configuration setting is changed! For example if your plugin is defined
as

```go
type MyPlugin struct {
    Server  string          `toml:"server"`
    Token   string          `toml:"token"`
    Timeout config.Duration `toml:"timeout"`

    offset int
}
```

with `offset` being your state, the plugin ID will change if a user changes the
`timeout` setting in the configuration file. As a consequence the state cannot
be restored. This might be undesirable for your plugin, therefore you can
overwrite the ID generation by implementing the `PluginWithID` interface (see
`plugin.go`). This interface defines a `ID() string` function returning the
identifier o the current plugin _instance_. When implementing this function you
should take the following criteria into account:

1. The identifier has to be _unique_ for your plugin _instance_ (not only for
   the plugin type) to make sure the state is assigned to the correct instance.
1. The identifier has to be _consistent_ across startups/restarts of Telegraf
   as otherwise the state cannot be restored. Make sure the order of
   configuration settings doesn't matter.
1. Make sure to _include all settings relevant for state assignment_. In
   the example above, the plugin's `token` setting might or might not be
   relevant to identify the plugin instance.
1. Make sure to _leave out all settings irrelevant for state assignment_. In
   the example above, the plugin's `timeout` setting likely is not relevant
   for the state and can be left out.

Which settings are relevant for the state are plugin specific. For example, if
the `offset` is a property of the _server_ the `token` setting is irrelevant.
However, if the `offset` is specific for a certain user suddenly the `token`
setting is relevant.

Alternatively to generating an identifier automatically, the plugin can allow
the user to specify that ID directly in a configuration setting. However, please
not that this might lead to colliding IDs in larger setups and should thus be
avoided.
