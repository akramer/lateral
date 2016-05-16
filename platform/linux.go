// +build linux

package platform

func Getexe() (string, error) {
	return "/proc/self/exe", nil
}

func s_isreg(v uint32) bool {
	return v&0170000 == 0100000
}

func s_isfifo(v uint32) bool {
	return v&0170000 == 0010000
}
