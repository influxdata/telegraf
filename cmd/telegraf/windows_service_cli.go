package main

import "github.com/urfave/cli/v2"

type windowsServiceActions struct {
	install   cli.ActionFunc
	uninstall cli.ActionFunc
	start     cli.ActionFunc
	stop      cli.ActionFunc
	status    cli.ActionFunc
}

func windowsServiceFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "service",
			Usage: "operate on the service (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-name",
			Value: "telegraf",
			Usage: "service name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-display-name",
			Value: "Telegraf Data Collector Service",
			Usage: "service display name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-restart-delay",
			Value: "5m",
			Usage: "duration for delaying the Windows service restart on failure",
		},
		&cli.BoolFlag{
			Name:  "service-auto-restart",
			Usage: "auto restart service on failure (windows only)",
		},
		&cli.BoolFlag{
			Name:  "console",
			Usage: "run as console application (windows only)",
		},
	}
}

func windowsServiceCommands(actions windowsServiceActions) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "service",
			Usage: "commands for operate on the Windows service",
			Flags: nil,
			Subcommands: []*cli.Command{
				{
					Name:  "install",
					Usage: "install Telegraf as a Windows service",
					Description: `
The 'install' command with create a Windows service for automatically starting
Telegraf with the specified configuration and service parameters. If no
configuration(s) is specified the service will use the file in
"C:\Program Files\Telegraf\telegraf.conf".

To install Telegraf as a service use

> telegraf service install

In case you are planning to start multiple Telegraf instances as a service,
you must use distinctive service-names for each instance. To install two
services with different configurations use

> telegraf --config "C:\Program Files\Telegraf\telegraf-machine.conf" --service-name telegraf-machine service install
> telegraf --config "C:\Program Files\Telegraf\telegraf-service.conf" --service-name telegraf-service service install
`,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "display-name",
							Value: "Telegraf Data Collector Service",
							Usage: "service name as displayed in the service manager",
						},
						&cli.StringFlag{
							Name:  "restart-delay",
							Value: "5m",
							Usage: "duration for delaying the service restart on failure",
						},
						&cli.BoolFlag{
							Name:  "auto-restart",
							Usage: "enable automatic service restart on failure",
						},
					},
					Action: actions.install,
				},
				{
					Name:  "uninstall",
					Usage: "remove the Telegraf Windows service",
					Description: `
The 'uninstall' command removes the Telegraf service with the given name. To
remove a service use

> telegraf service uninstall

In case you specified a custom service-name during install use

> telegraf --service-name telegraf-machine service uninstall
`,
					Action: actions.uninstall,
				},
				{
					Name:  "start",
					Usage: "start the Telegraf Windows service",
					Description: `
The 'start' command triggers the start of the Windows service with the given
name. To start the service either use the Windows service manager or run

> telegraf service start

In case you specified a custom service-name during install use

> telegraf --service-name telegraf-machine service start
`,
					Action: actions.start,
				},
				{
					Name:  "stop",
					Usage: "stop the Telegraf Windows service",
					Description: `
The 'stop' command triggers the stop of the Windows service with the given
name and will wait until the service is actually stopped. To stop the service
either use the Windows service manager or run

> telegraf service stop

In case you specified a custom service-name during install use

> telegraf --service-name telegraf-machine service stop
`,
					Action: actions.stop,
				},
				{
					Name:  "status",
					Usage: "query the Telegraf Windows service status",
					Description: `
The 'status' command queries the current state of the Windows service with the
given name. To query the service either check the Windows service manager or run

> telegraf service status

In case you specified a custom service-name during install use

> telegraf --service-name telegraf-machine service status
`,
					Action: actions.status,
				},
			},
		},
	}
}
