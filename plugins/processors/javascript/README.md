# JavaScript Processor Plugin

The javascript processor is used if any other processor isn't enough for your needs.
If you need more powerful processor for your metrics, this is the right one for you.

### Configuration:

```toml
[[processors.javascript]]
  ## Code to run
  ##
  ## Path to JavaScript file or raw code string to run in the VM.
  ## If you put raw code, make sure to prepend "// javascript" as the first line of the code.
  code = "/path/to/file.js"

  ## Tags to set
  ##
  ## Set which tags from the metric to be set as variable
  ## in JavaScript VM.
  ## If tag is not found, it will be ignored and an error message will appear
  set_tags = [
	"mytag",
	"myothertag"
  ]

  ## Fields to set
  ##
  ## Set which fields from the metric to be set as variable
  ## in JavaScript VM.
  ## If field is not found, it will be ignored and an error message will appear.
  set_fields = [
	"myfield",
	"myotherfield"
  ]
	
  ## Tags to get
  ##
  ## Set to your target variable name, it must be a string
  ## and will replace original tags in the metric with the new values.
  ## If tag is not found, it will be ignored and an error message will appear.
  get_tags = [
	"mytag",
	"myothertag"
  ]

  ## Fields to get
  ##
  ## Consider JSON.stringify the variable to get if
  ## the value is not a boolean, string, float, or integer.
  ## It will be automatically unmarshalled into data type of your choice.
  ## Allowed data_type are:
  ## boolean, string, float, and integer.
  ## The value will then be resetted into the metric, replacing the original one.
  ## If field is not found, it will be ignored and an error message will appear.
  [[processors.javascript.get_fields]]
	name = "myfield",
	data_type = "string"
  [[processors.javascript.get_fields]]
	name = "myotherfield",
	data_type = "integer"
```

### Tags:

This processor does not add tags by default. 

### Fields:

This processor does not add fields by default. But the setting `[[processors.javascript.get_fields]]` with non-existent field name will add new field, 
providing it is exist as a variable in the script.
