# Systemd Secret-Store Plugin

The `systemd` plugin allows to utilize credentials and secrets provided by
[systemd][] to the Telegraf service. Systemd ensures that only the intended
service can access the credentials for the live-time of this service. The
credentials appear as plaintext files but are stored encrypted with TPM2
protection if available (see [this article][systemd-descr] for details).

This plugin does not support setting the credentials. See the
[credentials management section](#credential-management) below for how to
setup systemd credentials and how to add credentials

**Note**: Secrets of this plugin are static and are not updated after startup.

## Requirements and caveats

This plugin requires **systemd version 250+** with a correctly set-up
credentials via [systemd-creds][] (see [setup section](#credential-management)).
Furthermore, provisioning of the created credentials must by enabled
via `LoadCredentialEncrypted` in the service file. This is the case for the
Telegraf service provided in this repository. It expects encrypted credentials
to be stored in `/etc/telegraf/credentials`.

It is important to note that credentials can only be created and used on
the **same machine** as decrypting the secrets requires the encryption
key *and* a key stored in TPM2. Therefore, creating credentials and then
copying it to another machine will fail!

Please be aware that, due to its nature, this plugin is **ONLY** available
when started as a service. It does **NOT** find any credentials when started
manually via the command line! Therefore, `secrets` commands should **not**
be used with this plugin.

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

## Configuration

```toml @sample.conf
# Secret-store to access systemd secrets
[[secretstores.systemd]]
  ## Unique identifier for the secretstore.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "my_secretstore"

  ## Path to systemd credentials directory
  ## This should not be required as systemd indicates this directory
  ## via the CREDENTIALS_DIRECTORY environment variable.
  # path = "${CREDENTIALS_DIRECTORY}"
```

Each Secret provided by systemd will be available as file under
`${CREDENTIALS_DIRECTORY}/<secret-name>` for the service. You will **not**
be able to see them as a regular user, not are those files accessible to other
services.

## Credential management

Most steps here are condensed from the [systemd-creds man-page][systemd-creds]
and are using this command. Please also check that man-page as the options
or verbs used here might be outdated for the systemd version you are using.

**Please note**: We are using `/etc/telegraf/credentials` as our storage
location for encrypted credentials throughout the examples below. This is
because the Telegraf service file expects credentials to be located there
by default. Furthermore, we assume the secret-store ID to be set to `syscreds`
in the examples.

### Setup

If you have not done it already, systemd requires a first-time setup of the
credential system. If you are planning to use the TPM2 chip of your system
for protecting the credentials, you should first make sure that it is
available using

```shell
# sudo systemd-creds has-tpm2
partial
-firmware
+driver
+system
+subsystem
```

The output should look similar to the above.

Now setup the credentials by creating the root key.

```shell
# sudo systemd-creds setup
Credential secret file '/var/lib/systemd/credential.secret' is not located on encrypted media, using anyway.
4096 byte credentials host key set up.
```

The warning only appears if you are storing the generated key on an unencrypted
disk which is not recommended. With this, we are all set to create credentials.

### Creating credentials

After setting up the encryption key we can create a new credential using

```shell
# echo -n "john-doe-jr" | systemd-creds encrypt - /etc/telegraf/credentials/http_user
```

You should now have a file named `http_user` that contains the encrypted
username. Please note that systemd credentials are named, so the name
`http_user` is also stored in the file. The secret-store later provides
the secret using this name as the secret's key.

You can explicitly name credentials using the `--name` parameter, however,
in the interest of simplicity, you should follow the filename equals the
credential/secret's name rule.

We can now add more secrets. To avoid potentially leaking the plain-text
credentials through command-history or showing it on the screen we use

```shell
systemd-ask-password -n | | systemd-creds encrypt - /etc/telegraf/credentials/http_password
```

to interactively enter the password.

### Enabling credentials in the service

TODO

### Using credentials as secrets

TODO

### Chaining for unattended start

TODO

### Troubleshooting

TODO

[systemd]: https://www.freedesktop.org/wiki/Software/systemd/
[systemd-descr]: https://systemd.io/CREDENTIALS
[systemd-creds]: https://www.freedesktop.org/software/systemd/man/systemd-creds.html
