//go:build !windows

package procutil

import "syscall"

// DetachSysProcAttr returns the SysProcAttr needed to start a child
// process in its own session so it survives the parent CLI exiting and
// the controlling terminal closing.
func DetachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
