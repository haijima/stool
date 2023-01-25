package cmd

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

type IO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
	Fs  afero.Fs
}

func NewStdIO() IO {
	return IO{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
		Fs:  afero.NewOsFs(),
	}
}
