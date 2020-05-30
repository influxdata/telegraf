package opcua

import (
	"testing"

	copier "github.com/jinzhu/copier"
)

// struct for generic test validation
type TestValidation struct {
	Name           string
	Input          interface{}
	ExpectedOutput interface{}
	ActualOutput   interface{}
}

// Check - performs check of TestValidation object, comparing ExpectedOutput to ActualOutput
func Check(tv TestValidation, t *testing.T) {
	if tv.ExpectedOutput != tv.ActualOutput {
		t.Errorf("\nERROR on %s\n\ninput: %s\nexpected: %s\nactual: %s\n", tv.Name, tv.Input, tv.ExpectedOutput, tv.ActualOutput)
	}
}

// reusable vars
var defaultOpcua = &Opcua{
	Endpoint:   "opc.tcp://localhost:50000",
	Policy:     "Auto",
	Mode:       "Auto",
	CertFile:   "",
	KeyFile:    "",
	AuthMethod: "Anonymous",
}

var certificateOpcua = &Opcua{
	Endpoint:   "opc.tcp://localhost:50000",
	Policy:     "Auto",
	Mode:       "Auto",
	CertFile:   "/home/app/cert.pem",
	KeyFile:    "/home/app/key.pem",
	AuthMethod: "Certificate",
}

// Start of tests

// #######################################
// ###### TestHasSharedAccessKeyName #####
// #######################################

// Tests to run, format: input, expected output, actual output
var TableTests = []TestValidation{
	{"Default OPC UA client", defaultOpcua, "", nil},
}

func TestHasErrorMessage(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("##### Running tests for TestHasErrorMessage #####")
	// for each TestValidation item in TableTests
	for _, row := range TableTests {
		t.Logf("Testing against: %s\n", row.Name)
		temp := &Opcua{}
		copier.Copy(temp, row.Input)
		result := temp.Description()
		// update TestValidation item with result
		row.ActualOutput = result
		// check result against expected result
		Check(row, t)
	}
}
