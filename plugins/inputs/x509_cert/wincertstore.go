//go:build windows

package x509_cert

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type wincertstore struct {
	store  *uint16
	flags  uint32
	source string
}

func getWincertStore(path string) (*wincertstore, error) {
	var flags uint32 = windows.CERT_STORE_READONLY_FLAG | windows.CERT_STORE_OPEN_EXISTING_FLAG

	// Accept store names containing locations of the forms [<location>:]name
	location, folder, found := strings.Cut(path, ":")
	if !found {
		location = "machine"
		folder = path
	}
	source := location + ":" + folder
	switch location {
	case "machine":
		flags |= windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
	case "user":
		flags |= windows.CERT_SYSTEM_STORE_CURRENT_USER
	default:
		return nil, fmt.Errorf("unknown store location %q", location)
	}
	store, err := windows.UTF16PtrFromString(folder)
	if err != nil {
		return nil, fmt.Errorf("converting store folder %q failed: %w", folder, err)
	}

	return &wincertstore{store: store, flags: flags, source: source}, nil
}

func (c *X509Cert) processWinCertStore(path string) ([]*x509.Certificate, error) {
	// Get the store parameters
	store, err := getWincertStore(path)
	if err != nil {
		return nil, fmt.Errorf("configuring store %q failed: %w", path, err)
	}

	// Open the actual store for reading
	handle, err := windows.CertOpenStore(
		windows.CERT_STORE_PROV_SYSTEM_W,
		0,
		0,
		store.flags,
		uintptr(unsafe.Pointer(store.store)), //nolint:gosec // G103: Valid use of unsafe call to pass store to API
	)
	if err != nil {
		return nil, fmt.Errorf("opening store %q failed: %w", store.source, err)
	}
	defer windows.CertCloseStore(handle, 0)

	// Enumerate all available certificates
	var certificates []*x509.Certificate
	var certctx *windows.CertContext
	for {
		// Get the next certificate in the store
		var err error
		if certctx, err = windows.CertEnumCertificatesInStore(handle, certctx); certctx == nil {
			break
		} else if err != nil {
			if errors.Is(err, windows.Errno(windows.CRYPT_E_NOT_FOUND)) {
				return certificates, nil
			}
			if err := windows.CertFreeCertificateContext(certctx); err != nil {
				c.Log.Errorf("Freeing context for store %q failed: %v", store.source, err)
			}
			return nil, fmt.Errorf("enumerating certificates in %q failed: %w", store.source, err)
		}

		// Convert the returned byte pointer into an usable byte-slice and parse
		// the certificate. We need to copy the byte-slice to avoid
		// modifications during processing...
		buf := unsafe.Slice(certctx.EncodedCert, certctx.Length) //nolint:gosec // G103: Valid use of unsafe call to extract cert data
		cert, err := x509.ParseCertificate(bytes.Clone(buf))
		if err != nil {
			name := make([]uint16, 256)
			n := windows.CertGetNameString(certctx, windows.CERT_NAME_SIMPLE_DISPLAY_TYPE, 0, nil, &name[0], uint32(len(name)))
			subject := windows.UTF16ToString(name[:n])
			c.Log.Errorf("parsing certificate for %q in store %q failed: %v", subject, store.source, err)
			continue
		}
		certificates = append(certificates, cert)
	}
	return certificates, nil
}
