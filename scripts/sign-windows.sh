#!/bin/bash
set -eux

# Install dependencies
sudo apt update && sudo apt install --yes 7zip default-jre-headless osslsigncode wget
wget https://github.com/ebourg/jsign/releases/download/5.0/jsign_5.0_all.deb
sha256sum="9877a0949a9c9ac4485155bbb8679ac863d3ec3d67e0a380b880eed650d06854"
if ! echo "${sha256sum}  jsign_5.0_all.deb" | sha256sum --check -; then
    echo "Checksum for jsign deb failed" >&2
    exit 1
fi
sudo dpkg -i jsign_5.0_all.deb

# Load certificates
touch "$SM_CLIENT_CERT_FILE"
set +x
echo "$SM_CLIENT_CERT_FILE_B64" > "$SM_CLIENT_CERT_FILE.b64"
set -x
base64 -d "$SM_CLIENT_CERT_FILE.b64" > "$SM_CLIENT_CERT_FILE"

# Loop through and sign + verify the binaries
artifactDirectory="./build/dist"
extractDirectory="$artifactDirectory/extracted"
for file in "$artifactDirectory"/*windows*; do
    7zz x "$file" -o$extractDirectory
    subDirectoryPath=$(find $extractDirectory -mindepth 1 -maxdepth 1 -type d)
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

    7zz a -r "$file" "$subDirectoryPath"
    rm -rf "$extractDirectory"
done
