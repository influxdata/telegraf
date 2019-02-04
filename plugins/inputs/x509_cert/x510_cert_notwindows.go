// +build !windows
package x509_cert

import (
	"crypto/x509"
	"fmt"
)

//storeName may be: My, Root, AuthRoot, CA, AddressBook, TrustedPeople, TrustedPublisher, Disallowed
func loadCertificatesFromWinStore(location string, storeName string) ([]*x509.Certificate, error) {
	return nil, fmt.Errorf("windows stores works only in windows")
}
