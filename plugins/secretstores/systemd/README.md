# Systemd Secret-Store Plugin

The `systemd` plugin allows utilizing credentials and secrets provided by
[systemd][] to the Telegraf service. Systemd ensures that only the intended
service can access the credentials for the lifetime of this service. The
credentials appear as plaintext files to the consuming service but are stored
encrypted on the host system. This encryption can also use TPM2 protection if
available (see [this article][systemd-descr] for details).

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
  id = "systemd"

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
location for encrypted credentials throughout the examples below assuming a
Telegraf install via package manager. If you are using some other means to
install Telegraf you might need to create that directory.
Furthermore, we assume the secret-store ID to be set to `systemd` in the
examples.

Setting up systemd-credentials might vary on your distribution or version so
please also check the documentation there. You might also need to install
supporting packages such as `tpm2-tools`.

### Setup

If you have not done it already, systemd requires a first-time setup of the
credential system. If you are planning to use the TPM2 chip of your system
for protecting the credentials, you should first make sure that it is
available using

```shell
$ sudo systemd-creds has-tpm2
partial
-firmware
+driver
+system
+subsystem
```

The output should look similar to the above. If TPM2 is available on your system
credentials can also be tied to the device by utilizing TPM2 sealing.
See the  [systemd-creds man-page][systemd-creds] for details.

Now setup the credentials by creating the root key.

```shell
$ sudo systemd-creds setup
Credential secret file '/var/lib/systemd/credential.secret' is not located on encrypted media, using anyway.
4096 byte credentials host key set up.
```

The warning only appears if you are storing the generated key on an unencrypted
disk which is not recommended. With this, we are all set to create credentials.

### Creating credentials

After setting up the encryption key we can create a new credential using

```shell
$ echo -n "john-doe-jr" | sudo systemd-creds encrypt - /etc/telegraf/credentials/http_user
Credential secret file '/var/lib/systemd/credential.secret' is not located on encrypted media, using anyway.
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
$ systemd-ask-password -n | sudo systemd-creds encrypt - /etc/telegraf/credentials/http_password
Password: (press TAB for no echo)
systemd-ask-password -n | systemd-creds encrypt - /etc/telegraf/credentials/http_password
```

to interactively enter the password.

### Enabling credentials in the service

To actually provide credentials to the Telegraf service, you need to list them
in the service file. You can use

```shell
$ sudo systemctl edit telegraf
...
```

to overwrite parts of the service file. On some systems you need to create the
overriding directory `/etc/systemd/system/telegraf.service.d` first. The
resulting override can be found in
`/etc/systemd/system/telegraf.service.d/override.conf`. The following is an
example for the content of the file

```text
[Service]
LoadCredentialEncrypted=http_user:/etc/telegraf/credentials/http_user
LoadCredentialEncrypted=http_passwd:/etc/telegraf/credentials/http_password
```

This will load two credentials, named `http_user` and `http_passwd` which are
then accessible by Telegraf with those names. Please note that the names have
to match the names used during encryption of the credentials.

You can add an arbitrary list of credentials to the service as long as the name
is unique.

### Using credentials as secrets

To use the credentials as secrets you need to first instantiate a `systemd`
secret-store by adding

```toml
[[secretstores.systemd]]
  id = "systemd"
```

to your Telegraf configuration. Assuming the two example credentials
`http_user` and `http_passwd` you can now use those as secrets via

```toml
[[inputs.http]]
  urls = ["http://localhost/metrics"]
  username = "@{systemd:http_user}"
  password = "@{systemd:http_passwd}"

```

in your plugins.

### Chaining for unattended start

When using many secrets or when secrets need to be shared among hosts, listing
all of them in the service file might be cumbersome. Additionally, it is hard
to manually test Telegraf configurations with the `systemd` secret-store as
those secrets are only available when started as a service.

Here, secret-store chaining comes into play, denoting a setup where one
secret-store, in our case `secretstores.systemd`, is used to unlock another
secret-store (`secretstores.jose` in this example).

```toml
[[secretstores.systemd]]
  id = "systemd"

[[secretstores.jose]]
  id = "mysecrets"
  path = "/etc/telegraf/secrets"
  password = "@{systemd:initial}"
```

Here we assume that an `initial` credential was injected through the service
file. This `initial` secret is then used to unlock the `jose` secret-store
which might provide many different secrets backed by encrypted files.

Input and output plugins can the use the `jose` secrets (via `@{mysecrets:...}`)
to fill sensitive data such as usernames, passwords or tokens.

### Troubleshooting

Please always make sure your systemd version matches Telegraf's requirements,
i.e. you do have version 250 or later.

When not being able to start the service please check the logs. A common issue
is a mismatch between the name stored in the credential (given during
`systemd-creds encrypt`) does not match the one used in the
`LoadCredentialEncrypted` statement.

In case you are having trouble to reference credentials in Telegraf, you should
check what is available via

```shell
$ CREDENTIALS_DIRECTORY=/etc/telegraf/credentials sudo systemd-creds list
NAME          SECURE   SIZE PATH
-------------------------------------------------------------------
http_password insecure 146B /etc/telegraf/credentials/http_password
http_user     insecure 142B /etc/telegraf/credentials/http_user
```

Please note that Telegraf's secret management functionality is not helpful here
as credentials are *only* available to the systemd service, not via the command
line.

As you can see, the above list also provides the *name* of the credential which
is the key you need to reference the secret.

To get the actual value of a credential use

```shell
$ sudo systemd-creds decrypt /etc/telegraf/credentials/http_password -
whooohabooo
```

Please use the above command(s) with care as they do reveal the secret value
of the credential!

[systemd]: https://www.freedesktop.org/wiki/Software/systemd/
[systemd-descr]: https://systemd.io/CREDENTIALS
[systemd-creds]: https://www.freedesktop.org/software/systemd/man/systemd-creds.html
