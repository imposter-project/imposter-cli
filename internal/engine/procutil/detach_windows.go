//go:build windows

package procutil

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// DetachSysProcAttr returns the SysProcAttr needed to start a child
// process detached from the parent CLI's process group and console so it
// survives the parent exiting and the console closing.
func DetachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
	}
}
