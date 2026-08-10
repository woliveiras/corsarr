//go:build windows

package storage

import (
	"syscall"
	"unsafe"
)

func availableDiskBytes(path string) (uint64, error) {
	pathPointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var available uint64
	getDiskFreeSpaceEx := syscall.NewLazyDLL("kernel32.dll").NewProc("GetDiskFreeSpaceExW")
	result, _, callErr := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPointer)),
		uintptr(unsafe.Pointer(&available)),
		0,
		0,
	)
	if result == 0 {
		return 0, callErr
	}
	return available, nil
}
