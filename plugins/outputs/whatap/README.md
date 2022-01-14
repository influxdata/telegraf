# WhaTap Output Plugin

This plugin writes to the WhaTap(https://www.whatap.io) APM(via TCP).

```toml
  ## You can create a project on the WhaTap site(https://www.whatap.io) 
  ## to get license, project code and server IP information.

  ## WhaTap license. Required
  license = "xxxx-xxxx-xxxx"

  ## WhaTap project code. Required
  project_code = 1111

  ## WhaTap server IP. Required
  ## Put multiple IPs. ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]
  servers = ["tcp://1.1.1.1:6600","tcp://2.2.2.2:6600"]

  ## Connection timeout.
  # timeout = "60s"
```