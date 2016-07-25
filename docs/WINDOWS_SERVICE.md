# Running Telegraf as a Windows Service

Telegraf natively supports running as a Windows Service. Outlined below is are
the general steps to set it up.

1. Obtain the telegraf windows distribution
2. Create the directory C:\telegraf (if you install in a different location you
   will need to edit `cmd/telegraf/telegraf.go` and change the config file
   location and recompile to use your location)
3. Place the executable and the config file into C:\telegraf
4. Run `C:\telegraf\telegraf.exe --service install` as an administrator
5. Edit the configuration file to meet your needs
6. Run `C:\telegraf\telegraf.exe --config C:\telegraf\telegraf.conf --test` to
   check that it works
7. Run `net start telegraf` to start collecting data

## Other supported operations

Telegraf can manage its own service through the --service flag:

| Command                            | Effect                        |
|------------------------------------|-------------------------------|
| `telegraf.exe --service install`   | Install telegraf as a service |
| `telegraf.exe --service uninstall` | Remove the telegraf service   |
| `telegraf.exe --service start`     | Start the telegraf service    |
| `telegraf.exe --service stop`      | Stop the telegraf service     |

