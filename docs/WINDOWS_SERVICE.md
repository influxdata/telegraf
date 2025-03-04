# Running Telegraf as a Windows Service

Telegraf natively supports running as a Windows Service. Outlined below is are
the general steps to set it up.

1. Obtain the telegraf windows distribution
2. Create the directory `C:\Program Files\Telegraf` or use a custom directory
   if desired
3. Place the telegraf.exe and the telegraf.conf config file into the directory,
   either `C:\Program Files\Telegraf` or the custom directory of your choice.
   If you install in a different location simply specify the `--config`
   parameter with the desired location.
4. To install the service into the Windows Service Manager, run the command
   as administrator. Make sure to wrap parameters containing spaces in double
   quotes:

   ```shell
   > "C:Program Files\Telegraf\telegraf.exe" service install
   ```

5. Edit the configuration file to meet your needs
6. To check that it works, run:

   ```shell
   > "C:\Program Files\Telegraf\telegraf.exe" --config "C:\Program Files\Telegraf\telegraf.conf" --test
   ```

7. To start collecting data, run:

   ```shell
   > net start telegraf
   ```

   or

   ```shell
   > "C:\Program Files\Telegraf\telegraf.exe" service start
   ```

   or use the Windows service manager to start the service

Please also check the Windows event log or your configured log-file for errors
during startup.

## Config Directory

You can also specify a `--config-directory` for the service to use:

1. Create a directory for config snippets: `C:\Program Files\Telegraf\telegraf.d`
2. Include the `--config-directory` option when registering the service:

   ```shell
   > "C:\Program Files\Telegraf\telegraf.exe" --config C:\"Program Files"\Telegraf\telegraf.conf --config-directory C:\"Program Files"\Telegraf\telegraf.d service install
   ```

## Other supported operations

Telegraf can manage its own service through the --service flag:

| Command                          | Effect                                   |
|----------------------------------|------------------------------------------|
| `telegraf.exe service install`   | Install telegraf as a service            |
| `telegraf.exe service uninstall` | Remove the telegraf service              |
| `telegraf.exe service start`     | Start the telegraf service               |
| `telegraf.exe service stop`      | Stop the telegraf service                |
| `telegraf.exe service status`    | Query the status of the telegraf service |

## Install multiple services

Running multiple instances of Telegraf is seldom needed, as you can run
multiple instances of each plugin and route metric flow using the metric
filtering options. However, if you do need to run multiple telegraf instances
on a single system, you can install the service with the `--service-name` and
`--display-name` flags to give the services unique names:

```shell
> "C:\Program Files\Telegraf\telegraf.exe" --service-name telegraf-1 service install --display-name "Telegraf 1"
> "C:\Program Files\Telegraf\telegraf.exe" --service-name telegraf-2 service install --display-name "Telegraf 2"
```

## Auto restart and restart delay

By default the service will not automatically restart on failure. Providing the
`--auto-restart` flag during installation will always restart the service with
a default delay of 5 minutes. To modify this to for example 3 minutes,
additionally provide `--restart-delay 3m` flag. The delay can be any valid
`time.Duration` string.

## Troubleshooting

When Telegraf runs as a Windows service, Telegraf logs all messages concerning
the service startup to the Windows event log. All messages and errors occuring
during runtime will be logged to the log-target you configured.
Check the event log for errors reported by the `telegraf` service (or the
service-name you configured) during service startup:
`Event Viewer -> Windows Logs -> Application`

### Common error #1067

When installing as service in Windows, always double check to specify full path
of the config file, otherwise windows service will fail to start. Use

```shell
> "C:\Program Files\Telegraf\telegraf.exe" --config "C:\MyConfigs\telegraf.conf" service install
```

instead of

```shell
> "C:\Program Files\Telegraf\telegraf.exe" --config "telegraf.conf" service install
```

### Service is killed during shutdown

When shuting down Windows the Telegraf service tries to cleanly stop when
receiving the corresponding notification from the Windows service manager. The
exit process involves stopping all inputs, processors and aggregators and
finally to flush all remaining metrics to the output(s). In case many metrics
are not yet flushed this final step might take some time. However, Windows will
kill the service and the corresponding process after a predefined timeout
(usually 5 seconds).

You can change that timeout in the registry under

````text
HKLM\SYSTEM\CurrentControlSet\Control\WaitToKillServiceTimeout
```

**NOTE:** The value is in milliseconds and applies to **all** services!
