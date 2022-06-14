# Logging

## Plugin Logging

You can access the Logger for a plugin by defining a field named `Log`.  This
`Logger` is configured internally with the plugin name and alias so they do not
need to be specified for each log call.

```go
type MyPlugin struct {
    Log telegraf.Logger `toml:"-"`
}
```

You can then use this Logger in the plugin.  Use the method corresponding to
the log level of the message.

```go
p.Log.Errorf("Unable to write to file: %v", err)
```

## Agent Logging

In other sections of the code it is required to add the log level and module
manually:

```go
log.Printf("E! [agent] Error writing to %s: %v", output.LogName(), err)
```

## When to Log

Log a message if an error occurs but the plugin can continue working.  For
example if the plugin handles several servers and only one of them has a fatal
error, it can be logged as an error.

Use logging judiciously for debug purposes.  Since Telegraf does not currently
support setting the log level on a per module basis, it is especially important
to not over do it with debug logging.

If the plugin is listening on a socket, log a message with the address of the socket:

```go
p.log.InfoF("Listening on %s://%s", protocol, l.Addr())
```

## When not to Log

Don't use logging to emit performance data or other meta data about the plugin,
instead use the `internal` plugin and the `selfstats` package.

Don't log fatal errors in the plugin that require the plugin to return, instead
return them from the function and Telegraf will handle the logging.

Don't log for static configuration errors, check for them in a plugin `Init()`
function and return an error there.

Don't log a warning every time a plugin is called for situations that are
normal on some systems.

## Log Level

The log level is indicated by a single character at the start of the log
message.  Adding this prefix is not required when using the Plugin Logger.

- `D!` Debug
- `I!` Info
- `W!` Warning
- `E!` Error

## Style

Log messages should be capitalized and be a single line.

If it includes data received from another system or process, such as the text
of an error message, the text should be quoted with `%q`.

Use the `%v` format for the Go error type instead of `%s` to ensure a nil error
is printed.
