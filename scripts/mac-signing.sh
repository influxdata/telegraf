#!/bin/bash

function cleanup () {
  echo "Cleaning up any existing Telegraf or Telegraf.app"
  printf "\n"
  rm -rf Telegraf
  rm -rf Telegraf.app
}

function archive_notarize()
{
  target="${1}"

  # submit archive for notarization, extract uuid
  uuid="$(
    # This extracts the value from `notarytool's` output. Unfortunately,
    # the 'id' is written to multiple times in the output. This requires
    # `awk` to `exit` after the first instance. However, doing so closes
    # `stdout` for `notarytool` which results with error code 141. This
    # takes the *complete* output from `notarytool` then
    # parses it with `awk`.
    awk '{ if ( $1 == "id:" ) { $1 = ""; print $0; exit 0; } }' \
      <<< "$(
        # shellcheck disable=SC2154
        xcrun notarytool submit \
          --apple-id "${AppleUsername}" \
          --password "${ApplePassword}" \
          --team-id 'M7DN9H35QT' \
          "${target}"
      )"
  )"
  shopt -s extglob
  uuid="${uuid%%+([[:space:]])}"  # strips leading whitespace
  uuid="${uuid##+([[:space:]])}"  # strips trailing whitespace

  if [[ -z "${uuid}" ]]; then
    exit 1
  fi

  # loop until notarization is complete
  while true ; do
    sleep 10

    response="$(
      # This extracts the value from `notarytool's` output. Unfortunately,
      # the 'id' is written to multiple times in the output. This requires
      # `awk` to `exit` after the first instance. However, doing so closes
      # `stdout` for `notarytool` which results with error code 141. This
      # takes the *complete* output from `notarytool` then
      # parses it with `awk`.
      awk '{ if ( $1 == "status:" ) { $1 = ""; print $0; exit 0; } }' \
        <<< "$(
          # shellcheck disable=SC2154
          xcrun notarytool info \
            --apple-id "${AppleUsername}" \
            --password "${ApplePassword}" \
            --team-id 'M7DN9H35QT' \
            "${uuid}"
        )"
    )"
    shopt -s extglob
    response="${response%%+([[:space:]])}"  # strips leading whitespace
    response="${response##+([[:space:]])}"  # strips trailing whitespace

    if [[ "${response}" != 'In Progress' ]] ; then
      break
    fi
  done

  if [[ "${response}" != 'Accepted' ]]; then
    exit 1
  fi
}

# Acquire the necessary certificates.
# MacCertificate, MacCertificatePassword, AppleSigningAuthorityCertificate are environment variables, to follow convention they should have been all caps.
# shellcheck disable=SC2154
base64 -D -o MacCertificate.p12 <<< "$MacCertificate"
# shellcheck disable=SC2154
sudo security import MacCertificate.p12 -k /Library/Keychains/System.keychain -P "$MacCertificatePassword" -A
# shellcheck disable=SC2154
base64 -D -o AppleSigningAuthorityCertificate.cer <<< "$AppleSigningAuthorityCertificate"
sudo security import AppleSigningAuthorityCertificate.cer -k '/Library/Keychains/System.keychain' -A

amdFile=$(find "$HOME/project/dist" -name "*darwin_amd64.tar*")
armFile=$(find "$HOME/project/dist" -name "*darwin_arm64.tar*")
macFiles=("${amdFile}" "${armFile}")

version=$(make version)
plutil -insert CFBundleShortVersionString -string "$version" ~/project/info.plist
plutil -insert CFBundleVersion -string "$version" ~/project/info.plist

for tarFile in "${macFiles[@]}";
do
  cleanup

  # Create the .app bundle directory structure
  RootAppDir="Telegraf.app/Contents"
  mkdir -p "$RootAppDir"
  mkdir -p "$RootAppDir/MacOS"
  mkdir -p "$RootAppDir/Resources"

  DeveloperID="Developer ID Application: InfluxData Inc. (M7DN9H35QT)"

  # Sign telegraf binary and the telegraf_entry_mac script
  echo "Extract $tarFile to $RootAppDir/Resources"
  tar -xzvf "$tarFile" --strip-components=2 -C "$RootAppDir/Resources"
  printf "\n"
  TelegrafBinPath="$RootAppDir/Resources/usr/bin/telegraf"
  codesign --force -s "$DeveloperID" --timestamp --options=runtime "$TelegrafBinPath"
  echo "Verify if $TelegrafBinPath was signed"
  codesign -dvv "$TelegrafBinPath"

  printf "\n"

  cp ~/project/scripts/telegraf_entry_mac "$RootAppDir"/MacOS
  EntryMacPath="$RootAppDir/MacOS/telegraf_entry_mac"
  codesign -s "$DeveloperID" --timestamp --options=runtime "$EntryMacPath"
  echo "Verify if $EntryMacPath was signed"
  codesign -dvv "$EntryMacPath"

  printf "\n"

  cp ~/project/info.plist "$RootAppDir"
  cp  ~/project/assets/windows/icon.icns "$RootAppDir/Resources"

  chmod +x "$RootAppDir/MacOS/telegraf_entry_mac"

  # Sign the entire .app bundle, and wrap it in a DMG.
  codesign -s "$DeveloperID" --timestamp --options=runtime --deep --force Telegraf.app
  baseName=$(basename "$tarFile" .tar.gz)
  echo "$baseName"
  hdiutil create -size 500m -volname Telegraf -srcfolder Telegraf.app "$baseName".dmg
  codesign -s "$DeveloperID" --timestamp --options=runtime "$baseName".dmg

  archive_notarize "${baseName}.dmg"

  # Attach the notarization to the DMG.
  xcrun stapler staple "$baseName".dmg
  cleanup

  mkdir -p ~/project/build/dist
  mv "$baseName".dmg ~/project/build/dist

  echo "$baseName.dmg signed and notarized!"
done
