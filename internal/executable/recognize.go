package executable

import (
	"debug/buildinfo"
	"errors"
	"fmt"
	"github.com/xakep666/monkey/internal/replacer"
	"io"
)

// ErrUnknownExecutable returned if executable format was not recognized.
var ErrUnknownExecutable = fmt.Errorf("unknown executable format")

// Recognize attempts to recognize executable by first bytes.
func Recognize(rw ReadWriterAt) (replacer.Executable, error) {
	var notGo ErrNotGo

	if ret, err := NewELF(rw); err == nil {
		return ret, nil
	} else if errors.As(err, &notGo) {
		return nil, fmt.Errorf("elf: %w", err)
	}

	if ret, err := NewMachO(rw); err == nil {
		return ret, nil
	} else if errors.As(err, &notGo) {
		return nil, fmt.Errorf("mach-o: %w", err)
	}

	if ret, err := NewPE(rw); err == nil {
		return ret, nil
	} else if errors.As(err, &notGo) {
		return nil, fmt.Errorf("pe: %w", err)
	}

	// TODO: XCOFF support

	return nil, ErrUnknownExecutable
}

func getGOARCH(r io.ReaderAt) string {
	bi, err := buildinfo.Read(r)
	if err != nil {
		return ""
	}

	for _, setting := range bi.Settings {
		if setting.Key == "GOARCH" {
			return setting.Value
		}
	}

	return ""
}
