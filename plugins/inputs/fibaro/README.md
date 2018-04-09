# Fibaro Input Plugin

The Fibaro plugin makes HTTP calls to the Fibaro controller API to gather values of hooked devices.  
Those values could be true (1) or false (0) for switches, percentage for dimmers, temperature, etc.


### Configuration:

```toml
# Read devices value(s) from a Fibaro controller
[[inputs.fibaro]]
  ## Required Fibaro controller address/hostname.
  ## Note: connection is done over http/80 as Fibaro did not implement https.
  server = "<controller>"
  ## Required credentials to access the API (http://<controller/api/<component>)
  username = "<username>"
  password = "<password>"
```


### Tags:

	section: section's name
	room: room's name
	name: device's name
	type: device's type


### Fields:

	value float
	value2 float (when available from device)
