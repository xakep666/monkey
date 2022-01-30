//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package monkey

import (
	"os"
	"syscall"
)

func execWithEnv(path string, environ []string) error { return syscall.Exec(path, os.Args, environ) }
