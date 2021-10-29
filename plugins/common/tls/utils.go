package tls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

// ParseCiphers returns a `[]uint16` by received `[]string` key that represents ciphers from crypto/tls.
// If some of ciphers in received list doesn't exists  ParseCiphers returns nil with error
func ParseCiphers(ciphers []string) ([]uint16, error) {
	suites := []uint16{}

	for _, cipher := range ciphers {
		if v, ok := tlsCipherMap[cipher]; ok {
			suites = append(suites, v)
		} else {
			return nil, fmt.Errorf("unsupported cipher %q", cipher)
		}
	}

	return suites, nil
}

// ParseTLSVersion returns a `uint16` by received version string key that represents tls version from crypto/tls.
// If version isn't supported ParseTLSVersion returns 0 with error
func ParseTLSVersion(version string) (uint16, error) {
	if v, ok := tlsVersionMap[version]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unsupported version %q", version)
}

func readFile(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("opening %q: %v", filename, err))
	}
	octets, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("reading %q: %v", filename, err))
	}
	return octets
}

func ReadCertificate(filename string) string {
	octets := readFile(filename)
	return string(octets)
}

func ReadKey(filename string, password string) string {
	octets := readFile(filename)
	var allBlocks string
	currentBlock, remainingBytes := pem.Decode(octets)
	for {
		if x509.IsEncryptedPEMBlock(currentBlock) {
			decryptedBytesDER, err := x509.DecryptPEMBlock(currentBlock, []byte(password))
			if err != nil {
				panic(fmt.Sprintf("incorrect password for key file %q: %v", filename, err))
			}
			decryptedBytesPEM, err := x509.ParsePKCS1PrivateKey(decryptedBytesDER)
			if err != nil {
				panic(fmt.Sprintf("unable to convert from DER to PEM format: %v", err))
			}
			rsaKey := string(pem.EncodeToMemory(
				&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(decryptedBytesPEM),
				},
			))
			allBlocks += rsaKey
		} else {
			cert := string(pem.EncodeToMemory(
				&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: currentBlock.Bytes,
				},
			))
			allBlocks += cert
		}
		currentBlock, remainingBytes = pem.Decode(remainingBytes)
		if currentBlock == nil {
			break
		}
	}
	return allBlocks
}
