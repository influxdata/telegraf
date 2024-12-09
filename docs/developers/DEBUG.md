# Debug

The following describes how to use the [delve][1] debugger with telegraf
during development. Delve has many, very well documented [subcommands][2] and
options.

[1]: https://github.com/go-delve/delve
[2]: https://github.com/go-delve/delve/blob/master/Documentation/usage/README.md

## CLI

To run telegraf manually, users can run:

```bash
go run ./cmd/telegraf --config config.toml
```

To attach delve with a similar config users can run the following. Note the
additional `--` to specify flags passed to telegraf. Additional flags need to
go after this double dash:

```bash
$ dlv debug ./cmd/telegraf -- --config config.toml
Type 'help' for list of commands.
(dlv)
```

At this point a user could set breakpoints and continue execution.

## Visual Studio Code

Visual Studio Code's [go language extension][20] includes the ability to easily
make use of [delve for debugging][21]. Check out this [full tutorial][22] from
the go extension's wiki.

A basic config is all that is required along with additional arguments to tell
Telegraf where the config is located:

```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${fileDirname}",
            "args": ["--config", "/path/to/config"]
        }
    ]
}
```

[20]: https://code.visualstudio.com/docs/languages/go
[21]: https://code.visualstudio.com/docs/languages/go#_debugging
[22]: https://github.com/golang/vscode-go/wiki/debugging

## GoLand

JetBrains' [GoLand][30] also includes full featured [debugging][31] options.

The following is an example debug config to run Telegraf with a config:

```xml
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="build &amp; run" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="telegraf" />
    <working_directory value="$PROJECT_DIR$" />
    <parameters value="--config telegraf.conf" />
    <kind value="DIRECTORY" />
    <package value="github.com/influxdata/telegraf" />
    <directory value="$PROJECT_DIR$/cmd/telegraf" />
    <filePath value="$PROJECT_DIR$" />
    <method v="2" />
  </configuration>
</component>
```

[30]: https://www.jetbrains.com/go/
[31]: https://www.jetbrains.com/help/go/debugging-code.html
