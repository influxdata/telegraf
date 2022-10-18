# Telegraf Commands & Flags

The following page describes some of the commands and flags available via the
Telegraf command line interface.

## Usage

General usage of Telegraf, requires passing in at least one config file with
the plugins the user wishes to use:

```bash
telegraf --config config.toml
```

## Help

|command|description|
|--------|-----------------------------------------------|
|`config` |print out full sample configuration to stdout|
|`secret` |manage secret-store secrets|
|`version`|print the version to stdout|

To get the full list of subcommands and flags run:

```bash
telegraf help
```

Here are some commonly used flags that users should be aware of:

* `--config-directory`: Read all config files from a directory
* `--debug`: Enable additional debug logging
* `--once`: Run one collection and flush interval then exit
* `--test`: Run only inputs, output to stdout, and exit

Check out the full help out for more available flags and options.

## Version

While telegraf will print out the version when running, if a user is uncertain
what version their binary is, run the version subcommand:

```bash
telegraf version
```

## Config

The config subcommand allows users to print out a sample configuration to
stdout. This subcommand can very quickly print out the default values for all
or any of the plugins available in Telegraf.

For example to print the example config for all plugins run:

```bash
telegraf config > telegraf.conf
```

If a user only wanted certain inputs or outputs, then the filters can be used:

```bash
telegraf config --input-filter cpu --output-filter influxdb
```

**Run telegraf with pprof:**

`telegraf --config telegraf.conf --pprof-addr localhost:6060`

## Secret-store management

You can use telegraf to manage the secrets in the configured secret-stores.
Please make sure you specify a config containing the secret store
configuration!

To list all available secret-stores run

`telegraf list-secretstores`

which will print a list of all known secret-store IDs that can be used for
accessing the secrets in that store

`telegraf list-secrets someid`

The above command will now list the *keys* of all secrets in the
secret-store with ID `someid`. To also reveal the *values* of all
the secrets use

`telegraf list-secrets --reveal-secret someid`

You can also pass a list of secret-stores and the command will
print all secrets in those stores. If no secret-store ID is provided, i.e.

`telegraf list-secrets`

the command will list the *keys* of all secrets in all known secret-stores.

To access the *value* of a secret you can use the `get-secret` command

`telegraf get-secret someid a_secret_key`

to output the *value* of the secret `a_secret_key` in the secret-store
with ID `someid`.

All commands above will *read* secrets stored in the given store(s). To add
or modify keys use the `set-secret` command

`telegraf set-secret someid a_secret_key the_new_value`

to add or overwrite a secret named `a_secret_key` with the value
`the_new_value` in the secret-store with ID `someid`. If the secret with
the given key did not exist, it will be created.

You may now reference the secret in your telegraf configuration anywhere
you would normally use a password or secret. If your configuration has e.g.

```toml
  token = "<your secret token>"
```

 ou can replace it with a reference in the form
`@{<secret store id>:<secret name>}`. For the example above you would write

```toml
  token = @{someid:a_secret_key}
```

to reference the secret with the key `a_secret_key` of the secret-store with
ID `someid`.
