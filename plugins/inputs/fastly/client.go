package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"log"
	"os"
)

// ensureFastlyClients makes sure we have a pair of non-zero Fastly clients.
// This allows us to avoid re-creating the client on every Gather().
func (f *Fastly) ensureFastlyClients() error {
	if f.rtClient != nil {
		return nil
	}
	client, err := fastly.NewClient(f.ApiKey)
	if err == nil {
		log.Println("D! [inputs.fastly] Initializing new Fastly client.")
		f.client = client
	}
	// go-fastly has some inconsistency between NewClient (above) and
	// NewRealtimeStatsClient below. The API keys are passed in via env var
	// in the latter.
	if err = os.Setenv(fastly.APIKeyEnvVar, f.ApiKey); err != nil {
		return err
	}
	// Also, errors while creating the realtime client result in a panic.
	f.rtClient = fastly.NewRealtimeStatsClient()
	return err
}
