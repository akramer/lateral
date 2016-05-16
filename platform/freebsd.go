// +build freebsd

package platform

import (
	"fmt"
	"os"
)

func Getexe() (string, error) {
	pid := os.Getpid()
	return fmt.Sprintf("/proc/%d/file", pid), nil
}

func s_isreg(v uint16) bool {
	return v&0170000 == 0100000
}

func s_isfifo(v uint16) bool {
	return v&0170000 == 0010000
}
