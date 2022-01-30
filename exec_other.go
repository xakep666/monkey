//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package monkey

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func execWithEnv(path string, environ []string) error {
	cmd := exec.Command(path, os.Args[1:]...)
	cmd.Env = environ
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var exitErr *exec.ExitError
	err := cmd.Run()
	switch {
	case errors.Is(err, nil):
		os.Exit(0)
		return nil
	case errors.As(err, &exitErr):
		if exitErr.ExitCode() == -1 {
			return fmt.Errorf("exec error: %w", err)
		}

		os.Exit(exitErr.ExitCode())
		return nil
	default:
		return fmt.Errorf("exec error: %w", err)
	}
}
