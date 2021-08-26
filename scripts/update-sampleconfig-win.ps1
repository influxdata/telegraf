# update-sampleconfig-win.ps1 is responsible to generate a new telegraf_windows.conf
# telegraf_windows.conf: https://github.com/influxdata/telegraf/blob/master/etc/telegraf_windows.conf

# These artifact directories are defined in .circleci/config.yml
$artifactDirectory = "./build/dist"
$extractDirectory = $artifactDirectory + "\" + "extracted"

# Get telegraf.exe from windows artifact zip file
$windowsZip = $artifactDirectory + "\" + (Get-ChildItem "./Sandbox/" | where {$_.name -like "*windows*"} | Select-Object -first 1).Name
Expand-Archive -LiteralPath $windowsZip -DestinationPath $extractDirectory -Force
$subDirectoryPath = $extractDirectory + "\" + (Get-ChildItem -Path $extractDirectory | Select-Object -First 1).Name
$telegrafExePath = $subDirectoryPath + "\" + "telegraf.exe"

# Generate a new windows config and compare with current, if different create pull request
& $telegrafExePath config > telegraf_windows_new.conf
$pathToTelegrafWindowsConf = "\etc\telegraf_windows.conf"

if((Get-FileHash "telegraf_windows_new.conf").Hash -ne (Get-FileHash $pathToTelegrafWindowsConf).Hash){
    Write-Output "Difference found, creating pull request"
    choco install gh
    # gh auth login --with-token xxxx
    # gh pr create
}
