Monkey
========

[![Go Reference](https://pkg.go.dev/badge/github.com/xakep666/monkey.svg)](https://pkg.go.dev/github.com/xakep666/monkey)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Test](https://github.com/xakep666/monkey/actions/workflows/testing.yml/badge.svg)](https://github.com/xakep666/monkey/actions/workflows/testing.yml)

"Monkey" is a library for monkey-patching functions. 
This may be useful to get determined test result with functions that have side effect (like `time.Now()`).

# Why does this library exist?

Earlier I found library [github.com/bouk/monkey](https://github.com/bouk/monkey) with same name and functionality and sometimes used it in tests.
But this library was unstable, and currently it's archived. So I decided to create new one with different approach.

# Usage

Monkey-patching `time.Now()` in tests:
```go
package sometest

import (
	"testing"
	"time"
	
	"github.com/xakep666/monkey"
)

func init() {
	monkey.NewPatcher().
		Apply(func(patcher *monkey.Patcher) {
			monkey.RegisterReplacement(patcher, time.Now, func() time.Time {
				return time.Date(1980, 1, 2, 3, 4, 5, 6, time.UTC)
			})
		}).
		MustPatchAndExec()
}

func TestTime(t *testing.T) {
	if now := time.Now(); !now.Equal(time.Date(1980, 1, 2, 3, 4, 5, 6, time.UTC)) {
		t.Errorf("Time not patched, returned: %s", now)
	}
}
```

More examples can be found [here](example/main.go).

# How does it work

* Developer register own replacements for specified functions.
* Some checks performed i.e. for cyclic replacements.
* Current executable copied to temporary directory. Further operations will be performed with new temporary executable.
* Unconditional jump instructions inserted at the beginning of specified functions.
* New patched binary executed.

To prevent recursive self-(re)start this library adds special environment variable when it starts patched binary.

# Comparison with `github.com/bouk/monkey`

Unlike mentioned library this performs patching _before_ binary execution. This results in major advantages but
has same disadvantages.

Advantages:
* No `unsafe` imported.
* No `mprotect`-like system calls. Some systems refused to set writeable and executable flag on pages.
* Process memory (executable code) not modified in runtime.
* No data-races during patch and call processes. It follows from the previous paragraph.

Disadvantages:
* Disk activity (writing to temporary folder).
* Impossible to "unpatch" or somehow call original version of function.
* Sometimes may fail to locate address of function inside executable.

Here is some points why patch may fail:
* OS temp directory not available for writing or binaries executing.
* Unsupported architecture. This library contains binary opcodes of unconditional jump instructions for different architectures.
* Target function inlined by compiler. To avoid this use `//go:noinline` pragma or `-gcflags=-l` compiler flag.
* Attempt to patch interface method. But sometimes it may work (see example).
* Missing symbol table and/or PC-Line table needed to locate function address in executable by name. I've seen this only on Windows with under such circumstances:
  * `go test` without `-o`
  * `go run`
  * `go build` with `-ldflags=-s`
