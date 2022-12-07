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
|---------|-----------------------------------------------|
|`config` |print out full sample configuration to stdout|
|`secrets`|manage secret-store secrets|
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

## Secrets management

You can use telegraf to manage the secrets in the configured secret-stores.
Please make sure you specify a config containing the secret store configuration!

To list all available secret-stores with all known secret *keys* run

`telegraf secrets list`

You can also specify a secret-store ID and only get the keys for that store

`telegraf secrets list someid`

The above command will now list the *keys* of all secrets in the
secret-store with ID `someid`. To also reveal the *values* of all
the secrets use

`telegraf secrets list --reveal-secret someid`

You can also pass a list of secret-stores and the command will
print all secrets in those stores.

To access the *value* of a specific secret you can use

`telegraf secrets get someid a_secret_key`

to output the *value* of the secret `a_secret_key` in the secret-store
with ID `someid`.

All commands above will *read* secrets stored in the given store(s). To add
or modify keys use the `set` command

`telegraf secrets set someid a_secret_key the_new_value`

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
