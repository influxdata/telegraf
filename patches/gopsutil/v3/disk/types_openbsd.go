//go:build ignore
// +build ignore

// Hand writing: _Ctype_struct___0

/*
Input to cgo -godefs.
*/

package disk

/*
#include <sys/types.h>
#include <sys/disk.h>
#include <sys/mount.h>
*/
import "C"

const (
	devstat_NO_DATA = 0x00
	devstat_READ    = 0x01
	devstat_WRITE   = 0x02
	devstat_FREE    = 0x03
)

const (
	sizeOfDiskstats = C.sizeof_struct_diskstats
)

type (
	Diskstats C.struct_diskstats
	Timeval   C.struct_timeval
)

type (
	Diskstat C.struct_diskstat
	bintime  C.struct_bintime
)
