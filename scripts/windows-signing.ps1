# Fail on first error.
$ErrorActionPreference = "Stop"

# Update the version of Compress-Archive to support zipping correctly.
Install-Module Microsoft.PowerShell.Archive -MinimumVersion 1.2.3.0 -Repository PSGallery -Force
Import-Module Microsoft.PowerShell.Archive

# Save the certfile locally.
if (Test-Path C:\CERT_FILE.p12.b64) {
    Remove-Item -Force -Path C:\CERT_FILE.p12.b64
  }

if (Test-Path C:\Certificate_pkcs12.p12) {
    Remove-Item -Force -Path C:\Certificate_pkcs12.p12
}

Write-Output "Saving certificate locally"
New-Item C:\CERT_FILE.p12.b64
Set-Content -Path C:\CERT_FILE.p12.b64 -Value $env:SM_CLIENT_CERT_FILE_B64
certutil -decode C:\CERT_FILE.p12.b64 C:\Certificate_pkcs12.p12

# Download and install signing tools.
if (!(Test-Path "C:\Program Files\DigiCert\DigiCert One Signing Manager Tools\smctl.exe")) {
    Write-Output "Installing smctl"
    curl.exe -X GET https://one.digicert.com/signingmanager/api-ui/v1/releases/smtools-windows-x64.msi/download -H "x-api-key:$env:SM_API_KEY" -o smtools-windows-x64.msi
    msiexec.exe /i smtools-windows-x64.msi /quiet /qn
}

certutil.exe -csp "DigiCert Software Trust Manager KSP" -key -user
& "C:\Program Files\DigiCert\DigiCert One Signing Manager Tools\smctl.exe" windows certsync

# Go through the artifacts directory and sign the 'windows' artifacts.
$artifactDirectory = "./build/dist"
$extractDirectory = $artifactDirectory + "/" + "extracted"
foreach ($file in get-ChildItem $artifactDirectory | where {$_.name -like "*windows*"} | select name)
{
    $artifact = $artifactDirectory + "/" + $file.Name
    Expand-Archive -LiteralPath $artifact -DestinationPath $extractDirectory -Force

    $subDirectoryPath = $extractDirectory + "/" + (Get-ChildItem -Path $extractDirectory | Select-Object -First 1).Name
    $telegrafExePath = $subDirectoryPath + "/" + "telegraf.exe"

    & "C:\Program Files\DigiCert\DigiCert One Signing Manager Tools\smctl.exe" sign --input "$telegrafExePath" --fingerprint $env:SM_FINGERPRINT --verbose
    & "C:\Program Files\DigiCert\DigiCert One Signing Manager Tools\smctl.exe" sign verify --input "$telegrafExePath"

    Compress-Archive -Path $subDirectoryPath -DestinationPath $artifact -Force
    Remove-Item $extractDirectory -Force -Recurse
}

Remove-Item $finalFileName -Force
