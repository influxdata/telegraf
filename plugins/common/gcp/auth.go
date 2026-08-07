package gcp

import (
	"encoding/json"
	"fmt"
	"os"
)

func ParseCredentialType(credentialsFile string) (string, error) {
	serviceAccount, err := os.ReadFile(credentialsFile)
	if err != nil {
		return "", fmt.Errorf("cannot load the credential file: %w", err)
	}

	type fileTypeChecker struct {
		Type string `json:"type"`
	}
	var f fileTypeChecker
	if err := json.Unmarshal(serviceAccount, &f); err != nil {
		return "", fmt.Errorf("cannot parse the credential file: %w", err)
	}

	return f.Type, nil
}
