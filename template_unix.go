//go:build !windows

package main

import (
	"os"
	"syscall"
)

// applyOwnership applies Unix file ownership if possible.
func applyOwnership(file *os.File, like os.FileInfo) error {
	likeSys, ok := like.Sys().(*syscall.Stat_t)
	if ok {
		err := file.Chown(int(likeSys.Uid), int(likeSys.Gid))
		if err != nil && !os.IsPermission(err) {
			return err
		}
	}
	return nil
}

// applyDirOwnership applies Unix directory ownership if possible.
func applyDirOwnership(dst string, like os.FileInfo) error {
	likeSys, ok := like.Sys().(*syscall.Stat_t)
	if ok {
		err := os.Chown(dst, int(likeSys.Uid), int(likeSys.Gid))
		if os.IsPermission(err) {
			err = nil
		}
		return err
	}
	return nil
}
