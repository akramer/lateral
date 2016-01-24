// Implement the getsid system call.
// This appears to exist on Linux and FreeBSD at the very least.
package getsid

import "syscall"

func Getsid(pid int) (sid int, err error) {
	r0, _, e1 := syscall.RawSyscall(syscall.SYS_GETSID, uintptr(pid), 0, 0)
	sid = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}
