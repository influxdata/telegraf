# TOML

Telegraf uses TOML as the configuration language. The following outlines a few
common questions and issues that cause questions or confusion.

## Reference and Validator

For all things TOML related, please consult the [TOML Spec][] and consider
using a TOML validator. In VSCode the [Even Better TOML][] extension or use the
[TOML Lint][] website to validate your TOML config.

[TOML Spec]: https://toml.io/en/v1.0.0
[Even Better TOML]: https://marketplace.visualstudio.com/items?itemName=tamasfe.even-better-toml
[TOML Lint]: https://www.toml-lint.com/

## Multiple TOML Files

TOML technically does not support multiple files, this is done as a convenience for
users.

Users should be aware that when Telegraf reads a user's config, if multiple
files or directories are read in, each file at a time and all
settings are combined as if it were one big file.

## Single Table vs Array of Tables

Telegraf uses a single agent table (e.g. `[agent]`) to control high-level agent
specific configurations. This section can only be defined once for all config
files and should be in the first file read in to take effect. This cannot be
defined per-config file.

Telegraf also uses array of tables (e.g. `[[inputs.file]]`) to define multiple
plugins. These can be specified as many times as a user wishes.

## In-line Table vs Table

In some cases, a configuration option for a plugin may define a table of
configuration options. Take for example, the ability to add arbitrary tags to
an input plugin:

```toml
[[inputs.cpu]]
  percpu = false
  totalcpu = true
  [inputs.cpu.tags]
    tag1 = "foo"
    tag2 = "bar"
```

User's should understand that these tables *must* be at the end of the plugin
definition, because any key-value pair is assumed to be part of that table. The
following demonstrates how this can cause confusion:

```toml
[[inputs.cpu]]
  totalcpu = true
  [inputs.cpu.tags]
    tag1 = "foo"
    tag2 = "bar"
  percpu = false  # this is treated as a tag to add, not a config option
```

Note TOML does not care about how a user indents the config or whitespace, so
the `percpu` option is considered a tag.

A far better approach to avoid this situation is to use inline table syntax:

```toml
[[inputs.cpu]]
  tags = {tag1 = "foo", tag2 = "bar"}
  percpu = false
  totalcpu = true
```

This way the tags value can go anywhere in the config and avoids possible
confusion.

## Basic String vs String Literal

In basic strings, signified by double-quotes, certain characters like the
backslash and double quote contained in a basic string need to be escaped for
the string to be valid.

For example the following invalid TOML, includes a Windows path with
unescaped backslashes:

```toml
path = "C:\Program Files\"  # this is invalid TOML
```

User's can either escape the backslashes or use a literal string, which is
signified by single-quotes:

```toml
path = "C:\\Program Files\\"
path = 'C:\Program Files\'
```

Literal strings return exactly what you type. As there is no escaping in literal
strings you cannot have an apostrophe in a literal string.
