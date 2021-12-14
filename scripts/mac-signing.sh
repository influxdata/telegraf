# Acquire the necessary certificates.
base64 -D -o MacCertificate.p12 <<< $MacCertificate
sudo security import MacCertificate.p12 -k /Library/Keychains/System.keychain -P $MacCertificatePassword -A
base64 -D -o AppleSigningAuthorityCertificate.cer <<< $AppleSigningAuthorityCertificate
sudo security import AppleSigningAuthorityCertificate.cer -k '/Library/Keychains/System.keychain' -A

# Extract the built mac binary and sign it.
cd dist
tarFile=$(find . -name "*darwin_amd64.tar*")
tar -xzvf $tarFile
baseName=$(basename $tarFile .tar.gz)
cd $(find . -name "*telegraf-*" -type d)
cd usr/bin
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime telegraf
codesign -v telegraf

# Reset back out to the main directory.
cd
cd project/dist
extractedFolder=$(find . -name "*telegraf-*" -type d)

# Sign the 'telegraf entry' script, which is required to open Telegraf upon opening the .app bundle.
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime ../scripts/telegraf_entry_mac
codesign -v ../scripts/telegraf_entry_mac

# Create the .app bundle.
mkdir Telegraf
cd Telegraf
mkdir Contents
cd Contents
mkdir MacOS
mkdir Resources
cd ../..
cp ../info.plist Telegraf/Contents
cp -R "$extractedFolder"/ Telegraf/Contents/Resources
cp ../scripts/telegraf_entry_mac Telegraf/Contents/MacOS
cp ../assets/icon.icns Telegraf/Contents/Resources
chmod +x Telegraf/Contents/MacOS/telegraf_entry_mac
mv Telegraf Telegraf.app

# Sign the entire .app bundle, and wrap it in a DMG.
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime --deep --force Telegraf.app
hdiutil create -size 500m -volname Telegraf -srcfolder Telegraf.app "$baseName".dmg
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime "$baseName".dmg

# Send the DMG to be notarized.
uuid=$(xcrun altool --notarize-app --primary-bundle-id "com.influxdata.telegraf" --username "$AppleUsername" --password "$ApplePassword" --file "$baseName".dmg | awk '/RequestUUID/ { print $NF; }')
echo $uuid
if [[ $uuid == "" ]]; then 
  echo "Could not upload for notarization."
  exit 1
fi

# Wait until the status returns something other than 'in progress'.
request_status="in progress"
while [[ "$request_status" == "in progress" ]]; do
  sleep 10
  request_status=$(xcrun altool --notarization-info $uuid --username "$AppleUsername" --password "$ApplePassword" 2>&1 | awk -F ': ' '/Status:/ { print $2; }' )
done

if [[ $request_status != "success" ]]; then
  echo "Failed to notarize."
  exit 1
fi

# Attach the notarization to the DMG.
xcrun stapler staple "$baseName".dmg
rm -rf Telegraf.app
rm -rf $extractedFolder
ls

echo "Signed and notarized!"
