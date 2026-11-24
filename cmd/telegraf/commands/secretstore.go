package commands

import "github.com/urfave/cli/v2"

type SecretStoreActions struct {
	List   cli.ActionFunc
	Get    cli.ActionFunc
	Set    cli.ActionFunc
	Remove cli.ActionFunc
}

func secretstoreCommands(actions SecretStoreActions) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "secrets",
			Usage: "commands for listing, adding and removing secrets on all known secret stores",
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "list known secrets and secret stores",
					Description: `
The 'list' command requires passing in your configuration file
containing the secret store definitions you want to access. To get a
list of available secret store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.

For help on how to define secret stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to list the keys of all known secrets in ALL
available stores

> telegraf secrets list

To get the keys of all known secrets in a particular store, you can run

> telegraf secrets list mystore

To also reveal the actual secret, i.e. the value, you can pass the
'--reveal-secret' flag.
`,
					ArgsUsage: "[secret store ID]...[secret store ID]",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "reveal-secret",
							Usage: "also print the secret value",
						},
					},
					Action: actions.List,
				},
				{
					Name:  "get",
					Usage: "retrieves value of given secret from given store",
					Description: `
The 'get' command requires passing in your configuration file
containing the secret store definitions you want to access. To get a
list of available secret store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use the 'secrets list' command to get the IDs of available stores and
key(s) of available secrets.

For help on how to define secret stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to retrieve a secret from a secret store
available stores

> telegraf secrets get mystore mysecretkey

This will fetch the secret with the key 'mysecretkey' from the secret store
with the ID 'mystore'.
`,
					ArgsUsage: "<secret store ID> <secret key>",
					Action:    actions.Get,
				},
				{
					Name:  "set",
					Usage: "create or modify a secret in the given store",
					Description: `
The 'set' command requires passing in your configuration file
containing the secret store definitions you want to access. To get a
list of available secret store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use the 'secrets list' command to get the IDs of available stores and keys.

For help on how to define secret stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to create a secret in anm available secret store

> telegraf secrets set mystore mysecretkey mysecretvalue

This will create a secret with the key 'mysecretkey' in the secret store
with the ID 'mystore' with the value being set to 'mysecretvalue'. If a
secret with that key ('mysecretkey') already existed in that store, its
value will be modified.

When you leave out the value of the secret like

> telegraf secrets set mystore mysecretkey

you will be prompted to enter the value of the secret.
`,
					ArgsUsage: "<secret store ID> <secret key>",
					Action:    actions.Set,
				},
				{
					Name:  "remove",
					Usage: "remove a secret from the given store",
					Description: `
The 'remove' command requires passing in your configuration file
containing the secret store definitions you want to access. To get a
list of available secret store plugins, please have a look at
https://github.com/influxdata/telegraf/tree/master/plugins/secretstores.
and use the 'secrets list' command to get the IDs of available stores and keys.

For help on how to define secret stores, check the documentation of the
different plugins.

Assuming you use the default configuration file location, you can run
the following command to remove a secret from an available secret store

> telegraf secrets remove mystore mysecretkey

This will remove the secret with the key 'mysecretkey' from the secret
store with the ID 'mystore'. Removing a key that does not exist in that
store results in an error.
`,
					ArgsUsage: "<secret store ID> <secret key>",
					Action:    actions.Remove,
				},
			},
		},
	}
}
