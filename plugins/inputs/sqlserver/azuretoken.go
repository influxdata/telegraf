package sqlserver

import "time"

// New token structure for Azure Identity SDK
type azureToken struct {
	token     string
	expiresOn time.Time
}

// IsExpired helper method for Azure token expiry
func (t *azureToken) IsExpired() bool {
	if t == nil {
		return true
	}
	return time.Now().After(t.expiresOn)
}
