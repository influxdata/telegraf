# WhaTap Output Plugin

This plugin writes to the WhaTap(https://www.whatap.io) APM(via TCP).

```toml
## You can create a project on the WhaTap site(https://www.whatap.io) 
  ## to get license, project code and server IP information.

  ## WhaTap license. Required
  #license = "xxxx-xxxx-xxxx"

  ## WhaTap project code. Required
  #pcode = 1111

  ## WhaTap server IP. Required
  # Put multiple IPs with / as delimiters. e.g. "1.1.1.1/2.2.2.2"
  #server = "1.1.1.1"

  # WhaTap base port 
  # port = 6600

  ## Connection timeout.
  # timeout = "60s"
```