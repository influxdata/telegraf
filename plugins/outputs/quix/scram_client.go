// scram_client.go
package quix

import (
	"github.com/xdg-go/scram"
)

// XDGSCRAMClient wraps the SCRAM client for SCRAM-SHA-256 authentication
type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	HashGeneratorFcn scram.HashGeneratorFcn
}

// Begin initializes the SCRAM client with username and password
func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
	client, err := x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.Client = client
	x.ClientConversation = client.NewConversation()
	return nil
}

// Step processes the server's challenge and returns the client's response
func (x *XDGSCRAMClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}

// Done returns true if the SCRAM conversation is complete
func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

// Define SHA256 and SHA512 hash generators
var SHA256 scram.HashGeneratorFcn = scram.SHA256
var SHA512 scram.HashGeneratorFcn = scram.SHA512
