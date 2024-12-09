# README.md linter

## Building

```shell
telegraf/tools/readme_linter$ go build .
```

## Running

Run readme_linter with the filenames of the readme files you want to lint.

```shell
telegraf/tools/readme_linter$ ./readme_linter <path to readme>
```

You can lint multiple filenames at once. This works well with shell globs.

To lint all the plugin readmes:

```shell
telegraf/tools/readme_linter$ ./readme_linter ../../plugins/*/*/README.md
```

To lint readmes for inputs starting a-d:

```shell
telegraf/tools/readme_linter$ ./readme_linter ../../plugins/inputs/[a-d]*/README.md
```
