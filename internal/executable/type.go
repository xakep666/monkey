package executable

import (
	"io"
)

// ErrNotGo returned if executable is not go-program.
type ErrNotGo string

func (e ErrNotGo) Error() string {
	return "not a golang executable: " + string(e)
}

// ReadWriterAt combines io.ReaderAt and io.WriterAt.
type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}
