package executable

import (
	"fmt"
	"io"
)

// ErrNotGo returned if executable is not go-program.
var ErrNotGo = fmt.Errorf("not a golang executable")

// ReadWriterAt combines io.ReaderAt and io.WriterAt.
type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}
