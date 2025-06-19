package sys

import "runtime"

func GetInfo() (os, arch string) {
	return runtime.GOOS, runtime.GOARCH
}
