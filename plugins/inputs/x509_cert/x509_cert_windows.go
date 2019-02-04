// +build windows

package x509_cert

import (
	"crypto/x509"
	"fmt"
	"syscall"
	"unsafe"
)

//storeName may be: My, Root, AuthRoot, CA, AddressBook, TrustedPeople, TrustedPublisher, Disallowed
func loadCertificatesFromWinStore(location string, storeName string) ([]*x509.Certificate, error) {
	const (
		CRYPT_E_NOT_FOUND                          = 0x80092004
		CERT_STORE_PROV_SYSTEM_W           uintptr = 10
		CERT_SYSTEM_STORE_CURRENT_USER_ID  uintptr = 1
		CERT_SYSTEM_STORE_LOCAL_MACHINE_ID uintptr = 2
		CERT_SYSTEM_STORE_LOCATION_SHIFT   uintptr = 16
		CERT_STORE_READONLY_FLAG                   = 0x00008000
		CERT_SYSTEM_STORE_CURRENT_USER             = uint32(CERT_SYSTEM_STORE_CURRENT_USER_ID << CERT_SYSTEM_STORE_LOCATION_SHIFT)
		CERT_SYSTEM_STORE_LOCAL_MACHINE            = uint32(CERT_SYSTEM_STORE_LOCAL_MACHINE_ID << CERT_SYSTEM_STORE_LOCATION_SHIFT)
	)

	var locations = map[string]uint32{
		"LocalMachine": CERT_SYSTEM_STORE_LOCAL_MACHINE,
		"CurrentUser":  CERT_SYSTEM_STORE_CURRENT_USER,
	}
	store, err := syscall.CertOpenStore(
		CERT_STORE_PROV_SYSTEM_W,
		0,
		0,
		locations[location]|CERT_STORE_READONLY_FLAG,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(storeName))))
	if err != nil {
		return nil, fmt.Errorf("failed to load cert - - %s\n", err.Error())
	}
	fmt.Println(store)
	defer syscall.CertCloseStore(store, 0)
	var certificates []*x509.Certificate
	var cert *syscall.CertContext
	for {
		cert, err = syscall.CertEnumCertificatesInStore(store, cert)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == CRYPT_E_NOT_FOUND {
					break
				}
			}
			return nil, err
		}
		if cert == nil {
			break
		}
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		if c, err := x509.ParseCertificate(buf2); err == nil {
			certificates = append(certificates, c)
		}
	}
	return certificates, nil
}
