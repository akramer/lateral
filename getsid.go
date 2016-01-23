package main

import "fmt"
import "syscall"

func Getpgid(pid int) (sid int, err error) {
	r0, _, e1 := syscall.RawSyscall(syscall.SYS_GETSID, uintptr(pid), 0, 0)
	sid = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func main() {
  sid, err := Getpgid(0)
  if err != nil {
    panic(err)
  }
  fmt.Printf("%d\n", sid)
}
