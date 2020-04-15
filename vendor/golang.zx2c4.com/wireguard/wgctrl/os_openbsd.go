//+build openbsd

package wgctrl

import (
	"os"

	"golang.zx2c4.com/wireguard/wgctrl/internal/wginternal"
	"golang.zx2c4.com/wireguard/wgctrl/internal/wgopenbsd"
	"golang.zx2c4.com/wireguard/wgctrl/internal/wguser"
)

// Since the OpenBSD implementation and the code to interact with it are both
// very experimental, make the user explicitly opt-in to use it.
var useKernel = func() bool {
	return os.Getenv("WGCTRL_OPENBSD_KERNEL") == "1"
}()

// newClients configures wginternal.Clients for OpenBSD systems.
func newClients() ([]wginternal.Client, error) {
	var clients []wginternal.Client

	// Make the user opt in explicitly for kernel implementation support.
	if useKernel {
		// OpenBSD has an experimental in-kernel WireGuard implementation:
		// https://git.zx2c4.com/wireguard-openbsd/about/. Determine if it is
		// available and make use of it if so.
		kc, ok, err := wgopenbsd.New()
		if err != nil {
			return nil, err
		}
		if ok {
			clients = append(clients, kc)
		}
	}

	uc, err := wguser.New()
	if err != nil {
		return nil, err
	}

	clients = append(clients, uc)
	return clients, nil
}
