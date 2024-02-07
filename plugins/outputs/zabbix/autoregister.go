package zabbix

import (
	"time"
)

// Add adds a host to the list of hosts to send autoregister data to Zabbix.
// Only store information if autoregister is enabled (config Autoregister is not empty).
func (z *Zabbix) autoregisterAdd(hostname string) {
	if z.Autoregister == "" {
		return
	}

	if _, exists := z.autoregisterLastSend[hostname]; !exists {
		z.autoregisterLastSend[hostname] = time.Time{}
	}
}

// Push sends autoregister data to Zabbix for each host.
func (z *Zabbix) autoregisterPush() {
	if z.Autoregister == "" {
		return
	}

	// For each "host" tag seen, send an autoregister request to Zabbix server.
	// z.AutoregisterSendPeriod is the interval at which requests are resend.
	for hostname, timeLastSend := range z.autoregisterLastSend {
		if time.Since(timeLastSend) > time.Duration(z.AutoregisterResendInterval) {
			z.Log.Debugf("Autoregistering host %q", hostname)

			if err := z.sender.RegisterHost(hostname, z.Autoregister); err != nil {
				z.Log.Errorf("Autoregistering host %q: %v", hostname, err)
			}

			z.autoregisterLastSend[hostname] = time.Now()
		}
	}
}
