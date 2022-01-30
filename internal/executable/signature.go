package executable

import (
	"debug/buildinfo"
	"fmt"
	"github.com/xakep666/monkey/internal/replacer"
	"io"
)

// ErrUnknownExecutable returned if executable format was not recognized.
var ErrUnknownExecutable = fmt.Errorf("unknown executable format")

// BySignature attempts to recognize executable by first bytes.
func BySignature(rw ReadWriterAt) (replacer.Executable, error) {
	var signature [4]byte

	_, err := rw.ReadAt(signature[:], 0)
	if err != nil {
		return nil, fmt.Errorf("signature read: %w", err)
	}

	var ret replacer.Executable

	switch {
	case signature == [...]byte{0x7f, 'E', 'L', 'F'}: // ELF
		ret, err = NewELF(rw)
	case signature == [...]byte{0xf3, 0xed, 0xfa, 0xce}, // 32bit mach-o
		signature == [...]byte{0xf3, 0xed, 0xfa, 0xcf}: // 64bit mach-o
		ret, err = NewMachO(rw)
	case string(signature[:2]) == "MZ": // PE
		ret, err = NewPE(rw)
	// TODO: XCOFF support
	default:
		return nil, ErrUnknownExecutable
	}

	if err != nil {
		return nil, err
	}

	return ret, nil
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
