
# Systemd Secret Store Plugin

This plugin allows utilizing credentials and secrets provided to the Telegraf
service by [systemd][systemd]. Systemd ensures that only the intended service
can access the credentials for the lifetime of this service. The credentials
appear as plaintext files to the consuming service but are stored encrypted on
the host system.
This encryption can also use TPM2 protection if available
(see [this article][systemd-descr] for details).

This plugin does not support setting the credentials. See the
[credentials management section](#credential-management) below for how to
setup systemd credentials and how to add credentials

> [!NOTE]
> This plugin requires systemd version 250+.

⭐ Telegraf v1.29.0
🏷️ system
💻 all

[systemd]: https://www.freedesktop.org/wiki/Software/systemd/
[systemd-descr]: https://systemd.io/CREDENTIALS

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret store support` section, it will detail which options support secret
store usage.

## Configuration

```toml @sample.conf
# Reading systemd credentials
[[secretstores.systemd]]
  ## Unique identifier for the secret store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret store via @{<id>:<secret_key>} (mandatory)
  id = "systemd"

  ## Path to systemd credentials directory
  ## This should not be required as systemd indicates this directory
  ## via the CREDENTIALS_DIRECTORY environment variable.
  # path = "${CREDENTIALS_DIRECTORY}"

  ## Prefix to remove from systemd credential-filenames to derive secret names
  # prefix = "telegraf."

```

Each Secret provided by systemd will be available as file under
`${CREDENTIALS_DIRECTORY}/<secret-name>` for the service. You will **not** be
able to see them as a regular, non-telegraf user. Credential visibility from
other systemd services is mediated by the `User=` and `PrivateMounts=`
service-unit directives for those services. See the
[systemd.exec man-page][systemd-exec] for details.

> [!IMPORTANT]
> To use `ImportCredential`, as done in the default service file, you
> need **systemd version 254+** otherwise you need to specify the credentials
> using `LoadCredentialEncrypted` in a service-override.

## Credential management

Most steps here are condensed from the [systemd-creds man-page][systemd-creds].
Please also check that man-page as the options or verbs used here might be
outdated for the systemd version you are using.

> [!NOTE]
> We are using `/etc/credstore.encrypted` as our storage location for encrypted
> credentials throughout the examples below and assuming a Telegraf install via
> package manager. If you are using some other means to install Telegraf you
> might need to create that directory.
> Furthermore, we assume the secret store ID to be set to `systemd` in the
> examples.

Setting up systemd-credentials might vary on your distribution or version so
please also check the documentation there. You might also need to install
supporting packages such as `tpm2-tools`.

[systemd-creds]: https://www.freedesktop.org/software/systemd/man/systemd-creds.html

### Setup

Systemd requires a first-time setup of the credential system. If you are
planning to use the TPM2 chip of your system for protecting the credentials,
you should first make sure that it is available using

```shell
sudo systemd-creds has-tpm2
```

The output should look similar to

```text
partial
-firmware
+driver
+system
+subsystem
```

If TPM2 is available on your system, credentials can also be tied to the device
by utilizing TPM2 sealing. See the [systemd-creds man-page][systemd-creds] for
details.

> [!IMPORTANT]
> When TPM2 sealing is used credentials can only be created and used on the
> **same machine** and cannot be copied to other machines. This is because the
> required decryption key is stored in TPM2.

Now setup the credentials by creating the root key.

```shell
sudo systemd-creds setup
```

A warning may appears if you are storing the generated key on an unencrypted
disk which is not recommended. With this, we are all set to create credentials.

### Creating credentials

After setting up the encryption key we can create a new credential using

```shell
echo -n "john-doe-jr" | sudo systemd-creds encrypt - /etc/credstore.encrypted/telegraf.http_user
```

You should now have a file named `telegraf.http_user` containing the encrypted
username. The secret store later provides the secret using this filename as the
secret's key.

> [!NOTE]
> Telegraf expects credential files to be prefixed with `telegraf.` and without
> a custom name setting (no `--name`). By default Telegraf strips the
> `telegraf.` prefix. If you are using a different prefix or no prefix at all
> you need to adapt the `prefix` setting!

We can now add more secrets. To avoid potentially leaking the plain-text
credentials through command-history or showing it on the screen we use

```shell
systemd-ask-password -n | sudo systemd-creds encrypt - /etc/credstore.encrypted/telegraf.http_password
```

to interactively enter the password.

> [!NOTE]
> Due to its nature, this plugin is **ONLY** available when Telegraf is started
> as a systemd service. It does **NOT** find any credentials when started
> manually via the command line! Therefore, `secrets` commands should **not**
> be used with this plugin.

### Using credentials as secrets

To use the credentials as secrets you need to first instantiate a `systemd`
secret store by adding

```toml
[[secretstores.systemd]]
  id = "systemd"
```

to your Telegraf configuration. Assuming the two example credentials
`http_user` and `http_password` you can now use those as secrets via

```toml
[[inputs.http]]
  urls = ["http://localhost/metrics"]
  username = "@{systemd:http_user}"
  password = "@{systemd:http_password}"

```

in your plugins.

> [!NOTE]
> Secrets of this plugin are static and are not updated after startup.

### Chaining for unattended start

When using many secrets or when secrets need to be shared among hosts, listing
all of them in the service file might be cumbersome. Additionally, it is hard
to manually test Telegraf configurations with the `systemd` secret store as
those secrets are only available when started as a service.

Here, secret store chaining comes into play, denoting a setup where one
secret store, in our case `secretstores.systemd`, is used to unlock another
secret store (`secretstores.jose` in this example).

```toml
[[secretstores.systemd]]
  id = "systemd"

[[secretstores.jose]]
  id = "mysecrets"
  path = "/etc/telegraf/secrets"
  password = "@{systemd:initial}"
```

Here we assume that an `initial` credential was injected through the service
file. This `initial` secret is then used to unlock the `jose` secret store
which might provide many different secrets backed by encrypted files.

Input and output plugins can the use the `jose` secrets (via `@{mysecrets:...}`)
to fill sensitive data such as usernames, passwords or tokens.

## Troubleshooting

Please always make sure your systemd version matches Telegraf's requirements.

When not being able to start the service please check the logs. A common issue
is using the `--name` option which does not work with systemd's
`ImportCredential` setting.
Check for a mismatch between the name stored in the credential (given during
`systemd-creds encrypt`) and the one used in the
`LoadCredentialEncrypted` statement.

In case you are having trouble referencing credentials in Telegraf, you should
check what is available via

```shell
CREDENTIALS_DIRECTORY=/etc/credstore.encrypted sudo systemd-creds list
```

for the example above you should see

```text
NAME                   SECURE   SIZE PATH
-------------------------------------------------------------------
telegraf.http_password insecure 146B /etc/credstore.encrypted/telegraf.http_password
telegraf.http_user     insecure 142B /etc/credstore.encrypted/telegraf.http_user
```

Remember to remove the `prefix` configured in your secret store from the `NAME`
column to get the secrets' `key`.

To get the actual value of a credential use

```shell
sudo systemd-creds decrypt /etc/credstore.encrypted/telegraf.http_password -
```

Please use the above command(s) with care as they do reveal the secret value
of the credential!

[systemd-exec]: https://www.freedesktop.org/software/systemd/man/systemd.exec.html

## Additional Information

This plugin only supports reading the secrets, it cannot create or modify them.
