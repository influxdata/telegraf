package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"log"
	"time"
)

// ensureFastlyServiceList populates the service list then starts a goroutine
// to handle periodic updates.
func (f *Fastly) ensureFastlyServiceList() error {
	if f.services != nil {
		return nil
	}
	log.Println("D! [inputs.fastly] Initializing Fastly service list.")
	if err := f.updateFastlyServiceList(); err != nil {
		return err
	}
	go f.periodicallyUpdateFastlyServiceList()
	return nil
}

func (f *Fastly) updateFastlyServiceList() error {
	services, err := f.client.ListServices(&fastly.ListServicesInput{})
	if err != nil {
		return err
	}
	f.services = services
	log.Println("D! [inputs.fastly] Updated Fastly services list. Found:", len(f.services))
	return nil
}

// periodicallyUpdateFastlyServiceList is ran as a goroutine so that we're
// not bogging down our Gather() calls.
func (f *Fastly) periodicallyUpdateFastlyServiceList() {
	log.Println("I! [inputs.fastly] Fastly service list update period is:", f.ServiceUpdatePeriod)
	for range time.Tick(f.ServiceUpdatePeriod.Duration) {
		if err := f.updateFastlyServiceList(); err != nil {
			log.Println("E! [inputs.fastly] error during Fastly service list update:", err)
		}
	}
}
