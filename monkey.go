package monkey

import (
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/xakep666/monkey/internal/executable"
	"github.com/xakep666/monkey/internal/replacer"
)

var (
	// ErrCyclicReplacement returned if you want to perform replacements like "a"->"b" and "b"->"a".
	// Such replacements lead to infinite loop.
	ErrCyclicReplacement = fmt.Errorf("cyclic replacement detected")

	// ErrFunctionNotFound returned when provided function was not found.
	ErrFunctionNotFound = replacer.ErrFunctionNotFound

	// ErrUnsupportedArchitecture returned if cpu architecture currently unsupported.
	ErrUnsupportedArchitecture = replacer.ErrUnsupportedArchitecture

	// ErrShortFunction returned if function too short for trampoline.
	ErrShortFunction = replacer.ErrShortFunction

	// ErrLongDistance returned if functions located too far for trampoline.
	ErrLongDistance = replacer.ErrLongDistance
)

// Patcher is a registry of function replacements applied to executable
type Patcher struct {
	replacements map[string]string // original function name to new function name
	stickyErr    error
}

// NewPatcher constructs Patcher.
func NewPatcher() *Patcher {
	return &Patcher{
		replacements: map[string]string{},
	}
}

// Apply just calls callback with itself. This needed to support "flow-style" interface.
func (p *Patcher) Apply(cb func(patcher *Patcher)) *Patcher {
	cb(p)
	return p
}

// RegisterReplacement registers function replacement in patcher.
// Defined as function because it's impossible to use different type parameters in methods.
// Note that arguments must be functions despite "any" used as constraint
//	because generics doesn't allow to specify that parameter must be "any function".
func RegisterReplacement[T any](p *Patcher, original, replacement T) {
	originalValue := reflect.ValueOf(original)
	replacementValue := reflect.ValueOf(replacement)

	if originalValue.Kind() != reflect.Func || replacementValue.Kind() != reflect.Func {
		p.stickyErr = ErrFunctionNotFound
		return
	}

	originalFunc := runtime.FuncForPC(uintptr(originalValue.UnsafePointer()))
	replacementFunc := runtime.FuncForPC(uintptr(replacementValue.UnsafePointer()))

	if originalFunc == nil || replacementFunc == nil {
		p.stickyErr = ErrFunctionNotFound
		return
	}

	p.replacements[originalFunc.Name()] = replacementFunc.Name()
}

func (p *Patcher) detectCyclicReplacements() error {
	visitedAll := make(map[string]struct{})
	queue := make([]string, 0)

	for original := range p.replacements {
		if _, ok := visitedAll[original]; ok {
			continue // already checked this chain
		}

		visited := make(map[string]struct{})

		queue = append(queue[:0], original)
		for i := 0; i < len(queue); i++ {
			visitedAll[queue[i]] = struct{}{}
			visited[queue[i]] = struct{}{}

			replacement, ok := p.replacements[queue[i]]
			if !ok {
				continue // no replacement registered
			}

			if _, ok = visited[replacement]; ok {
				return ErrCyclicReplacement // back-reference detected so it's a cycle
			}

			queue = append(queue, replacement)
		}
	}

	return nil
}

func (p *Patcher) makeReplacements(rw executable.ReadWriterAt) error {
	if err := p.detectCyclicReplacements(); err != nil {
		return err
	}

	exe, err := executable.BySignature(rw)
	if err != nil {
		return err
	}

	r, err := replacer.NewReplacer(exe)
	if err != nil {
		return err
	}

	for originalName, replacementName := range p.replacements {
		if err = r.Replace(originalName, replacementName); err != nil {
			return err
		}
	}

	return nil
}

// PatchAndExec makes patches according to registered replacements and re-runs executable.
// Algorithm:
// 0) Check if we are not running patched executable, otherwise go to 1.
//	This made by checking presence (or exact value if specified) of special environment variable.
// 0.1) Remove itself if such option specified.
// 1) Copy current executable to temporary file
// 2) For each replacement: replace beginning of original function with "trampoline" to replacement function.
// 3) Run patched executable with special environment variable to avoid recursions (this is terminal condition).
// 	'execve' system call used on *nix systems, 'exec.Command' with stdin/out/err attached and 'os.Exit' after termination on others.
//  So on successful run all code after this function in original executable will become unreachable.
func (p *Patcher) PatchAndExec(opts ...PatchAndExecOption) error {
	if p.stickyErr != nil {
		return p.stickyErr
	}

	settings := patchAndExecOptions{envVarName: "XXX_REPLACED"}
	settings.applyAll(opts...)

	myPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	if os.Getenv(settings.envVarName) != settings.envVarValue {
		if settings.removePatched {
			_ = os.Remove(myPath)
		}

		return nil
	}

	tmp, err := copyToTemp(myPath)
	if err != nil {
		return fmt.Errorf("copy to temp file: %w", err)
	}

	tmpPath := tmp.Name()

	if err = p.makeReplacements(tmp); err != nil {
		return err
	}

	_ = tmp.Sync()
	_ = tmp.Close()

	envVarValue := settings.envVarValue
	if envVarValue == "" {
		envVarValue = "1"
	}

	return execWithEnv(tmpPath, append(os.Environ(), settings.envVarName+"="+envVarValue))
}

// MustPatchAndExec acts like PatchAndExec but panics on errors.
func (p *Patcher) MustPatchAndExec(opts ...PatchAndExecOption) {
	if err := p.PatchAndExec(opts...); err != nil {
		panic(err)
	}
}
