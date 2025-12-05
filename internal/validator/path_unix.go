//go:build unix || darwin || linux

package validator

import "syscall"

// getAvailableDiskSpaceGB returns available disk space in GB (Unix/Linux/macOS)
func getAvailableDiskSpaceGB(path string) float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0
	}

	// Available blocks * block size / GB
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)
	
	return availableGB
}
