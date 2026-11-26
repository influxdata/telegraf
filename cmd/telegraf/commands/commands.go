package commands

import "github.com/urfave/cli/v2"

// AvailableCommands describes the CLI across all supported platforms.
func AvailableCommands() *cli.Command {
	return &cli.Command{
		Name:        "telegraf",
		Flags:       RootFlags("all"),
		Subcommands: getSubcommands("all", commandActions{}),
	}
}

// RootFlags returns the root flags supported on the specified platform.
func RootFlags(goos string) []cli.Flag {
	flags := append(getRunFlags(), getGlobalFlags()...)
	switch goos {
	case "all", "windows":
		flags = append(flags, windowsServiceFlags()...)
	}
	return flags
}

// Subcommands returns the subcommands supported on the specified platform.
func Subcommands(goos string, stores SecretStores) []*cli.Command {
	secrets := secretActions{stores: stores}
	return getSubcommands(goos, commandActions{
		version:             versionAction,
		configCreate:        ConfigCreateAction,
		configCheck:         configCheckAction,
		configMigrate:       configMigrateAction,
		secretsList:         secrets.secretsListAction,
		secretsGet:          secrets.secretsGetAction,
		secretsSet:          secrets.secretsSetAction,
		secretsRemove:       secrets.secretsRemoveAction,
		plugins:             pluginsAction,
		pluginsInputs:       pluginsInputsAction,
		pluginsOutputs:      pluginsOutputsAction,
		pluginsProcessors:   pluginsProcessorsAction,
		pluginsAggregators:  pluginsAggregatorsAction,
		pluginsSecretstores: pluginsSecretstoresAction,
		pluginsParsers:      pluginsParsersAction,
		pluginsSerializers:  pluginsSerializersAction,
		service:             getServiceActions(),
	})
}
