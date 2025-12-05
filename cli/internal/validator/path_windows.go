//go:build windows

package validator

import (
	"syscall"
	"unsafe"
)

// getAvailableDiskSpaceGB returns available disk space in GB (Windows)
func getAvailableDiskSpaceGB(path string) float64 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}

	ret, _, _ := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return 0
	}

	// Convert bytes to GB
	availableGB := float64(freeBytesAvailable) / (1024 * 1024 * 1024)
	return availableGB
}
