# IIS App Pool Input Plugin

Reports information about IIS app pools.

It requires running Telegraf with administrator privileges and  WebManagementTools feature should be enbaled.


### Configuration:

```toml
[[inputs.win_iisproc]]
  ## No need to add anything else
```

### Measurements & Fields:

- win_iisproc
    - mem : float
    - cpu : float

The `mem` is current value PM(k) of app pool's worker process
The `cpu` is he amount of processor time that the process has used on all processors, in seconds so it's not cpu percentage of worker process.


### Tags:

- All measurements have the following tag:
    - appPool
 

### Example Output:
```
iis_proc,host=WIN2008R2H401,appPool=DefaultAppPool mem=276463616,cpu=4.515625 1500040669000000000
```
