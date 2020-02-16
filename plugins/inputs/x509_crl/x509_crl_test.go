package x509_crl

import (
	"encoding/base64"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

// Make sure X509CRL implements telegraf.Input
var _ telegraf.Input = &X509CRL{}

var InvalidCRL = fmt.Sprintf("-----BEGIN X509 CRL-----\n%s\n-----END X509 CRL-----\n", base64.StdEncoding.EncodeToString([]byte("Invalid CRL")))
var ValidCRL = fmt.Sprintf("-----BEGIN X509 CRL-----\n%s\n-----END X509 CRL-----\n", `MIIB2zCBxDANBgkqhkiG9w0BAQsFADCBlDELMAkGA1UEBhMCRlIxDzANBgNVBAgT
BkFsc2FjZTETMBEGA1UEBxMKU3RyYXNib3VyZzEdMBsGA1UEChMUQWxzYWNlIFJl
c2VhdSBOZXV0cmUxCzAJBgNVBAMTAmFjMQ8wDQYDVQQpEwZBQyBWUE4xIjAgBgkq
hkiG9w0BCQEWE2V4YW1wbGVAZXhhbXBsZS5jb20XDTIwMDIwNTE1NDUyM1oXDTIw
MDMwNjE1NDUyM1owDQYJKoZIhvcNAQELBQADggEBAMSRAPF5HIUfffXAxddjGfsS
bos+LN3YkPV6RnHzddLviwmJEDTrrU+j7I1aMcbSPvYXfLCyRqgIH8KaXp3WzOlb
EwMY6e3O2MHV0xnFG57O/OGwp8EDZt+LWhxPqoIdu67dWzyZabQtoWTxXBpabNk/
ffLXxnj9p+9oBTbdV0uEzau9rZ8o0fER8+2KJg09x5QtRkb6DHKinN1m6wjWTZcp
ketfLeZgB1eH8Gg0QDU5nXls34Eenqx9vpZly/LY/WgT4Oy70mvzfnBvsx+0kLXT
NE30kCyqvZVWvnJ3abSkIEUk6vZ9oeJCK5xA3Sfikw2RWrgrmJ8fB4pphp+QxhY=
`)

func TestWhenSourceIsNotFileGotFailure(test *testing.T) {
	if testing.Short() {
		test.Skip("Skipping network-dependent parametrizedTest in short fileMode.")
	}

	parametrizedTests := []struct {
		testName   string
		source     string
		shouldFail bool
	}{
		{testName: "HTTPS", source: "https://example.org:443", shouldFail: true},
		{testName: "UDP", source: "udp://example.org:443", shouldFail: true},
		{testName: "UDP4", source: "udp4://example.org:443", shouldFail: true},
		{testName: "UDP6", source: "udp6://example.org:443", shouldFail: true},
		{testName: "TCP", source: "tcp://example.org:443", shouldFail: true},
		{testName: "TCP4", source: "tcp4://example.org:443", shouldFail: true},
		{testName: "TCP6", source: "tcp6://example.org:443", shouldFail: true},
		{testName: "Unsupported scheme", source: "foo://", shouldFail: true},
	}

	for _, parametrizedTest := range parametrizedTests {
		test.Run(parametrizedTest.testName, func(test *testing.T) {

			x509crl := X509CRL{
				Sources: []string{parametrizedTest.source},
			}
			_ = x509crl.Init()

			testErr := false

			acc := testutil.Accumulator{}
			err := x509crl.Gather(&acc)
			if len(acc.Errors) > 0 {
				testErr = true
			}

			if testErr != parametrizedTest.shouldFail {
				test.Errorf("%s", err)
			}
		})
	}
}

func TestWhenSourceIsFile(test *testing.T) {
	parametrizedTests := []struct {
		testName    string
		fileMode    os.FileMode
		fileContent string
		shouldFail  bool
	}{
		{testName: "Permission denied", fileMode: 0001, shouldFail: true},
		{testName: "Not a CRL", fileMode: 0640, fileContent: "not a CRL", shouldFail: true},
		{testName: "Wrong CRL", fileMode: 0640, fileContent: InvalidCRL, shouldFail: true},
		{testName: "Correct CRL", fileMode: 0640, fileContent: ValidCRL, shouldFail: false},
		{testName: "Correct CRL and extra trailing space", fileMode: 0640, fileContent: ValidCRL + " ", shouldFail: false},
		{testName: "Correct CRL and extra leading space", fileMode: 0640, fileContent: " " + ValidCRL, shouldFail: false},
		{testName: "Correct multiple CRLs", fileMode: 0640, fileContent: ValidCRL + ValidCRL, shouldFail: false},
		{testName: "Correct CRL and wrong CRL", fileMode: 0640, fileContent: ValidCRL + "\n" + InvalidCRL, shouldFail: true},
		{testName: "Correct CRL and not a CRL", fileMode: 0640, fileContent: ValidCRL + "\nparametrizedTest", shouldFail: true},
		{testName: "Correct multiple CRLs and extra trailing space", fileMode: 0640, fileContent: ValidCRL + ValidCRL + " ", shouldFail: false},
		{testName: "Correct multiple CRLs and extra leading space", fileMode: 0640, fileContent: " " + ValidCRL + ValidCRL, shouldFail: false},
		{testName: "Correct multiple CRLs and extra middle space", fileMode: 0640, fileContent: ValidCRL + " " + ValidCRL, shouldFail: false},
	}

	for _, parametrizedTest := range parametrizedTests {
		test.Run(parametrizedTest.testName, func(test *testing.T) {
			crlFile := givenCRLFile(test, parametrizedTest.fileMode, parametrizedTest.fileContent)
			defer os.Remove(crlFile.Name())

			x509crl := X509CRL{
				Sources: []string{crlFile.Name()},
			}
			_ = x509crl.Init()

			failed := false

			acc := testutil.Accumulator{}
			err := x509crl.Gather(&acc)
			if len(acc.Errors) > 0 {
				failed = true
			}

			if failed != parametrizedTest.shouldFail {
				test.Errorf("%s", err)
			}
		})
	}
}

func TestTags(test *testing.T) {
	crlFile := givenCRLFile(test, 0640, ValidCRL)
	defer os.Remove(crlFile.Name())

	x509crl := &X509CRL{
		Sources: []string{crlFile.Name()},
	}
	_ = x509crl.Init()

	acc := testutil.Accumulator{}
	err := x509crl.Gather(&acc)
	require.NoError(test, err)

	assert.True(test, acc.HasMeasurement(x509CrlMeasurement))
	thenTagIsPresentAndEquals(test, &acc, "source", crlFile.Name())
	thenTagIsPresentAndEquals(test, &acc, "issuer", "1.2.840.113549.1.9.1=#0c136578616d706c65406578616d706c652e636f6d,2.5.4.41=#130641432056504e,CN=ac,O=Alsace Reseau Neutre,L=Strasbourg,ST=Alsace,C=FR")
	thenTagIsPresentAndEquals(test, &acc, "version", "0")
}

func TestFields(test *testing.T) {
	crlFile := givenCRLFile(test, 0640, ValidCRL)
	defer os.Remove(crlFile.Name())

	x509crl := &X509CRL{
		Sources: []string{crlFile.Name()},
	}
	_ = x509crl.Init()

	acc := testutil.Accumulator{}
	err := x509crl.Gather(&acc)
	require.NoError(test, err)

	assert.True(test, acc.HasMeasurement(x509CrlMeasurement))
	thenFieldIsPresentAndEquals(test, &acc, "start_date", 1580917523000)
	thenFieldIsPresentAndEquals(test, &acc, "end_date", 1583509523000)
	thenFieldIsPresentAndEquals(test, &acc, "revoked_certificates", 0)

	effective, _ := acc.BoolField(x509CrlMeasurement, "has_expired")
	assert.True(test, acc.HasField(x509CrlMeasurement, "has_expired"), fmt.Sprintf("Field %s not present", "has_expired"))
	assert.Equal(test, false, effective, fmt.Sprintf("Invalid field '%s'", "has_expired"))
}

func TestStrings(test *testing.T) {

	crlFile := givenCRLFile(test, 0640, ValidCRL)
	defer os.Remove(crlFile.Name())

	x509crl := &X509CRL{
		Sources: []string{crlFile.Name()},
	}

	_ = x509crl.Init()

	parametrizedTests := []struct {
		testName  string
		method    string
		effective string
		expected  string
	}{
		{testName: "description", method: "Description", effective: x509crl.Description(), expected: description},
		{testName: "sample config", method: "SampleConfig", effective: x509crl.SampleConfig(), expected: sampleConfig},
	}

	for _, parametrizedTest := range parametrizedTests {
		test.Run(parametrizedTest.testName, func(test *testing.T) {
			if parametrizedTest.effective != parametrizedTest.expected {
				test.Errorf("Expected method %s to return '%s', found '%s'.", parametrizedTest.method, parametrizedTest.expected, parametrizedTest.effective)
			}
		})
	}
}

// Given
func givenCRLFile(test *testing.T, fileMode os.FileMode, fileContent string) *os.File {
	crlFile, err := ioutil.TempFile("", "x509crl_tmp_file")
	if err != nil {
		test.Fatal(err)
	}

	_, err = crlFile.Write([]byte(fileContent))
	if err != nil {
		test.Fatal(err)
	}

	err = crlFile.Chmod(fileMode)
	if err != nil {
		test.Fatal(err)
	}

	err = crlFile.Close()
	if err != nil {
		test.Fatal(err)
	}

	if crlFile == nil {
		test.Fatal()
	}

	return crlFile
}

// Then
func thenTagIsPresentAndEquals(test *testing.T, acc *testutil.Accumulator, tag string, expected string) {
	assert.True(test, acc.HasTag(x509CrlMeasurement, tag), fmt.Sprintf("Tag %s not present", tag))
	assert.Equal(test, expected, acc.TagValue(x509CrlMeasurement, tag), fmt.Sprintf("Invalid tag '%s'", tag))
}

func thenFieldIsPresentAndEquals(test *testing.T, acc *testutil.Accumulator, field string, expected int64) {
	effective, _ := acc.Int64Field(x509CrlMeasurement, field)
	assert.True(test, acc.HasField(x509CrlMeasurement, field), fmt.Sprintf("Field %s not present", field))
	assert.Equal(test, expected, effective, fmt.Sprintf("Invalid field '%s'", field))
}
