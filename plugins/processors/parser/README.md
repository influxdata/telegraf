# Parser Processor Plugin
This plugin parses through fields in a pre-formatted string

## Configuration
```var SampleConfig = `

[processors.parser]
  ## specify the name of the field[s] whose value will be parsed
  parse_fields = []

  data_format = "logfmt"
  ## additional configurations for parser go here
`
```
