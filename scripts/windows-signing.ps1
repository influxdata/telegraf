$tempCertFile = New-TemporaryFile

# Retrieve environment variables for cert/password.
$certText = $env:windowsCert
$CertPass = $env:windowsCertPassword

# Create a Cert object by converting the cert string to bytes.
$finalFileName = $tempCertFile.FullName
$certBytes = [Convert]::FromBase64String($certText)
[System.IO.File]::WriteAllBytes($finalFileName, $certBytes)
$CertPath = $finalFileName
$Cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($CertPath, $CertPass)

# Go through the artifacts directory and sign the 'windows' artifacts.
$artifactDirectory = "./build/dist"
$extractDirectory = $artifactDirectory + "\" + "extracted"
foreach ($file in get-ChildItem $artifactDirectory | where {$_.name -like "*windows*"} | select name) 
{
    $artifact = $artifactDirectory + "\" + $file.Name
    Expand-Archive -LiteralPath $artifact -DestinationPath $extractDirectory -Force

    $subDirectoryPath = $extractDirectory + "\" + (Get-ChildItem -Path $extractDirectory | Select-Object -First 1).Name
    $telegrafExePath = $subDirectoryPath + "\" + "telegraf.exe"
    Set-AuthenticodeSignature -Certificate $Cert -FilePath  $telegrafExePath -TimestampServer http://timestamp.digicert.com
    Compress-Archive -Path $subDirectoryPath -DestinationPath $artifact -Force
    Remove-Item $extractDirectory -Force -Recurse
}

Remove-Item $finalFileName -Force
