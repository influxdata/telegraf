Set-PSDebug -Trace 1
New-Item -ItemType Directory -Force -Path "${PSScriptRoot}\output"
docker run --rm -ti -v "${PSScriptRoot}:C:\src" -v "${PSSCriptRoot}\output:C:\output" golang:1.9.2-windowsservercore-ltsc2016 powershell C:\src\scripts\build_sfx.ps1