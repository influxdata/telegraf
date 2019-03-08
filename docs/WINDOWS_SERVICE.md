# Running Telegraf as a Windows Service

Telegraf natively supports running as a Windows Service. Outlined below is are
the general steps to set it up.

1. Obtain the telegraf windows distribution
2. Create the directory `C:\Program Files\Telegraf` (if you install in a different
   location simply specify the `--config` parameter with the desired location)
3. Place the telegraf.exe and the telegraf.conf config file into `C:\Program Files\Telegraf`
4. To install the service into the Windows Service Manager, run the following in PowerShell as an administrator (If necessary, you can wrap any spaces in the file paths in double quotes ""):

   ```
   > C:\"Program Files"\Telegraf\telegraf.exe --service install
   ```

5. Edit the configuration file to meet your needs
6. To check that it works, run:

   ```
   > C:\"Program Files"\Telegraf\telegraf.exe --config C:\"Program Files"\Telegraf\telegraf.conf --test
   ```

7. To start collecting data, run:

   ```
   > net start telegraf
   ```

## Config Directory

You can also specify a `--config-directory` for the service to use:
1. Create a directory for config snippets: `C:\Program Files\Telegraf\telegraf.d`
2. Include the `--config-directory` option when registering the service:
   ```
   > C:\"Program Files"\Telegraf\telegraf.exe --service install --config C:\"Program Files"\Telegraf\telegraf.conf --config-directory C:\"Program Files"\Telegraf\telegraf.d
   ```

## Other supported operations

Telegraf can manage its own service through the --service flag:

| Command                            | Effect                        |
|------------------------------------|-------------------------------|
| `telegraf.exe --service install`   | Install telegraf as a service |
| `telegraf.exe --service uninstall` | Remove the telegraf service   |
| `telegraf.exe --service start`     | Start the telegraf service    |
| `telegraf.exe --service stop`      | Stop the telegraf service     |

## Install multiple services

You can install multiple telegraf instances with --service-name flag:

```
   > C:\"Program Files"\Telegraf\telegraf.exe --service install --service-name telegraf-1
   > C:\"Program Files"\Telegraf\telegraf.exe --service install --service-name telegraf-2
   > C:\"Program Files"\Telegraf\telegraf.exe --service uninstall --service-name telegraf-1
```

Troubleshooting  common error #1067

When installing as service in Windows, always double check to specify full path of the config file, otherwise windows service will fail to start

 --config C:\"Program Files"\Telegraf\telegraf.conf
