# Postfix Input Plugin

The postfix plugin reports metrics on the postfix queues.

For each of the active, hold, incoming, maildrop, and deferred queues (http://www.postfix.org/QSHAPE_README.html#queues), it will report the queue length (number of items), size (bytes used by items), and age (age of oldest item in seconds).

### Configuration

```toml
[[inputs.postfix]]
  ## Postfix queue directory. If not provided, telegraf will try to use
  ## 'postconf -h queue_directory' to determine it.
  # queue_directory = "/var/spool/postfix"
```

#### Permissions:

Telegraf will need read access to the files in the queue directory.  You may
need to alter the permissions of these directories to provide access to the
telegraf user.

Unix permissions:
```sh
$ sudo chgrp -R telegraf /var/spool/postfix/{active,hold,incoming,deferred}
$ sudo chmod -R g+rXs /var/spool/postfix/{active,hold,incoming,deferred}
$ sudo usermod -a -G postdrop telegraf
$ sudo chmod g+r /var/spool/postfix/maildrop
```

Posix ACL:
```sh
$ sudo setfacl -Rdm u:telegraf:rX /var/spool/postfix/{active,hold,incoming,deferred,maildrop}
```

### Measurements & Fields:

- postfix_queue
    - length (integer)
    - size (integer, bytes)
    - age (integer, seconds)

### Tags:

- postfix_queue
    - queue

### Example Output

```
postfix_queue,queue=active length=3,size=12345,age=9
postfix_queue,queue=hold length=0,size=0,age=0
postfix_queue,queue=maildrop length=1,size=2000,age=2
postfix_queue,queue=incoming length=1,size=1020,age=0
postfix_queue,queue=deferred length=400,size=76543210,age=3600
```
