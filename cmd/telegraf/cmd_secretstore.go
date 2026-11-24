// Command handling for secret stores' "secrets" command
package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/cmd/telegraf/commands"
)

func processFilterOnlySecretStoreFlags(ctx *cli.Context) Filters {
	sectionFilters := []string{"inputs", "outputs", "processors", "aggregators"}
	inputFilters := []string{"-"}
	outputFilters := []string{"-"}
	processorFilters := []string{"-"}
	aggregatorFilters := []string{"-"}

	// Only load the secret stores
	var secretstore string
	if len(ctx.Lineage()) >= 2 {
		parent := ctx.Lineage()[1] // ancestor contexts in order from child to parent
		secretstore = parent.String("secretstore-filter")
	}

	// If both the parent and command filters are defined, append them together
	secretstore = appendFilter(secretstore, ctx.String("secretstore-filter"))
	secretstoreFilters := deleteEmpty(strings.Split(secretstore, ":"))
	return Filters{sectionFilters, inputFilters, outputFilters, aggregatorFilters, processorFilters, secretstoreFilters}
}

func getSecretStoreActions(m App) commands.SecretStoreActions {
	return commands.SecretStoreActions{
		List: func(cCtx *cli.Context) error {
			// Only load the secret stores
			filters := processFilterOnlySecretStoreFlags(cCtx)
			g := GlobalFlags{
				config:     cCtx.StringSlice("config"),
				configDir:  cCtx.StringSlice("config-directory"),
				plugindDir: cCtx.String("plugin-directory"),
				password:   cCtx.String("password"),
				debug:      cCtx.Bool("debug"),
			}
			w := WindowFlags{}
			m.Init(nil, filters, g, w)

			args := cCtx.Args()
			var storeIDs []string
			if args.Present() {
				storeIDs = args.Slice()
			} else {
				ids, err := m.ListSecretStores()
				if err != nil {
					return fmt.Errorf("unable to determine secret store IDs: %w", err)
				}
				storeIDs = ids
			}
			sort.Strings(storeIDs)

			reveal := cCtx.Bool("reveal-secret")
			for _, storeID := range storeIDs {
				store, err := m.GetSecretStore(storeID)
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
		},
		Get: func(cCtx *cli.Context) error {
			// Only load the secret stores
			filters := processFilterOnlySecretStoreFlags(cCtx)
			g := GlobalFlags{
				config:     cCtx.StringSlice("config"),
				configDir:  cCtx.StringSlice("config-directory"),
				plugindDir: cCtx.String("plugin-directory"),
				password:   cCtx.String("password"),
				debug:      cCtx.Bool("debug"),
			}
			w := WindowFlags{}
			m.Init(nil, filters, g, w)

			args := cCtx.Args()
			if !args.Present() || args.Len() != 2 {
				return errors.New("invalid number of arguments")
			}

			storeID := args.First()
			key := args.Get(1)

			store, err := m.GetSecretStore(storeID)
			if err != nil {
				return fmt.Errorf("unable to get secret store: %w", err)
			}
			value, err := store.Get(key)
			if err != nil {
				return fmt.Errorf("unable to get secret: %w", err)
			}
			fmt.Printf("%s:%s = %s\n", storeID, key, value)

			return nil
		},
		Set: func(cCtx *cli.Context) error {
			// Only load the secret stores
			filters := processFilterOnlySecretStoreFlags(cCtx)
			g := GlobalFlags{
				config:     cCtx.StringSlice("config"),
				configDir:  cCtx.StringSlice("config-directory"),
				plugindDir: cCtx.String("plugin-directory"),
				password:   cCtx.String("password"),
				debug:      cCtx.Bool("debug"),
			}
			w := WindowFlags{}
			m.Init(nil, filters, g, w)

			args := cCtx.Args()
			if !args.Present() || args.Len() < 2 {
				return errors.New("invalid number of arguments")
			}

			storeID := args.First()
			key := args.Get(1)
			value := args.Get(2)

			store, err := m.GetSecretStore(storeID)
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
		},
		Remove: func(cCtx *cli.Context) error {
			// Only load the secret stores
			filters := processFilterOnlySecretStoreFlags(cCtx)
			g := GlobalFlags{
				config:     cCtx.StringSlice("config"),
				configDir:  cCtx.StringSlice("config-directory"),
				plugindDir: cCtx.String("plugin-directory"),
				password:   cCtx.String("password"),
				debug:      cCtx.Bool("debug"),
			}
			w := WindowFlags{}
			m.Init(nil, filters, g, w)

			args := cCtx.Args()
			if !args.Present() || args.Len() != 2 {
				return errors.New("invalid number of arguments")
			}

			storeID := args.First()
			key := args.Get(1)

			store, err := m.GetSecretStore(storeID)
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
		},
	}
}
