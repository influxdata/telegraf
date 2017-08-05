# Fail2ban Plugin

The fail2ban plugin gathers counts of failed and banned ip addresses from fail2ban.

This plugin run fail2ban-client command, and fail2ban-client require root access.
You have to grant telegraf to run fail2ban-client:

- Run telegraf as root. (deprecate)
- Configure sudo to grant telegraf to fail2ban-client.

### Using sudo

You may edit your sudo configuration with the following:

``` sudo
telegraf ALL=(root) NOPASSWD: /usr/bin/fail2ban-client status *
```

### Configuration:

``` toml
# Read metrics from fail2ban.
[[inputs.fail2ban]]
  ## fail2ban-client require root access.
  ## Setting 'use_sudo' to true will make use of sudo to run fail2ban-client.
  ## Users must configure sudo to allow telegraf user to run fail2ban-client with no password.
  ## This plugin run only "fail2ban-client status".
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
$ ./telegraf --config telegraf.conf --input-filter fail2ban --test
fail2ban,jail=sshd failed=5i,banned=2i 1495868667000000000
```
