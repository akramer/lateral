package platform

import "syscall"

// Getsid implements the missing half of Setsid in syscall
func Getsid(pid int) (sid int, err error) {
	r0, _, e1 := syscall.Syscall(syscall.SYS_GETSID, uintptr(pid), 0, 0)
	sid = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

// Getfd returns true if fd is close on exec
func Getfd(fd int) (cloexec bool, err error) {
	r0, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), 0, 0)
	cloexec = int(r0)&syscall.FD_CLOEXEC != 0
	if e1 != 0 {
		err = e1
	}
	return
}

// Stat performs a stat without creating an os.File that has close-on-garbage-collect semantics
func stat(fd int) (*syscall.Stat_t, error) {
	var stat syscall.Stat_t
	err := syscall.Fstat(fd, &stat)
	return &stat, err
}
