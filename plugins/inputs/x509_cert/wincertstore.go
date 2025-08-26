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

	"github.com/influxdata/telegraf"
)

type wincertstore struct {
	store  *uint16
	flags  uint32
	source string
	log    telegraf.Logger
}

func newWincertStore(path string, log telegraf.Logger) (*wincertstore, error) {
	var flags uint32 = windows.CERT_STORE_READONLY_FLAG | windows.CERT_STORE_OPEN_EXISTING_FLAG

	// Accept store names containing locations of the forms [<location>:]name
	var source string
	before, name, found := strings.Cut(path, ":")
	if !found {
		flags |= windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
		name = path
		source = "HKEY_LOCAL_MACHINE:" + name
	} else {
		switch before {
		case "HKLM", "HKEY_LOCAL_MACHINE":
			flags |= windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
			source = "HKEY_LOCAL_MACHINE:" + name
		case "HKCU", "HKEY_CURRENT_USER":
			flags |= windows.CERT_SYSTEM_STORE_CURRENT_USER
			source = "HKEY_CURRENT_USER:" + name
		default:
			return nil, fmt.Errorf("unknown store location %q", before)
		}
	}
	store, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("converting store name %q failed: %w", name, err)
	}

	return &wincertstore{store: store, flags: flags, source: source, log: log}, nil
}

func (s *wincertstore) read() ([]*x509.Certificate, error) {
	// Open the actual store for reading
	handle, err := windows.CertOpenStore(
		windows.CERT_STORE_PROV_SYSTEM_W,
		0,
		0,
		s.flags,
		uintptr(unsafe.Pointer(s.store)), //nolint:gosec // G103: Valid use of unsafe call to pass store to API
	)
	if err != nil {
		return nil, fmt.Errorf("opening store failed: %w", err)
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
				s.log.Errorf("Freeing context for store %q failed: %v", s.source, err)
			}
			return nil, fmt.Errorf("enumerating certificates failed: %w", err)
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
			s.log.Errorf("parsing certificate for %q in store %q failed: %v", subject, s.source, err)
			continue
		}
		certificates = append(certificates, cert)
	}
	return certificates, nil
}
