package commands

import "github.com/urfave/cli/v2"

type PluginActions struct {
	Plugins      cli.ActionFunc
	Inputs       cli.ActionFunc
	Outputs      cli.ActionFunc
	Processors   cli.ActionFunc
	Aggregators  cli.ActionFunc
	Secretstores cli.ActionFunc
	Parsers      cli.ActionFunc
	Serializers  cli.ActionFunc
}

func pluginsCommands(actions PluginActions) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "plugins",
			Usage: "commands for printing available plugins",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "deprecated",
					Usage: "print only deprecated plugins",
				},
			},
			Action: actions.Plugins,
			Subcommands: []*cli.Command{
				{
					Name:  "inputs",
					Usage: "Print available input plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Inputs,
				},
				{
					Name:  "outputs",
					Usage: "Print available output plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Outputs,
				},
				{
					Name:  "processors",
					Usage: "Print available processor plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Processors,
				},
				{
					Name:  "aggregators",
					Usage: "Print available aggregator plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Aggregators,
				},
				{
					Name:  "secretstores",
					Usage: "Print available secretstore plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Secretstores,
				},
				{
					Name:  "parsers",
					Usage: "Print available parser plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Parsers,
				},
				{
					Name:  "serializers",
					Usage: "Print available serializer plugins",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "deprecated",
							Usage: "print only deprecated plugins",
						},
					},
					Action: actions.Serializers,
				},
			},
		},
	}
}
