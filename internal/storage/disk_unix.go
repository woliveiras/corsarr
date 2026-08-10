//go:build !windows

package storage

import "syscall"

func availableDiskBytes(path string) (uint64, error) {
	var status syscall.Statfs_t
	if err := syscall.Statfs(path, &status); err != nil {
		return 0, err
	}
	return status.Bavail * uint64(status.Bsize), nil
}
