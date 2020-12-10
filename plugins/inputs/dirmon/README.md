# dirmon input plugin
This plugin monitors a directory and its subfolders for csv.gz files, processes available files and writes metrics to Influx database

## Configuration
[[inputs.dirmon]]
  ```
  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  
  data_format = "influx "
  
  [[inputs.dirmon.directory]]
    directory = '/path/to/dir'
    dir_include = [ '[subfolderToBeIncluded]$' ]    
    dir_exclude = [] 
    file_include = [ 'filesToBeMonitored']
    num_processors = 
    [inputs.dirmon.directory.metric_tag_regex]
    ## provide the regex for identifying the following
    ## measurement : nique name provided to each piece of equipment
    ## site : the site that is reporting this metric
    ## ext : the file extension to look for
      measurement = " "
	    site = " "
      ext = " "
  	[inputs.dirmon.directory.file_regex]
    ## provide the regex for identifying the following
    ## source : file source
    ## ext : file extension 
    ## subdir : subdirectory to monitor
    ## dir_prefix : prefix to directory path
    ## relative : relative file path
    ## filename : name of the file
      source = " "
      ext = " "
      subdir = " "
      dir_prefix = " "
      relative = " "
      filename = " "
    [inputs.dirmon.directory.file_tag_regex]
      measurement = " "
			site = " "
        
  ```  

  ## Metrics
  Metrics related to mining equipment is collected

  * Measurement : equipment
  * Tags : 
    * make : make of the equipment (e.g. CAT, KOMATSU)
    * measurement : unique name provided to each piece of equipment
    * model : model of equipment 
    * quality : state of the equipment 
    * site : the site that is reporting this metric
