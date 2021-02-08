$tempCertFile = New-TemporaryFile

# Replace with method of retrieving cert and password.
$certText = $env:fakeWindowsCert
$CertPass = $env:fakeWindowsCertPass
$finalFileName = $tempCertFile.FullName
$certBytes = [Convert]::FromBase64String($certText)
[System.IO.File]::WriteAllBytes($finalFileName, $certBytes)
$CertPath = $finalFileName
$Cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($CertPath, $CertPass)

# Replace with circleCI artifact path.
$artifactDirectory = "./build/dist"
$extractDirectory = $artifactDirectory + "\" + "extracted"
foreach ($file in get-ChildItem $artifactDirectory -recurse | where {$_.name -like "*windows*"} | select name) 
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
