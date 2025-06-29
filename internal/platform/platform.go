package platform

import "runtime"

// GetPlatform returns the current operating system and architecture.
func GetPlatform() (os string, arch string) {
	return runtime.GOOS, runtime.GOARCH
}
