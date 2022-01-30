package monkey

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyToTemp(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("source open failed: %w", err)
	}

	defer f.Close()

	tmp, err := os.CreateTemp(os.TempDir(), "*"+filepath.Ext(path))
	if err != nil {
		return nil, fmt.Errorf("create temp file failed: %w", err)
	}

	err = os.Chmod(tmp.Name(), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("chmod failed: %w", err)
	}

	_, err = io.Copy(tmp, f)
	if err != nil {
		return nil, fmt.Errorf("copy to temp file failed: %w", err)
	}

	_, err = tmp.Seek(io.SeekStart, 0)
	if err != nil {
		return nil, fmt.Errorf("seek start failed: %w", err)
	}

	return tmp, nil
}
