package prometheus_client

import (
	"fmt"
	"testing"

	"github.com/mdlayher/vsock"
	"github.com/stretchr/testify/require"
)

// requireVsock skips unless this machine can actually bind a vsock listener.
//
// A successful vsock.ContextID() is not enough: in a privileged container on
// Docker Desktop /dev/vsock exists and ContextID() returns VMADDR_CID_ANY
// (0xffffffff), the wildcard, because the VM has no guest transport. Binding
// then fails with ENOSYS, so a guard on the error alone lets the test through
// on a host where it cannot pass.
func requireVsock(t *testing.T) uint32 {
	t.Helper()

	cid, err := vsock.ContextID()
	if err != nil {
		t.Skipf("vsock not available on this host: %v", err)
	}
	if cid == 0xffffffff {
		t.Skip("vsock reports the wildcard context ID, so this host has no guest transport")
	}
	return cid
}

// An address naming a CID must bind that CID. It used to be split off and
// discarded, so the listener came up on the local context ID and a scraper
// addressing the configured one never reached it (#19775).
func TestListenVsockBindsTheConfiguredContextID(t *testing.T) {
	cid := requireVsock(t)

	listener, err := listenVsock(fmt.Sprintf("%d:9273", cid))
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*vsock.Addr)
	require.True(t, ok)
	require.Equal(t, cid, addr.ContextID)
	require.Equal(t, uint32(9273), addr.Port)
}

// The documented form carries no CID, and binding the local context ID is
// correct for it. This is where prometheus_client differs from
// inputs.socket_listener, so the fix must not route it through
// ListenContextID.
func TestListenVsockWithoutContextIDBindsLocal(t *testing.T) {
	cid := requireVsock(t)

	listener, err := listenVsock(":9274")
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*vsock.Addr)
	require.True(t, ok)
	require.Equal(t, cid, addr.ContextID)
	require.Equal(t, uint32(9274), addr.Port)
}

// A CID that cannot belong to this machine has to be refused rather than
// silently replaced with our own.
func TestListenVsockRejectsForeignContextID(t *testing.T) {
	requireVsock(t)

	// The highest context ID below the wildcard is never ours, and the exact
	// errno depends on the kernel, so only require that binding is refused and
	// that the CID is named.
	_, err := listenVsock("4294967294:9275")
	require.ErrorContains(t, err, "listening on CID 4294967294 failed")
}

func TestListenVsockRejectsNonNumericContextID(t *testing.T) {
	_, err := listenVsock("host:9276")
	require.ErrorContains(t, err, "failed to parse CID host")
}
