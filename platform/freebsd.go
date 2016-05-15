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
