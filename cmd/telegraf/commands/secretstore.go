package commands

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/awnumar/memguard"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"github.com/influxdata/telegraf"
)

// SecretStores provides access to the stores in the active configuration.
type SecretStores interface {
	InitSecretStores(*cli.Context)
	ListSecretStores() ([]string, error)
	GetSecretStore(string) (telegraf.SecretStore, error)
}

func secretListFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "reveal-secret",
			Usage: "also print the secret value",
		},
	}
}

type secretActions struct{ stores SecretStores }

func secretstoreCommand(actions commandActions) *cli.Command {
	return &cli.Command{
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
				Flags:     secretListFlags(),
				Action:    actions.secretsList,
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
				Action:    actions.secretsGet,
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
				Action:    actions.secretsSet,
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
				Action:    actions.secretsRemove,
			},
		},
	}
}

func (a secretActions) secretsListAction(cCtx *cli.Context) error {
	a.stores.InitSecretStores(cCtx)

	args := cCtx.Args()
	var storeIDs []string
	if args.Present() {
		storeIDs = args.Slice()
	} else {
		ids, err := a.stores.ListSecretStores()
		if err != nil {
			return fmt.Errorf("unable to determine secret store IDs: %w", err)
		}
		storeIDs = ids
	}
	sort.Strings(storeIDs)

	reveal := cCtx.Bool("reveal-secret")
	for _, storeID := range storeIDs {
		store, err := a.stores.GetSecretStore(storeID)
		if err != nil {
			return fmt.Errorf("unable to get secret store %q: %w", storeID, err)
		}
		keys, err := store.List()
		if err != nil {
			return fmt.Errorf("unable to get secrets from store %q: %w", storeID, err)
		}
		sort.Strings(keys)

		fmt.Printf("Known secrets for store %q:\n", storeID)
		for _, k := range keys {
			var v []byte
			if reveal {
				if v, err = store.Get(k); err != nil {
					return fmt.Errorf("unable to get value of secret %q from store %q: %w", k, storeID, err)
				}
			}
			fmt.Printf("    %-30s  %s\n", k, string(v))
			memguard.WipeBytes(v)
		}
	}

	return nil
}

func (a secretActions) secretsGetAction(cCtx *cli.Context) error {
	a.stores.InitSecretStores(cCtx)

	args := cCtx.Args()
	if !args.Present() || args.Len() != 2 {
		return errors.New("invalid number of arguments")
	}

	storeID := args.First()
	key := args.Get(1)

	store, err := a.stores.GetSecretStore(storeID)
	if err != nil {
		return fmt.Errorf("unable to get secret store: %w", err)
	}
	value, err := store.Get(key)
	if err != nil {
		return fmt.Errorf("unable to get secret: %w", err)
	}
	fmt.Printf("%s:%s = %s\n", storeID, key, value)

	return nil
}

func (a secretActions) secretsSetAction(cCtx *cli.Context) error {
	a.stores.InitSecretStores(cCtx)

	args := cCtx.Args()
	if !args.Present() || args.Len() < 2 {
		return errors.New("invalid number of arguments")
	}

	storeID := args.First()
	key := args.Get(1)
	value := args.Get(2)

	store, err := a.stores.GetSecretStore(storeID)
	if err != nil {
		return fmt.Errorf("unable to get secret store: %w", err)
	}
	editor, ok := store.(telegraf.SecretStoreEditor)
	if !ok {
		return fmt.Errorf("secret store %q does not support setting secrets", storeID)
	}

	if value == "" {
		fmt.Printf("Enter secret value: ")
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		value = string(b)
	}

	if err := editor.Set(key, value); err != nil {
		return fmt.Errorf("unable to set secret: %w", err)
	}

	return nil
}

func (a secretActions) secretsRemoveAction(cCtx *cli.Context) error {
	a.stores.InitSecretStores(cCtx)

	args := cCtx.Args()
	if !args.Present() || args.Len() != 2 {
		return errors.New("invalid number of arguments")
	}

	storeID := args.First()
	key := args.Get(1)

	store, err := a.stores.GetSecretStore(storeID)
	if err != nil {
		return fmt.Errorf("unable to get secret store: %w", err)
	}
	editor, ok := store.(telegraf.SecretStoreEditor)
	if !ok {
		return fmt.Errorf("secret store %q does not support removing secrets", storeID)
	}

	if err := editor.Remove(key); err != nil {
		return fmt.Errorf("unable to remove secret: %w", err)
	}

	return nil
}
