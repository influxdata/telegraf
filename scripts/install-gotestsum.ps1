# install-gotestsum.ps1 is used in the CI pipeline to download the tool gotestsum for Windows
# Link to gotestsum: https://github.com/gotestyourself/gotestsum
# Execution example: ./scripts/install-gotestsum.ps1 1.7.0 C:\Users\circleci\

$version = $args[0]

function Get-Gotestsum {
    Write-Output "Downloading gotestsum"
    Invoke-WebRequest -Uri "https://github.com/gotestyourself/gotestsum/releases/download/v${version}/gotestsum_${version}_windows_amd64.tar.gz" -OutFile gotestsum.tar.gz 
    tar -v -C $path --extract --file=gotestsum.tar.gz gotestsum.exe   
}

if (-not(Test-Path -Path gotestsum.exe)) {
    Get-Gotestsum
}
else {
    $version_output = [string] (& gotestsum.exe--version)
    $expected_version_output = "gotestsum version $version"
    Write-Output $version_output
    if (-not($version_output -eq $expected_version_output) ) {
        Write-Output "Removing old version, and getting new version $version"
        Remove-Item gotestsum.exe
        Get-Gotestsum
    }
}
