// +build linux
// +build arm

package host

type exitStatus struct {
	Etermination int16 // Process termination status.
	Eexit        int16 // Process exit status.
}
type timeval struct {
	TvSec  uint32 // Seconds.
	TvUsec uint32 // Microseconds.
}

type utmp struct {
	Type    int16      // Type of login.
	Pid     int32      // Process ID of login process.
	Line    [32]byte   // Devicename.
	ID      [4]byte    // Inittab ID.
	User    [32]byte   // Username.
	Host    [256]byte  // Hostname for remote login.
	Exit    exitStatus // Exit status of a process marked
	Session int32      // Session ID, used for windowing.
	Tv      timeval    // Time entry was made.
	AddrV6  [16]byte   // Internet address of remote host.
	Unused  [20]byte   // Reserved for future use. // original is 20
}
