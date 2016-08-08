# Running Telegraf as a Windows Service

Telegraf natively supports running as a Windows Service. Outlined below is are
the general steps to set it up.

1. Obtain the telegraf windows distribution
2. Create the directory `C:\Program Files\Telegraf` (if you install in a different
   location simply specify the `-config` parameter with the desired location)
3. Place the executable and the config file into `C:\Program Files\Telegraf`
4. To install the service into the Windows Service Manager, run (as an
   administrator):
   ```ps
   C:\Program Files\Telegraf\telegraf.exe --service install
   ```
5. Edit the configuration file to meet your needs
6. To check that it works, run:
   ```ps
   C:\Program Files\Telegraf\telegraf.exe --config C:\Program Files\Telegraf\telegraf.conf --test
   ```
7. To start collecting data, run:
   ```ps
   net start telegraf
   ```

## Other supported operations

Telegraf can manage its own service through the --service flag:

| Command                            | Effect                        |
|------------------------------------|-------------------------------|
| `telegraf.exe --service install`   | Install telegraf as a service |
| `telegraf.exe --service uninstall` | Remove the telegraf service   |
| `telegraf.exe --service start`     | Start the telegraf service    |
| `telegraf.exe --service stop`      | Stop the telegraf service     |

