# Running Telegraf as a Windows Service

If you have tried to install Go binaries as Windows Services with the **sc.exe**
tool you may have seen that the service errors and stops running after a while.

**NSSM** (the Non-Sucking Service Manager) is a tool that helps you in a 
[number of scenarios](http://nssm.cc/scenarios) including running Go binaries
that were not specifically designed to run only in Windows platforms.

## NSSM Installation via Chocolatey

You can install [Chocolatey](https://chocolatey.org/) and [NSSM](http://nssm.cc/) 
with these commands

```powershell
iex ((new-object net.webclient).DownloadString('https://chocolatey.org/install.ps1'))
choco install -y nssm
```

## Installing Telegraf as a Windows Service with NSSM

You can download the latest Telegraf Windows binaries (still Experimental at 
the moment) from [the Telegraf Github repo](https://github.com/influxdata/telegraf).

Then you can create a C:\telegraf folder, unzip the binary there and modify the 
**telegraf.conf** sample to allocate the metrics you want to send to **InfluxDB**.

Once you have NSSM installed in your system, the process is quite straightforward.
You only need to type this command in your Windows shell

```powershell
nssm install Telegraf c:\telegraf\telegraf.exe -config c:\telegraf\telegraf.config
```

And now your service will be installed in Windows and you will be able to start and
stop it gracefully