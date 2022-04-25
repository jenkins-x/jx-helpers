package files

import (
	"io"
	"os"

	"github.com/AlecAivazis/survey/v2/terminal"
)

// IOFileHandles is a struct for holding CommonOptions' In, Out, and Err I/O handles, to simplify function calls.
type IOFileHandles struct {
	Err io.Writer
	In  terminal.FileReader
	Out terminal.FileWriter
}

// GetIOFileHandles lazily creates a file handles object if the input is nil
func GetIOFileHandles(h *IOFileHandles) IOFileHandles {
	if h == nil {
		h = &IOFileHandles{
			Err: os.Stderr,
			In:  os.Stdin,
			Out: os.Stdout,
		}
	}
	return *h
}
