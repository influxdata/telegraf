#!/bin/bash
set -eux

# Install dependencies
sudo apt update && sudo apt install --yes wget default-jre-headless osslsigncode
wget https://github.com/ebourg/jsign/releases/download/5.0/jsign_5.0_all.deb
sha256sum="9877a0949a9c9ac4485155bbb8679ac863d3ec3d67e0a380b880eed650d06854"
if ! echo "${sha256sum}  jsign_5.0_all.deb" | sha256sum --check -; then
    echo "Checksum for jsign deb failed" >&2
    exit 1
fi
sudo dpkg -i jsign_5.0_all.deb

# Load certificates
touch Certificate_pkcs12.p12
echo "$SM_CLIENT_CERT_FILE_B64" > CERT_FILE.p12.b64
base64 -d CERT_FILE.p12.b64 > Certificate_pkcs12.p12

# Loop through and sign + verify the binaries
artifactDirectory="./build/dist"
extractDirectory="$artifactDirectory/extracted"
for file in $artifactDirectory/*windows*; do
    7z x "$file" -o$extractDirectory
    subDirectoryPath="$extractDirectory/$(ls $extractDirectory | head -n 1)"
    telegrafExePath="$subDirectoryPath/telegraf.exe"

    jsign \
        -storetype DIGICERTONE \
        -alias "$SM_CERT_ALIAS" \
        -storepass "$SM_API_KEY|$SM_CLIENT_CERT_FILE|$SM_CLIENT_CERT_PASSWORD" \
        -alg SHA-256 \
        -tsaurl http://timestamp.digicert.com \
        "$telegrafExePath"

    osslsigncode verify \
        -CAfile /usr/share/ca-certificates/mozilla/DigiCert_Trusted_Root_G4.crt \
        -TSA-CAfile /usr/share/ca-certificates/mozilla/DigiCert_Trusted_Root_G4.crt \
        -in "$telegrafExePath"

    7z a "$file" "$subDirectoryPath/*"
    rm -rf "$subDirectoryPath"
done
