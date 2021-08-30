# install-gotestsum.ps1 is used in the CI pipeline to download the tool gotestsum for Windows
# Link to gotestsum: https://github.com/gotestyourself/gotestsum
# Execution example: ./scripts/install-gotestsum.ps1 1.7.0 C:\Users\circleci\

$version=$args[0]
$path=$args[1]
$exePath = Join-Path -Path $path -ChildPath gotestsum.exe
$tarPath = Join-Path -Path $path -ChildPath gotestsum.tar.gz

function Download-Gotestsum {
    Write-Output "Downloading gotestsum"
    Invoke-WebRequest -Uri "https://github.com/gotestyourself/gotestsum/releases/download/v${version}/gotestsum_${version}_windows_amd64.tar.gz" -OutFile $tarPath 
    tar -v -C $path --extract --file=$tarPath gotestsum.exe   
}

if(-not(Test-Path -Path $exePath)){
   Download-Gotestsum
} else {
    $version_output = [string] (& $exePath --version)
    $expected_version_output = "gotestsum version 1.7.0"
    Write-Output $version_output
    if(-not($version_output -eq $expected_version_output) ){
        Write-Output "Removing old version, and getting new version $version"
        Remove-Item $exePath
        Download-Gotestsum
    }
}
