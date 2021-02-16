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
tar -czvf "$extractedFolder"_darwin_amd64.tar.gz $extractedFolder
rm -rf $extractedFolder
ls
