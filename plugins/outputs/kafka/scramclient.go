package kafka

import (
	"crypto/sha512"
	"fmt"

	"github.com/xdg/scram"
)

// SHA512 hash generator function
var SHA512 scram.HashGeneratorFcn = sha512.New //nolint:gochecknoglobals

// SCRAMClient struct
type SCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

// Begin constructs the SCRAM client and conversation
func (sc *SCRAMClient) Begin(userName, password, authzID string) error {
	var err error

	sc.Client, err = sc.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return fmt.Errorf("ScramClient initialization error")
	}

	sc.ClientConversation = sc.Client.NewConversation()

	return nil
}

// Step attempts to move the authentication conversation forward
func (sc *SCRAMClient) Step(challenge string) (string, error) {
	res, err := sc.ClientConversation.Step(challenge)
	if err != nil {
		return "", fmt.Errorf("Unable to step conversation")
	}

	return res, nil
}

// Done returns whether the conversation is completed
func (sc *SCRAMClient) Done() bool {
	return sc.ClientConversation.Done()
}
