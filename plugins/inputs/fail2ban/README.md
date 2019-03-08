# Fail2ban Input Plugin

The fail2ban plugin gathers the count of failed and banned ip addresses using [fail2ban](https://www.fail2ban.org).

This plugin runs the `fail2ban-client` command which generally requires root access.
Acquiring the required permissions can be done using several methods:

- Use sudo run fail2ban-client.
- Run telegraf as root. (not recommended)

### Using sudo

You will need the following in your telegraf config:
```toml
[[inputs.fail2ban]]
  use_sudo = true
```

You will also need to update your sudoers file:
```bash
$ visudo
# Add the following line:
Cmnd_Alias FAIL2BAN = /usr/bin/fail2ban-client status, /usr/bin/fail2ban-client status *
telegraf  ALL=(root) NOEXEC: NOPASSWD: FAIL2BAN
Defaults!FAIL2BAN !logfile, !syslog, !pam_session
```

### Configuration:

```toml
# Read metrics from fail2ban.
[[inputs.fail2ban]]
  ## Use sudo to run fail2ban-client
  use_sudo = false
```

### Measurements & Fields:

- fail2ban
  - failed (integer, count)
  - banned (integer, count)

### Tags:

- All measurements have the following tags:
  - jail

### Example Output:

```
# fail2ban-client status sshd
Status for the jail: sshd
|- Filter
|  |- Currently failed: 5
|  |- Total failed:     20
|  `- File list:        /var/log/secure
`- Actions
   |- Currently banned: 2
   |- Total banned:     10
   `- Banned IP list:   192.168.0.1 192.168.0.2
```

```
fail2ban,jail=sshd failed=5i,banned=2i 1495868667000000000
```
