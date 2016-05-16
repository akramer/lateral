package platform

import (
	"fmt"
	"os"
	"strconv"
)

func GetFds() ([]int, error) {
	f, err := os.Open("/dev/fd")
	if err != nil {
		return nil, err
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	f.Close()

	var fds []int
	for _, name := range files {
		i, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		// Only append filedescriptors that clearly don't have anything to do with
		// the go runtime. On linux there's an epoll fd, and maybe more.
		// Any open regular files or pipes should be fair game, as are stdin/out/err.
		s, err := stat(i)
		if err != nil {
			continue
		}
		cloexec, err := Getfd(i)
		if err != nil {
			continue
		}
		fmt.Printf("fd %d: cloexec: %t\n", i, cloexec)
		fmt.Printf("fd %d: mode: %o\n", i, s.Mode&0170000)
		if s_isreg(s.Mode) { // Add any regular files
			fmt.Printf("Appending filedescriptor for being a regular file %d\n", i)
			fds = append(fds, i)
		} else if s_isfifo(s.Mode) { // add pipes
			fmt.Printf("Appending filedescriptor for being a pipe %d\n", i)
			fds = append(fds, i)
		} else if i < 3 {
			fmt.Printf("Appending filedescriptor for < 3: %d\n", i)
			fds = append(fds, i)
		} else {
			fmt.Printf("Ignoring fd %d\n", i)
		}
	}
	return fds, nil
}
