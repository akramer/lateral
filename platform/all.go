package platform

import (
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
		if s_isreg(s.Mode) { // Add any regular files
			fds = append(fds, i)
		} else if s_isfifo(s.Mode) { // add pipes
			fds = append(fds, i)
		} else if i < 3 {
			fds = append(fds, i)
		}
	}
	return fds, nil
}
