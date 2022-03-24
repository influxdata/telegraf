//go:build ignore
// +build ignore

// Hand writing: _Ctype_struct___3, 4

/*
Input to cgo -godefs.

*/

package net

/*
#include <sys/types.h>
#include <sys/socketvar.h>
#include <sys/proc_info.h>
#include <netinet/in_pcb.h>

enum {
	sizeofPtr = sizeof(void*),
};

*/
import "C"

// Machine characteristics; for internal use.

const (
	sizeofPtr        = C.sizeofPtr
	sizeofShort      = C.sizeof_short
	sizeofInt        = C.sizeof_int
	sizeofLong       = C.sizeof_long
	sizeofLongLong   = C.sizeof_longlong
	sizeofLongDouble = C.sizeof_longlong
)

// Basic types

type (
	_C_short       C.short
	_C_int         C.int
	_C_long        C.long
	_C_long_long   C.longlong
	_C_long_double C.longlong
)

type (
	Xinpgen          C.struct_xinpgen
	Inpcb            C.struct_inpcb
	in_addr          C.struct_in_addr
	Inpcb_list_entry C.struct__inpcb_list_entry
	Xsocket          C.struct_xsocket
	Xsockbuf         C.struct_xsockbuf
	Xinpcb           C.struct_xinpcb
)

// type u_quad_t C.struct_u_quad_t
