package tls

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/youmark/pkcs8"
)

// ParseCiphers returns a `[]uint16` by received `[]string` key that represents ciphers from crypto/tls.
// If some of ciphers in received list doesn't exists  ParseCiphers returns nil with error
func ParseCiphers(ciphers []string) ([]uint16, error) {
	suites := []uint16{}

	for _, cipher := range ciphers {
		v, ok := tlsCipherMap[cipher]
		if !ok {
			return nil, fmt.Errorf("unsupported cipher %q", cipher)
		}
		suites = append(suites, v)
	}

	return suites, nil
}

// ParseTLSVersion returns a `uint16` by received version string key that represents tls version from crypto/tls.
// If version isn't supported ParseTLSVersion returns 0 with error
func ParseTLSVersion(version string) (uint16, error) {
	if v, ok := tlsVersionMap[version]; ok {
		return v, nil
	}

	available := make([]string, 0, len(tlsVersionMap))
	for n := range tlsVersionMap {
		available = append(available, n)
	}
	sort.Strings(available)
	return 0, fmt.Errorf("unsupported version %q (available: %s)", version, strings.Join(available, ","))
}

func readFile(filename string) []byte {
	octets, err := os.ReadFile(filename)
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
	keyBytes := readFile(filename)
	currentBlock, remainingBlocks := pem.Decode(keyBytes)
	if currentBlock == nil {
		panic(errors.New("failed to decode private key: no PEM data found"))
	}
	var allBlocks string
	for {
		if currentBlock.Type == "ENCRYPTED PRIVATE KEY" {
			if password == "" {
				panic(errors.New("missing password for PKCS#8 encrypted private key"))
			}
			var decryptedKey *rsa.PrivateKey
			decryptedKey, err := pkcs8.ParsePKCS8PrivateKeyRSA(currentBlock.Bytes, []byte(password))
			if err != nil {
				panic(fmt.Errorf("failed to parse encrypted PKCS#8 private key: %w", err))
			}
			pemBlock := string(pem.EncodeToMemory(&pem.Block{Type: currentBlock.Type, Bytes: x509.MarshalPKCS1PrivateKey(decryptedKey)}))
			allBlocks += pemBlock
		} else if currentBlock.Headers["Proc-Type"] == "4,ENCRYPTED" {
			decryptedKeyDER, err := x509.DecryptPEMBlock(currentBlock, []byte(password))
			if err != nil {
				panic(fmt.Errorf("failed to parse encrypted private key %w", err))
			}
			decryptedKey, err := x509.ParsePKCS1PrivateKey(decryptedKeyDER)
			if err != nil {
				panic(fmt.Errorf("unable to convert from DER to PEM format: %w", err))
			}
			pemBlock := string(pem.EncodeToMemory(&pem.Block{Type: currentBlock.Type, Bytes: x509.MarshalPKCS1PrivateKey(decryptedKey)}))
			allBlocks += pemBlock
		} else {
			pemBlock := string(pem.EncodeToMemory(&pem.Block{Type: currentBlock.Type, Bytes: currentBlock.Bytes}))
			allBlocks += pemBlock
		}
		currentBlock, remainingBlocks = pem.Decode(remainingBlocks)
		if currentBlock == nil {
			break
		}
	}
	return allBlocks
}
