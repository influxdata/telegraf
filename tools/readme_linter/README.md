# README.md linter

Run readme_linter with the filenames of the readme files you want to lint. For
example:

```shell
~/go/src/github.com/influxdata/telegraf$ tools/readme_linter/readme_linter plugins/inputs/file/README.md
```

You can lint multiple filenames at once. This works well with shell globs.

To lint all the plugin readmes:

```shell
~/go/src/github.com/influxdata/telegraf$ tools/readme_linter/readme_linter plugins/*/*/README.md
```

To lint readmes for inputs starting a-d:

```shell
~/go/src/github.com/influxdata/telegraf$ tools/readme_linter/readme_linter plugins/inputs/[a-d]*/README.md
```
