base64 -D -o MacCertificate.p12 <<< $MacCertificate
sudo security import MacCertificate.p12 -k /Library/Keychains/System.keychain -P $MacCertificatePassword -A
base64 -D -o AppleSigningAuthorityCertificate.cer <<< $AppleSigningAuthorityCertificate
sudo security import AppleSigningAuthorityCertificate.cer -k '/Library/Keychains/System.keychain' -A

signingIdentity="Developer ID Application: InfluxData Inc. (M7DN9H35QT)"

cd dist
tar -xzvf $(find . -name "*darwin_amd64.tar*")
rm $(find . -name "*darwin_amd64.tar*")
cd $(find . -name "*telegraf-*" -type d)
cd usr/bin
codesign -s $signingIdentity --timestamp --options=runtime telegraf
codesign -v telegraf

cd
cd project/dist
extractedFolder=$(find . -name "*telegraf-*" -type d)
extractedPath=project/dist/"$extractedFolder"
echo $extractedPath
cp project/scripts/telegraf_entry_mac $extractedPath

echo "now attempting to sign the entry"
codesign -s $signingIdentity --timestamp --options=runtime "$extractedPath"/telegraf_entry
codesign -v "$extractedPath"/telegraf_entry

echo "now calling appmaker"
project/scripts/mac_app_bundler -bin telegraf_entry_mac -identifier com.influxdata.telegraf -name "Telegraf" -o project/dist -assets $extractedPath -icon /project/assets/icon.png

codesign -s $signingIdentity --timestamp --options=runtime --deep Telegraf.app
hdiutil create -size 500m -volname Telegraf -srcfolder Telegraf.app telegraf.dmg
codesign -s $signingIdentity --timestamp --options=runtime telegraf.dmg

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
