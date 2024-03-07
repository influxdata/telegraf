# OS Secret-store Plugin

The `os` plugin allows to manage and store secrets using the native Operating
System keyring. For Windows this plugin uses the credential manager, on Linux
the kernel keyring is used and on MacOS we use the Keychain implementation.

To manage your secrets you can either use Telegraf or the tools that natively
comes with your operating system. Run

```shell
telegraf secrets help
```

to get more information on how to do this with Telegraf.

## Usage <!-- @/docs/includes/secret_usage.md -->

Secrets defined by a store are referenced with `@{<store-id>:<secret_key>}`
the Telegraf configuration. Only certain Telegraf plugins and options of
support secret stores. To see which plugins and options support
secrets, see their respective documentation (e.g.
`plugins/outputs/influxdb/README.md`). If the plugin's README has the
`Secret-store support` section, it will detail which options support secret
store usage.

## Configuration

The configuration differs slightly depending on the Operating System. We first
describe the common options here and the refer to the individual interpretation
or options in the following sections.

All secret-store implementations require an `id` to be able to reference the
store when specifying the secret. The `id` needs to be unique in the
configuration.

For all operating systems, the keyring name can be chosen using the `keyring`
parameter. However, the interpretation is slightly different on the individual
implementations.

The `dynamic` flag allows to indicate secrets that change during the runtime of
Telegraf. I.e. when set to `true`, the secret will be read from the secret-store
on every access by a plugin. If set to `false`, all secrets in the secret store
are assumed to be static and are only read once at startup of Telegraf.

```toml @sample.conf
# Operating System native secret-store
[[secretstores.os]]
  ## Unique identifier for the secret-store.
  ## This id can later be used in plugins to reference the secrets
  ## in this secret-store via @{<id>:<secret_key>} (mandatory)
  id = "secretstore"

  ## Keyring Name & Collection
  ## * Linux: keyring name used for the secrets, collection is unused
  ## * macOS: keyring specifies the macOS' Keychain name and collection is an
  ##     optional Keychain service name
  ## * Windows: keys follow a fixed pattern in the form
  ##     `<collection>:<keyring>:<key_name>`. Please keep this in mind when
  ##     creating secrets with the Windows credential tool.
  # keyring = "telegraf"
  # collection = ""

  ## macOS Keychain password
  ## If no password is specified here, Telegraf will prompt for it at startup
  ## time.
  # password = ""

  ## Allow dynamic secrets that are updated during runtime of telegraf
  # dynamic = false
```

### Linux

On Linux the kernel keyring in the `user` scope is used to store the
secrets. The `collection` setting is ignored on Linux.

### MacOS

On MacOS the Keychain implementation is used. Here the `keyring` parameter
corresponds to the Keychain name and the `collection` to the optional Keychain
service name. Additionally a password is required to access the Keychain.
The `password` itself is also a secret and can be a string, an environment
variable or a reference to a secret stored in another secret-store.
If `password` is omitted, you will be prompted for the password on startup.

### Windows

On Windows you can use the Credential Manager in the Control Panel or
[Telegraf](../../../cmd/telegraf/README.md) to manage your secrets.

If using the Credential Manager, click "Windows Credentials" and then
"Add a generic credential" with the following:

* _Internet or network address_: Enter the secret name in the format of:
  `<collection>:<keyring>:<key_name>`
* _User name_: Use `telegraf`. This field is not used, but needs something
  entered.
* _Password_: The actual secret value

If using Telegraf, see the help output of `telegraf secrets set` to add
secrets. Again use the `<collection>:<keyring>:<key_name>` format of the secret
key name.

### Docker

Access to the kernel keyring is __disabled by default__ in docker containers
(see [documentation](https://docs.docker.com/engine/security/seccomp/)).
In this case you will get an
`opening keyring failed: Specified keyring backend not available` error!

You can enable access to the kernel keyring, but as the keyring is __not__
namespaced, you should be aware of the security implication! One implication
is for example that keys added in one container are accessible by __all__
other containers running on the same host, not only within the same container.

### systemd-nspawn

The memguard dependency that Telegraf uses to secure memory for secret storage
requires the `CAP_IPC_LOCK` capability to correctly lock memory. Without this
capability Telegraf will panic. Users will need to start a container with the
`--capability=CAP_IPC_LOCK` flag for telegraf to correctly work.

See [github.com/awnumar/memguard#144][memguard-issue] for more information.

[memguard-issue]: https://github.com/awnumar/memguard/issues/144

### Windows

The memguard dependency that Telegraf uses to secure memory for secret storage
requires the `CAP_IPC_LOCK` capability to correctly lock memory. Without this
capability Telegraf will panic. Users will need to start a container with the
`--capability=CAP_IPC_LOCK` flag for telegraf to correctly work.

See [github.com/awnumar/memguard#144][memguard-issue] for more information.

[memguard-issue]: https://github.com/awnumar/memguard/issues/144
