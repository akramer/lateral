// +build linux

package platform

func Getexe() (string, error) {
	return "/proc/self/exe", nil
}
