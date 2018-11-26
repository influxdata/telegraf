# NSD Input Plugin

This plugin gathers stats from [NSD](https://nlnetlabs.nl/projects/nsd/about/) - an authoritative DNS server.

### Configuration

```toml
 # A plugin to collect stats from the NSD DNS server
 [[inputs.unbound]]
   ## Address of server to connect to, read from nsd conf default, optionally ':port'
   ## Will lookup IP if given a hostname
   server = "127.0.0.1:8952"
 
   ## If running as a restricted user you can prepend sudo for additional access:
   # use_sudo = false
 
   ## The default location of the nsd-control binary can be overridden with:
   # binary = "/usr/sbin/nsd-control"
 
   ## The default timeout of 1s can be overriden with:
   # timeout = "1s"
 
   ## When set to true, thread metrics are tagged with the thread id.
   ##
   ## The default is false for backwards compatibility, and will be change to
   ## true in a future version.  It is recommended to set to true on new
   ## deployments.
   thread_as_tag = false
```

#### Permissions:

It's important to note that this plugin references nsd-control, which may require additional permissions to execute successfully.
Depending on the user/group permissions of the telegraf user executing this plugin, you may need to alter the group membership, set facls, or use sudo.

**Group membership (Recommended)**:
```bash
$ groups telegraf
telegraf : telegraf

$ usermod -a -G nsd telegraf

$ groups telegraf
telegraf : telegraf nsd
```

**Sudo privileges**:
If you use this method, you will need the following in your telegraf config:
```toml
[[inputs.nsd]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
telegraf ALL=(ALL) NOPASSWD: /usr/sbin/nsd-control
```

Please use the solution you see as most appropriate.

### Metrics:

This is the full list of stats provided by nsd-control and potentially collected
depending of your nsd configuration. In the output, the dots in the nsd-control stat name are replaced by underscores(see
https://nlnetlabs.nl/documentation/nsd/nsd-control for details).

Shown metrics are with `thread_as_tag` enabled.