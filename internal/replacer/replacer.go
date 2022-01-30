package replacer

import (
	"debug/gosym"
	"fmt"
	"io"
)

var (
	// ErrFunctionNotFound returned when provided function was not found.
	ErrFunctionNotFound = fmt.Errorf("function was not found")

	// ErrUnsupportedArchitecture returned if cpu architecture currently unsupported.
	ErrUnsupportedArchitecture = fmt.Errorf("unsupported cpu architecture")

	// ErrShortFunction returned if function too short for trampoline.
	ErrShortFunction = fmt.Errorf("too short function")

	// ErrLongDistance returned if functions located too far for trampoline.
	ErrLongDistance = fmt.Errorf("long distance between functions")
)

// Executable contains methods to fetch information required for patching.
type Executable interface {
	io.WriterAt

	// GOARCH returns "GOARCH" string of executable.
	GOARCH() string

	// TextAddr returns 'text' (executable code) section address.
	TextAddr() uint64

	// GoSymTabData returns reader for 'gosymtab' section.
	GoSymTabData() io.Reader

	// GoPCLnTabData returns reader for 'gopclntab' section.
	GoPCLnTabData() io.Reader

	// Offset returns function offset from beginning of executable.
	Offset(p *gosym.Func) int64
}

type trampolineGenerator interface {
	GenerateTrampoline(source, target *gosym.Func) ([]byte, error)
}

type Replacer struct {
	executable Executable
	generator  trampolineGenerator
	gosymtab   *gosym.Table
	funcIdx    map[string]gosym.Func
}

func NewReplacer(executable Executable) (*Replacer, error) {
	generator, err := trampolineFromGOARCH(executable.GOARCH())
	if err != nil {
		return nil, err
	}

	symtabData, err := io.ReadAll(executable.GoSymTabData())
	if err != nil {
		return nil, fmt.Errorf("gosymtab read failed: %w", err)
	}

	pclntabData, err := io.ReadAll(executable.GoPCLnTabData())
	if err != nil {
		return nil, fmt.Errorf("gopclntab read failed: %w", err)
	}

	gosymtab, err := gosym.NewTable(symtabData, gosym.NewLineTable(pclntabData, executable.TextAddr()))
	if err != nil {
		return nil, fmt.Errorf("gosym.Newtable failed: %w", err)
	}

	idx := make(map[string]gosym.Func)
	for _, fn := range gosymtab.Funcs {
		idx[fn.Name] = fn
	}

	return &Replacer{
		executable: executable,
		generator:  generator,
		gosymtab:   gosymtab,
		funcIdx:    idx,
	}, nil
}

// Replace puts "trampoline code" to beginning of function with sourceName that redirects to function with targetName.
// Function names here are "raw" (no mangling, etc. performed before search).
// There is no checks about "cyclic replacement" (i.e. "a"->"b" than "b"->"a") so be careful to avoid infinite loops.
func (r *Replacer) Replace(sourceName, targetName string) error {
	sourceFunc, ok := r.funcIdx[sourceName]
	if !ok {
		return fmt.Errorf("source %s: %w", sourceName, ErrFunctionNotFound)
	}

	targetFunc, ok := r.funcIdx[targetName]
	if !ok {
		return fmt.Errorf("target %s: %w", targetName, ErrFunctionNotFound)
	}

	trampoline, err := r.generator.GenerateTrampoline(&sourceFunc, &targetFunc)
	if err != nil {
		return err
	}
	if uint64(len(trampoline)) > (sourceFunc.End - sourceFunc.Entry) {
		return ErrShortFunction
	}

	_, err = r.executable.WriteAt(trampoline, r.executable.Offset(&sourceFunc))
	if err != nil {
		return fmt.Errorf("write trampoline: %w", err)
	}

	return nil
}
