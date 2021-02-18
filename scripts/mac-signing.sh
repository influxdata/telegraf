base64 -D -o MacCertificate.p12 <<< $MacCertificate
sudo security import MacCertificate.p12 -k /Library/Keychains/System.keychain -P $MacCertificatePassword -A
base64 -D -o AppleSigningAuthorityCertificate.cer <<< $AppleSigningAuthorityCertificate
sudo security import AppleSigningAuthorityCertificate.cer -k '/Library/Keychains/System.keychain' -A

cd dist
tar -xzvf $(find . -name "*darwin_amd64.tar*")
rm $(find . -name "*darwin_amd64.tar*")
cd $(find . -name "*telegraf-*" -type d)
cd usr/bin
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime telegraf
codesign -v telegraf

cd
cd project/dist
extractedFolder=$(find . -name "*telegraf-*" -type d)

codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime ../scripts/telegraf_entry_mac
codesign -v ../scripts/telegraf_entry_mac

mkdir Telegraf
cd Telegraf
mkdir Contents
cd Contents
mkdir MacOS
mkdir Resources
cd ../..
cp ../scripts/info.plist Telegraf/Contents
cp -R "$extractedFolder"/ Telegraf/Contents/Resources
cp ../scripts/telegraf_entry_mac Telegraf/Contents/MacOS
cp ../assets/icon.icns Telegraf/Contents/Resources
mv Telegraf Telegraf.app

codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime --deep --force Telegraf.app
hdiutil create -size 500m -volname Telegraf -srcfolder Telegraf.app telegraf.dmg
codesign -s "Developer ID Application: InfluxData Inc. (M7DN9H35QT)" --timestamp --options=runtime telegraf.dmg

uuid=$(xcrun altool --notarize-app --primary-bundle-id "com.influxdata.telegraf" --username $appleDevUsername --password $appleDevPassword --file telegraf.dmg | awk '/RequestUUID/ { print $NF; }')

if [[ $uuid == "" ]]; then 
  echo "Could not upload for notarization."
  exit 1
fi

# wait for status to be not "in progress" any more
request_status="in progress"
while [[ "$request_status" == "in progress" ]]; do
  sleep 10
  request_status=$(xcrun altool --notarization-info $requestUUID --username $appleDevUsername --password $appleDevPassword 2>&1 | awk -F ': ' '/Status:/ { print $2; }' )
done

if [[ $request_status != "success" ]]; then
  echo "Failed to notarize."
  exit 1
fi

echo "Signed and notarized!"

xcrun stapler staple telegraf.dmg
rm Telegraf.app
rm -rf $extractedFolder
ls
