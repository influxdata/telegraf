# Honeycomb Output Plugin

This plugin writes data to [Honeycomb.io](https://honeycomb.io).


### Configuration

```toml
  ## Honeycomb authentication token
  apiKey = "API_KEY"

  ## Dataset name in Honeycomb to send data to
  dataset = "my-dataset"  

  ## Special tags that will not get prefixed by the measurement name
  ## This should be set if you specified global tags, and it should include the list of all global tags + host
  ## Default value for this list is: host
  #specialTags = ["host"]
  
  ## Optional: the hostname for the Honeycomb API server
  #apiHost = "https://api.honeycomb.io/""
```

### Special Tags
This plugin will add the measurement name as a prefix to all fields and tags. Some tags should not receive 
this prefix, notably the default `host` tag. If you add any global tags to your telegraf config, they should
also be referenced in this property to avoid a prefix getting added to the tags.

If you specified tags for Environment (`env`) and Region (`region`), you would specify them as well as host in 
this property list.
```toml
  specialTags = ["host", "env", "region"]
```

