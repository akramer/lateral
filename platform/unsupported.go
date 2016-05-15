// +build !linux,!freebsd

package platform

import "fmt"

func Getexe() (string, error) {
	return "", fmt.Errorf("This platform is unsupported")
}
